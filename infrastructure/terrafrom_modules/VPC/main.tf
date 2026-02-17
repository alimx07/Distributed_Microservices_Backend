  terraform {
    backend "s3" {}
  }


resource "aws_vpc" "this" {
    cidr_block = var.cidr_block
    enable_dns_hostnames = var.enable_dns_hostnames
    enable_dns_support = var.enable_dns_support
    tags = local.default_tags
}

# PUBLIC SUBNETS

resource "aws_subnet" "public_subnets" {
    count = length(local.azs)
    vpc_id = aws_vpc.this.id
    availability_zone = element(local.azs , count.index)
    cidr_block = cidrsubnet(aws_vpc.this.cidr_block , var.cidr_subnet_mask , count.index)
    map_public_ip_on_launch = var.map_public_ip_on_launch
    tags = local.default_tags
}


resource "aws_internet_gateway" "IGW" {
    vpc_id = aws_vpc.this.id
    tags = var.default_tags
}

resource "aws_route_table" "public_tables" {
    count = length(local.azs)
    vpc_id = aws_vpc.this.id
    tags = local.default_tags
}

resource "aws_route" "public_routes" {
    count = length(local.azs)
    route_table_id = element(aws_route_table.public_tables[*].id , count.index)
    destination_cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.IGW.id
}


# PRIVATE SUBNETS
resource "aws_subnet" "private_subnets" {
    count = length(local.azs)
    vpc_id = aws_vpc.this.id
    availability_zone = element(local.azs , count.index)
    cidr_block = cidrsubnet(aws_vpc.this.cidr_block , var.cidr_subnet_mask , count.index + length(local.azs))
    tags = local.default_tags
}

resource "aws_eip" "EIPs" {
    count = local.nat_gateways_num
    tags = local.default_tags
}


resource "aws_nat_gateway" "NATs" {
    count = local.nat_gateways_num
    allocation_id = local.nat_gateways_num > 1 ? element(aws_eip.EIPs[*].id , count.index) : null
    availability_mode = local.nat_gateways_num > 1 ? "zonal" : "regional"
    vpc_id = local.nat_gateways_num > 1 ?  null : aws_vpc.this.id
    subnet_id = local.nat_gateways_num > 1 ? element(aws_subnet.public_subnets[*].id , count.index) : null 
    availability_zone_address {
      allocation_ids = local.nat_gateways_num > 1 ? null : aws_eip.EIPs[*].id 
      availability_zone_id = local.nat_gateways_num >1 ? null : data.aws_availability_zones.azs.id
    }
    tags = local.default_tags
}


resource "aws_route_table" "private_tables" {
    count = length(local.azs)
    vpc_id = aws_vpc.this.id
    tags = local.default_tags
}

resource "aws_route" "private_routes" {
    count = length(local.azs)
    route_table_id = element(aws_route_table.private_tables[*].id , count.index)
    destination_cidr_block = "0.0.0.0/0"
    nat_gateway_id = element(aws_nat_gateway.NATs[*].id , count.index)
}

resource "aws_route_table_association" "public_associations" {
  count          = length(local.azs)
  subnet_id      = element(aws_subnet.public_subnets[*].id, count.index)
  route_table_id = element(aws_route_table.public_tables[*].id, count.index)
}

resource "aws_route_table_association" "private_associations" {
  count          = length(local.azs)
  subnet_id      = element(aws_subnet.private_subnets[*].id, count.index)
  route_table_id = element(aws_route_table.private_tables[*].id, count.index)
}


# INFRA SUBNETS
resource "aws_subnet" "infra_subnets" {
    count = length(local.azs)
    vpc_id = aws_vpc.this.id
    availability_zone = element(local.azs , count.index)
    cidr_block = cidrsubnet(aws_vpc.this.cidr_block , var.cidr_subnet_mask , count.index + (2*length(local.azs)))
    tags = local.default_tags
}
