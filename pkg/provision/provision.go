package provision

import (
	"log"

	"github.com/apprenda/kismatic/pkg/install"
)

const terraformBinaryPath = "../../bin/terraform"

// Terraform provisioner
type Terraform struct {
	BinaryPath string
	Logger     *log.Logger
}

// Provisioner is responsible for creating and destroying infrastructure for
// a given cluster.
type Provisioner interface {
	Provision(install.Plan) (*install.Plan, error)
	Destroy(string) error
}
