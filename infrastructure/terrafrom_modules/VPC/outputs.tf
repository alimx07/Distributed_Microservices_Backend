output "vpc_id" {
    value = aws_vpc.this.id
}

output "igw_id" {
    value = aws_internet_gateway.IGW.id
}

output "public_subnets" {
    value = aws_subnet.public_subnets[*].id
}

output "private_subnets" {
    value = aws_subnet.private_subnets[*].id
}

output "infra_subnets" {
    value = aws_subnet.infra_subnets[*].id
}