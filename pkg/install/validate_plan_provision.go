package install

import (
	"fmt"
	"os"
	"regexp"

	"github.com/apprenda/kismatic/pkg/util"
)

//Every single possible EC2 instance type.
const EC2Regexp string = `((t2.(nano|micro|small|medium|(|x|2x)large))|
							(m4.((|x|2x|4x|10x|16x)large))|
							(m3.(medium|(|x|2x)large))|
							(c5.((|x|2x|4x|9x|18x)large))|
							(c4.((|x|2x|4x|8x)large))|
							(c3.((|x|2x|4x|8x)large))|
							(x1.(16|32)xlarge)|
							(x1e.32xlarge)|
							(r4.(|x|2x|4x|8x|16x)large)|
							(r3.((|x|2x|4x|8x)large))|
							(p3.(2|8|16)xlarge)|
							(p2.(x|8x|16x)large)|
							(g3.(4|8|16)xlarge)|
							(f1.16xlarge)|
							(i3.(|x|2x|4x|8x|16x)large)|
							(d2.(|2|4|8)xlarge))`

// ValidatePlanForProvisioner runs validation against the installation plan.
// It is similar to ValidatePlan but for a smaller set of rules that are critical to provisioning infrastructure.
func ValidatePlanForProvisioner(p *Plan) (bool, []error) {
	v := newValidator()
	v.validate(&p.Provisioner)
	v.validate(&provisionerNodeList{Nodes: p.getAllNodes()})
	v.validateWithErrPrefix("Etcd nodes", &provisionerNodeGroup{NodeGroup: p.Etcd})
	v.validateWithErrPrefix("Master nodes", &provisionerMasterNodeGroup{MasterNodeGroup: p.Master})
	v.validateWithErrPrefix("Worker nodes", &provisionerNodeGroup{NodeGroup: p.Worker})
	v.validateWithErrPrefix("Ingress nodes", &provisionerNodeGroup{NodeGroup: NodeGroup(p.Ingress)})
	return v.valid()
}

type provisionerNodeGroup struct {
	NodeGroup
}

type provisionerMasterNodeGroup struct {
	MasterNodeGroup
}

type provisionerOptionalNodeGroup provisionerNodeGroup

type provisionerNode struct {
	Node
}

type provisionerNodeList struct {
	Nodes []Node
}

func (p *Provisioner) validate() (bool, []error) {
	v := newValidator()
	if p.Provider == "" {
		v.addError(fmt.Errorf("Provisioner provider cannot be empty"))
		return v.valid()
	}
	if !util.Contains(p.Provider, InfrastructureProviders()) {
		v.addError(fmt.Errorf("%q is not a valid provisioner provider. Options are %v", p.Provider, InfrastructureProviders()))
	}
	if p.Provider != "" {
		switch p.Provider {
		case "aws":
			if aws := os.Getenv("AWS_ACCESS_KEY_ID"); aws == "" {
				v.addError(fmt.Errorf("AWS_ACCESS_KEY_ID not found"))
			}
			if aws := os.Getenv("AWS_SECRET_ACCESS_KEY"); aws == "" {
				v.addError(fmt.Errorf("AWS_SECRET_ACCESS_KEY not found"))
			}
			if aws := os.Getenv("AWS_DEFAULT_REGION"); aws == "" {
				v.addError(fmt.Errorf("AWS_DEFAULT_REGION not found"))
			}
			// if p.AWSOptions.PrivateSSHKeyPath == "" {
			// 	v.addError(fmt.Errorf("SSH private key path must be set to properly use kismatic"))
			// }
			// if p.AWSOptions.PublicSSHKey == "" {
			// 	v.addError(fmt.Errorf("SSH public key must be set to properly use kismatic"))
			// }

			validEC2Type, err := regexp.MatchString(EC2Regexp, p.AWSOptions.AMI)
			if err != nil {
				v.addError(fmt.Errorf("Could not determine if %q is an EC2 instance type: %v", p.AWSOptions.AMI, err))
			}
			if !validEC2Type {
				v.addError(fmt.Errorf("%q is not a valid EC2 instance", p.AWSOptions.AMI))
			}
			//TODO add the rest of the validation for AWS
		}
	}
	return v.valid()
}

