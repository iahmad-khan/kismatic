provider "aws" {
  /*
  $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY_ID
  $ export AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_ACCESS_KEY
  $ export AWS_DEFAULT_REGION=us-east-1
  */
  region = "${var.region}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["${var.ubuntu-os}"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_key_pair" "kismatic" {
  key_name   = "kismatic-key"
  public_key = "${var.ssh_key}"
}

resource "aws_instance" "master" {
  count         = "${var.master_count}"
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_size}"
  tags {
    Name = "kismatic - master"
  }
}

resource "aws_instance" "etcd" {
  count         = "${var.etcd_count}"
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_size}"
  tags {
    Name = "kismatic - etcd"
  }
}

resource "aws_instance" "worker" {
  count         = "${var.worker_count}"
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_size}"
  tags {
    Name = "kismatic - worker"
  }
}

resource "aws_instance" "ingress" {
  count         = "${var.ingress_count}"
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_size}"
  tags {
    Name = "kismatic - ingress"
  }
}

resource "aws_instance" "storage" {
  count         = "${var.storage_count}"
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_size}"
  tags {
    Name = "kismatic - storage"
  }
}


data "template_file" "kismatic_cluster" {
  template = "${file("${path.module}/../../clusters/dev/${var.cluster_name}.yaml.tpl")}"
  vars {
    etcd_ip = "${aws_instance.etcd.0.public_ip}"
    master_ip = "${aws_instance.master.0.public_ip}"
    worker_ip = "${aws_instance.worker.0.public_ip}"
    ingress_ip = "${aws_instance.ingress.0.public_ip}"
    ssh_key = "${vars.ssh_key}"
  }
}


