# Simple Security Groups Module

resource "aws_security_group" "this" {
  count       = length(var.security_groups)
  name        = var.security_groups[count.index].name
  description = var.security_groups[count.index].description
  vpc_id      = var.vpc_id

  tags = {
    Name = var.security_groups[count.index].name
  }
}

# Ingress Rules
resource "aws_vpc_security_group_ingress_rule" "ingress" {
  for_each = {
    for idx, rule in flatten([
      for sg_idx, sg in var.security_groups : [
        for rule_idx, rule in sg.ingress_rules : {
          key         = "${sg_idx}-${rule_idx}"
          sg_idx      = sg_idx
          cidr_block  = rule.cidr_block
          from_port   = rule.from_port
          to_port     = rule.to_port
          ip_protocol = rule.ip_protocol
          description = rule.description
        }
      ]
    ]) : rule.key => rule
  }

  security_group_id = aws_security_group.this[each.value.sg_idx].id
  cidr_ipv4         = each.value.cidr_block
  from_port         = each.value.ip_protocol == "-1" ? null : each.value.from_port
  to_port           = each.value.ip_protocol == "-1" ? null : each.value.to_port
  ip_protocol       = each.value.ip_protocol
  description       = each.value.description
}

# Egress Rules
resource "aws_vpc_security_group_egress_rule" "egress" {
  for_each = {
    for idx, rule in flatten([
      for sg_idx, sg in var.security_groups : [
        for rule_idx, rule in sg.egress_rules : {
          key         = "${sg_idx}-${rule_idx}"
          sg_idx      = sg_idx
          cidr_block  = rule.cidr_block
          from_port   = rule.from_port
          to_port     = rule.to_port
          ip_protocol = rule.ip_protocol
          description = rule.description
        }
      ]
    ]) : rule.key => rule
  }

  security_group_id = aws_security_group.this[each.value.sg_idx].id
  cidr_ipv4         = each.value.cidr_block
  from_port         = each.value.ip_protocol == "-1" ? null : each.value.from_port
  to_port           = each.value.ip_protocol == "-1" ? null : each.value.to_port
  ip_protocol       = each.value.ip_protocol
  description       = each.value.description
  tags = var.default_tags
}