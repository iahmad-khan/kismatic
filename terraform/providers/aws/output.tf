# Used for rendering the template into an actual plan.

output "master_node_ips" {
  value = "${join(",",aws_instance.master.*.ipv4_address)}"
}

output "etcd_node_ips" {
  value = "${join(",",aws_instance.etcd.*.ipv4_address)}"
}

output "worker_node_ips" {
  value = "${join(",",aws_instance.worker.*.ipv4_address)}"
}

output "ingress_node_ips" {
  value = "${join(",",aws_instance.ingress.*.ipv4_address)}"
}

output "storage_node_ips" {
  value = "${join(",",aws_instance.storage.*.ipv4_address)}"
}

output "rendered_template" {
    value = "${data.template_file.kismatic_cluster.rendered}"
}