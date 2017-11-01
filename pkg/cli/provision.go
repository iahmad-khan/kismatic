package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const terraform string = "./../../bin/terraform"

type provisionOpts struct {
	planFileTemplateName string
	tfClusterName        string
}

type destroyOpts struct {
	tfClusterName string
}

// NewCmdProvision creates a new provision command
func NewCmdProvision(in io.Reader, out io.Writer) *cobra.Command {
	opts := &provisionOpts{}

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "provision your Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Provision(in, out, opts)
		},
	}
	//TODO: add a cluster flag and plan file template flag
	return cmd
}

// NewCmdDestroy creates a new destroy command
func NewCmdDestroy(in io.Reader, out io.Writer) *cobra.Command {
	opts := &destroyOpts{}

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy your provisioned cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Destroy(in, out, opts)
		},
	}
	//TODO: add a cluster flag
	return cmd
}

//Provision provides a wrapper for terraform init, terraform plan, and terraform apply.
func Provision(in io.Reader, out io.Writer, opts *provisionOpts) error {
	// fp := install.FilePlanner{File: opts.planFileTemplateName}
	// plan, err := fp.Read()
	// if err != nil {
	// 	return fmt.Errorf("Plan file does not exist.")
	// }
	os.Chdir("terraform/clusters/dev/")
	tfInit := exec.Command(terraform, "init", "../../providers/aws/")
	// TODO: point it at a cluster folder and symlink it to the provider
	// TODO: use the plan file provider when you get kismatic plan upgraded
	if stdoutStderr, err := tfInit.CombinedOutput(); err != nil {
		return fmt.Errorf("Error initializing terraform: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner initialization successful.\n")

	tfPlan := exec.Command(terraform, "plan", "-out=kismatic-cluster", "../../providers/aws/")
	// TODO: make -out=plan.Name
	// TODO: make target=cluster and symlink it to the provider
	if stdoutStderr, err := tfPlan.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform plan: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner planning successful.\n")

	fmt.Fprintf(out, "Provisioning...\n")

	tfApply := exec.Command(terraform, "apply", "kismatic-cluster")
	if stdoutStderr, err := tfApply.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform apply: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioning successful!\n")
	fmt.Fprintf(out, "Rendering plan file...\n")
	tfOutput := exec.Command(terraform, "output", "rendered_template")
	stdoutStderr, err := tfOutput.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error collecting terraform output: %s", stdoutStderr)
	}

	if err := ioutil.WriteFile("kismatic-cluster.yaml", stdoutStderr, 0644); err != nil {
		return fmt.Errorf("Error writing rendered file to file system")
	}
	fmt.Fprintf(out, "Plan file %s rendered.\n", "kismatic-cluster.yaml")
	os.Chdir("../../../")
	// TODO: make sure this renders appropriately
	return nil
}

//Destroy destroys a provisioned cluster (using --force by default)
func Destroy(in io.Reader, out io.Writer, opts *destroyOpts) error {
	tfDestroy := exec.Command(terraform, "destroy", "-force")
	if stdoutStderr, err := tfDestroy.CombinedOutput(); err != nil {
		return fmt.Errorf("Error attempting to destroy: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Cluster destruction successful.\n")
	return nil
}

func getTfState(clusterName string) string {
	return fmt.Sprintf("-state=terraform/clusters/%s/terraform.tfstate", clusterName)
}

func getTfPlan(clusterName string) string {
	return fmt.Sprintf("-state=terraform/clusters/%s/kismatic-cluster", clusterName)
}

func generateSSHKey() (string, error) {
	return "", nil
}
