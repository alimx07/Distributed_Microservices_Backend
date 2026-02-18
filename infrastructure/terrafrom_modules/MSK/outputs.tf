output "bootstrap_brokers" {
  value       = aws_msk_cluster.this.bootstrap_brokers
}

output "cluster_arn" {
  value = aws_msk_cluster.this.arn
}

output "zookeeper_connect_string" {
  value = aws_msk_cluster.this.zookeeper_connect_string
}

# output "security_group_ids" {
#   value = module.sg.security_group_ids
# }
