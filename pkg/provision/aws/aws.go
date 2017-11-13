package provision

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/ssh"
)

func (aws AWS) getCommandEnvironment() []string {
	key := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", aws.AccessKey)
	secret := fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", aws.SecretKey)
	region := fmt.Sprintf("AWS_DEFAULT_REGION=%s", aws.Region)
	return []string{key, secret, region}
}

// Write the TFVars to the file system
func (aws *AWS) Write(file string) error {
	bytez, err := json.Marshal(aws)
	if err != nil {
		return fmt.Errorf("error marshalling tfvars to json: %v", err)
	}
	err := ioutil.WriteFile(file, bytez, 0700)
	if err != nil {
		return fmt.Errorf("error making tfvars file: %v", err)
	}
	return nil
}

//Provision provides a wrapper for terraform init, terraform plan, and terraform apply.
func (aws *AWS) Provision(out io.Writer, plan *install.Plan) error {

	// Create directory for keeping cluster state
 	clusterStateDir := fmt.Sprintf("terraform/clusters/%s/", plan.Cluster.Name)
 	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
 		return nil, fmt.Errorf("error creating directory to keep cluster state: %v", err)
 	}
 	if err := os.Chdir(clusterStateDir); err != nil {
 		return nil, fmt.Errorf("error switching dir to %s: %v", clusterStateDir, err)
 	}
	defer os.Chdir("../../../")
	 
	//Read tfvars from plan
	aws.Region = plan.Provisioner.AWSOptions.Region
	aws.AccessKey = os.Getenv

	//Write tfvars out
	file := "terraform.tfvars.json"
	aws.Write(file)

	clusterPathFromWd := fmt.Sprintf("terraform/clusters/%s/", plan.Cluster.Name)
	providerPathFromClusterDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)
	clusterYaml := fmt.Sprintf("%s.yaml", plan.Cluster.Name)

	os.MkdirAll(clusterPathFromWd, 0700)
	os.Chdir(clusterPathFromWd)

	fmt.Fprintf(out, "Generating SSH keys.\n")
	ssh.NewKeyPair("sshkey.pub", "sshkey.pem")
	pubKeyData, err := ioutil.ReadFile("sshkey.pub")
	if err != nil {

	}
	wd, err := os.Getwd()
	if err != nil {

	}
	privKeyPath := fmt.Sprintf("%s/sshkey.pem", wd)

	tfVars := install.AWSTerraformVariables{PrivateSSHKeyPath: privKeyPath, PublicSSHKey: pubKeyData}
	fmt.Fprintf(out, "Generating TFVars.\n")
	tfPlanner.Write(tfVars)

	tfInit := exec.Command(terraformBinaryPath, "init", providerPathFromClusterDir)
	if stdoutStderr, err := tfInit.CombinedOutput(); err != nil {
		return fmt.Errorf("Error initializing terraform: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner initialization successful.\n")

	tfPlan := exec.Command(terraformBinaryPath, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerPathFromClusterDir)

	if stdoutStderr, err := tfPlan.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform plan: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner planning successful.\n")

	fmt.Fprintf(out, "Provisioning...\n")
	tfApply := exec.Command(terraformBinaryPath, "apply", plan.Cluster.Name)
	if stdoutStderr, err := tfApply.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform apply: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioning successful!\n")

	fmt.Fprintf(out, "Rendering plan file...\n")
	err := plan.updatePlanNodes(aws.MasterCount, aws.WorkerCount, aws.EtcdCount, aws.StorageCount, aws.IngressCount)

	return nil
}


//This will probably have to accept update node groups
func (plan *install.Plan) updatePlanNodes(masters, workers, etcds, storages, ingresses int) error {
	masterDiff := plan.Master.ExpectedCount - masters
	if masterDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	etcdDiff := plan.Etcd.ExpectedCount - etcds
	if etcdDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	workerDiff := plan.Worker.ExpectedCount - workers
	if workerDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	storageDiff := plan.Storage.ExpectedCount - storages
	if storageDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	ingressDiff := plan.Ingress.ExpectedCount - ingresses
	if ingressDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	plan.Master.updateNodeIPs()
	plan.Etcd.updateNodeIPs()
	plan.Worker.updateNodeIPs()
	plan.Storage.updateNodeIPs()
	plan.Ingress.updateNodeIPs()

	if err := ioutil.WriteFile(clusterYaml, stdoutStderr, 0644); err != nil {
		return fmt.Errorf("Error writing rendered file to file system")
	}
	fmt.Fprintf(out, "Plan file %s rendered.\n", clusterYaml)
}

func (nodes *install.MasterNodeGroup) updateNodeIPs() error {

}

func (nodes *install.NodeGroup) updateNodeIPs(role string)  {
	aws:=&awsOutput{}
	tfOutPubIPs:=fmt.Sprintf("%s_pub_ips", role)
	tfOutPrivIPs:=fmt.Sprintf("%s_priv_ips", role)
	tfOutHosts:=fmt.Sprintf("%s_hosts",role)
	aws.tfOuput(role)
	for i:=0; i<nodes.ExpectedCount; i++ {
		nodes.Nodes.IP
	}
}


func (aws *awsOutput) tfOuput(variable string) error {
	tfCmdOutput := exec.Command(terraformBinaryPath, "output","-json" fmt.Sprintf("%s", variable))
	stdoutStderr, err := tfCmdOutput.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error collecting terraform output: %s", stdoutStderr)
	}
	json.Unmarshal(stdoutStderr,aws)
	return nil
}
