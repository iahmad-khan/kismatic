package provision

// AWS provider for creating and destroying infrastructure on AWS
type AWS struct {
	Region            string    `json:"region"`
	AccessKey         string    `json:"access_key"`
	SecretKey         string    `json:"secret_key"`
	PrivateSSHKeyPath string    `json:"private_ssh_key_path"`
	PublicSSHKey      []byte    `json:"public_ssh_key"`
	ClusterName       string    `json:"cluster_name"`
	AMI               string    `json:"ami"`
	EC2InstanceType   string    `json:"instance_size"`
	MasterCount       int       `json:"master_count"`
	EtcdCount         int       `json:"etcd_count"`
	WorkerCount       int       `json:"worker_count"`
	IngressCount      int       `json:"ingress_count"`
	StorageCount      int       `json:"storage_count"`
	Terraform         Terraform `json:"-"`
}

type awsOutput struct {
	sensitive  bool     `json"sensitive"`
	outputType string   `json:"type"`
	value      []string `json:"value"`
}