func (mng *provisionerMasterNodeGroup) validate() (bool, []error) {
	v := newValidator()

	if len(mng.Nodes) <= 0 {
		v.addError(fmt.Errorf("At least one node is required"))
	}
	if mng.ExpectedCount <= 0 {
		v.addError(fmt.Errorf("Node count must be greater than 0"))
	}
	if len(mng.Nodes) != mng.ExpectedCount && (len(mng.Nodes) > 0 && mng.ExpectedCount > 0) {
		v.addError(fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", mng.ExpectedCount, len(mng.Nodes)))
	}
	for i, n := range mng.Nodes {
		v.validateWithErrPrefix(fmt.Sprintf("Node #%d", i+1), &provisionerNode{Node: n})
	}

	if mng.LoadBalancedFQDN != "${load_balanced_fqdn}" {
		v.addError(fmt.Errorf("Load balanced FQDN is not a valid templated string, should be '${load_balanced_fqdn}'"))
	}

	if mng.LoadBalancedShortName != "${load_balanced_short_name}" {
		v.addError(fmt.Errorf("Load balanced shortname is not a valid templated string, should be '${load_balanced_short_name}'"))
	}

	return v.valid()
}

func (nl provisionerNodeList) validate() (bool, []error) {
	v := newValidator()
	v.addError(validateNoDuplicateNodeInfo(nl.Nodes)...)
	return v.valid()
}

func (ng *provisionerNodeGroup) validate() (bool, []error) {
	v := newValidator()
	if ng == nil || len(ng.Nodes) <= 0 {
		v.addError(fmt.Errorf("At least one node is required"))
	}
	if ng.ExpectedCount <= 0 {
		v.addError(fmt.Errorf("Node count must be greater than 0"))
	}
	if len(ng.Nodes) != ng.ExpectedCount && (len(ng.Nodes) > 0 && ng.ExpectedCount > 0) {
		v.addError(fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", ng.ExpectedCount, len(ng.Nodes)))
	}
	for i, n := range ng.Nodes {
		v.validateWithErrPrefix(fmt.Sprintf("Node #%d", i+1), &provisionerNode{Node: n})
	}

	return v.valid()
}

func (ong *provisionerOptionalNodeGroup) validate() (bool, []error) {
	if ong == nil {
		return true, nil
	}
	if len(ong.Nodes) == 0 && ong.ExpectedCount == 0 {
		return true, nil
	}
	if len(ong.Nodes) != ong.ExpectedCount {
		return false, []error{fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", ong.ExpectedCount, len(ong.Nodes))}
	}
	ng := provisionerNodeGroup(*ong)
	return ng.validate()
}

func (n *provisionerNode) validate() (bool, []error) {
	v := newValidator()
	hostPattern := `\${(master|etcd|worker|ingress|storage)_host_\d+}`
	pubIPPattern := `\${(master|etcd|worker|ingress|storage)_pub_ip_\d+}`
	privIPPattern := `\${(master|etcd|worker|ingress|storage)_priv_ip_\d+}`
	// Hostnames need to be templates ${(master|etcd|worker|ingress|storage)_host_#}
	template, err := regexp.MatchString(hostPattern, n.Host)
	if err != nil {
		v.addError(fmt.Errorf("Could not determine if %q is a templated value: %v", n.Host, err))
	}
	if !template {
		v.addError(fmt.Errorf("%q is not a valid IP templated string, should be '%s'", n.IP, hostPattern))
	}
	// IPs need to be templates ${(master|etcd|worker|ingress|storage)_pub_ip_#}
	template, err = regexp.MatchString(pubIPPattern, n.IP)
	if err != nil {
		v.addError(fmt.Errorf("Could not determine if %q is a templated value: %v", n.IP, err))
	}
	if !template {
		v.addError(fmt.Errorf("%q is not a valid IP templated string, should be '%s'", n.IP, pubIPPattern))
	}
	// InternalIPs need to be templates ${(master|etcd|worker|ingress|storage)_priv_ip_#}
	template, err = regexp.MatchString(privIPPattern, n.InternalIP)
	if err != nil {
		v.addError(fmt.Errorf("Could not determine if %q is a templated value: %v", n.InternalIP, err))
	}
	if !template && n.InternalIP != "" {
		v.addError(fmt.Errorf("%q is not a valid InternalIP templated string, should be '%s'", n.InternalIP, privIPPattern))
	}
	return v.valid()
}
