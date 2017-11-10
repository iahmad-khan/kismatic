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

const terraform string = "./../../bin/terraform"

type ProvisionOpts struct {
	ClusterName      string
	TemplateFileName string
}

type TerraformPlanner struct {
	File string
}

// Write the TFVars to the file system
func (tp *TerraformPlanner) Write(TFVars *install.TerraformVariables) error {
	bytez, err := json.Marshal(TFVars)
	if err != nil {
		return fmt.Errorf("error marshalling tfvars to json: %v", err)
	}
	f, err := os.Create(tp.File)
	if err != nil {
		return fmt.Errorf("error making plan file: %v", err)
	}
	defer f.Close()
	f.Write(bytez)
	return nil
}

//Provision provides a wrapper for terraform init, terraform plan, and terraform apply.
func Provision(out io.Writer, opts *ProvisionOpts, plan *install.Plan) error {

	clusterPathFromWd := fmt.Sprintf("terraform/clusters/%s/", opts.ClusterName)
	providerPathFromClusterDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)
	clusterYaml := fmt.Sprintf("%s.yaml", plan.Cluster.Name)
	tfPlanner := TerraformPlanner{File: "terraform.tfvars.json"}

	os.MkdirAll(clusterPathFromWd, 0755)
	os.Chdir(clusterPathFromWd)

	fmt.Fprintf(out, "Generating SSH keys.\n")
	ssh.NewKeyPair("sshkey.pub", "sshkey.pem")
	pubKeyData := ioutil.ReadFile("sshkey.pub")
	privKeyPath := fmt.Sprintf("%s/sshkey.pem", os.Getwd())

	tfVars := install.TerraformVariables{PrivateSSHKeyPath: privKeyPath, PublicSSHKey: pubKeyData}
	fmt.Fprintf(out, "Generating TFVars.\n")
	tfPlanner.Write(tfVars)

	tfInit := exec.Command(terraform, "init", providerPathFromClusterDir)
	if stdoutStderr, err := tfInit.CombinedOutput(); err != nil {
		return fmt.Errorf("Error initializing terraform: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner initialization successful.\n")

	tfPlan := exec.Command(terraform, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerPathFromClusterDir)

	if stdoutStderr, err := tfPlan.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform plan: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner planning successful.\n")

	fmt.Fprintf(out, "Provisioning...\n")

	tfApply := exec.Command(terraform, "apply", plan.Cluster.Name)
	if stdoutStderr, err := tfApply.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform apply: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioning successful!\n")
	fmt.Fprintf(out, "Rendering plan file...\n")

	// Render with KET in the future
	tfOutput := exec.Command(terraform, "output", "rendered_template")
	stdoutStderr, err := tfOutput.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error collecting terraform output: %s", stdoutStderr)
	}

	if err := ioutil.WriteFile(clusterYaml, stdoutStderr, 0644); err != nil {
		return fmt.Errorf("Error writing rendered file to file system")
	}
	fmt.Fprintf(out, "Plan file %s rendered.\n", clusterYaml)
	os.Chdir("../../../")
	return nil
}
