package cli

import (
	"io"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/provision"

	"github.com/spf13/cobra"
)

// NewCmdProvision creates a new provision command
func NewCmdProvision(in io.Reader, out io.Writer) *cobra.Command {
	opts := &provision.ProvisionOpts{}
	plan := &install.Plan{}
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "provision your Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return provision.Provision(out, opts, plan)
		},
	}

	// Flags
	cmd.Flags().StringVarP(&opts.PlanFileTemplateName, "template-file", "f", "kismatic-cluster.yaml.tpl", "name of the template file within the cluster (must also specify a cluster name if used)")
	cmd.Flags().StringVarP(&opts.ClusterName, "cluster-name", "n", "kismatic-cluster", "name of the kismatic cluster to stand up. cluster names must be unique, or else provisioning will simply modified the cluster that already exists.)")

	return cmd
}

// NewCmdDestroy creates a new destroy command
func NewCmdDestroy(in io.Reader, out io.Writer) *cobra.Command {
	opts := &provision.DestroyOpts{}

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy your provisioned cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return provision.Destroy(out, opts)
		},
	}
	//Flags
	cmd.Flags().StringVarP(&opts.ClusterName, "cluster-name", "n", "kismatic-cluster", "name of the kismatic cluster to destroy.)")

	return cmd
}
