// Just static Terraform file

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket       = "amx-bucket-724"
    key          = "EC2_atlantis/terraform.tfstate"
    region       = "eu-west-1"
    use_lockfile = true
  }
}

provider "aws" {
  region = "eu-west-1"
}

data "aws_vpc" "default" {
  id = var.vpc_id
}


data "aws_iam_policy_document" "atlantis_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "atlantis" {
  name               = "atlantis-instance-role"
  assume_role_policy = data.aws_iam_policy_document.atlantis_assume.json
}


resource "aws_iam_role_policy" "atlantis_permissions" {
  # name = aws_iam_role.atlantis.name
  role = aws_iam_role.atlantis.id

  # Terrafrom need some perm. to do the job
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["sts:AssumeRole"]
        Resource = "arn:aws:iam::408502715955:role/TerraformRole_v1" # This already configured
      }
    ]
  })
}


resource "aws_iam_role_policy" "s3_premissions" {
  # name = aws_iam_role.atlantis.name
  role = aws_iam_role.atlantis.id

  policy = jsonencode({
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::amx-bucket-724",
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject","s3:DeleteObject"],
      "Resource": [
        "arn:aws:s3:::amx-bucket-724/*"
      ]
    }
  ]
})
}




# HI SSM
resource "aws_iam_role_policy_attachment" "ssm" {
  role = aws_iam_role.atlantis.id
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}


resource "aws_iam_instance_profile" "atlantis" {
  name = "atlantis-instance-profile"
  role = aws_iam_role.atlantis.name

  depends_on = [ aws_iam_role_policy_attachment.ssm  , aws_iam_role_policy.atlantis_permissions , aws_iam_role_policy.s3_premissions]
}


resource "aws_security_group" "atlantis" {
  name        = "atlantis-sg"
  description = "Atlantis server"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "Atlantis webhook port"
    from_port   = 4141
    to_port     = 4141
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "atlantis-sg" }
}


resource "aws_internet_gateway" "this" {
  vpc_id = data.aws_vpc.default.id
  tags   = { Name = "atlantis-igw" }
}

resource "aws_route_table" "this" {
  vpc_id = data.aws_vpc.default.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = { Name = "atlantis-rt" }
}

resource "aws_subnet" "this" {
  vpc_id                  = data.aws_vpc.default.id
  cidr_block              = "10.0.0.0/25"
  availability_zone       = "eu-west-1a"
  map_public_ip_on_launch = true
}

resource "aws_route_table_association" "this" {
  subnet_id      = aws_subnet.this.id
  route_table_id = aws_route_table.this.id
}

resource "aws_instance" "atlantis" {
  ami                    = "ami-00c88d6feba889de5" # Ubuntu 22.04
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.this.id
  iam_instance_profile   = aws_iam_instance_profile.atlantis.name
  vpc_security_group_ids = [aws_security_group.atlantis.id]

  user_data = templatefile("${path.module}/user.sh.tpl", {
    tf_version        = var.terraform_version
    tg_version        = var.terragrunt_version
    tgac_version      = var.terragrunt_atlantis_config_version
    atlantis_version  = var.atlantis_version
    gh_user           = var.gh_user
    gh_token          = var.gh_token
    gh_webhook_secret = var.gh_webhook_secret
    repo_allowlist    = "github.com/alimx07/Distributed_Microservices_Backend"
    repos_yaml_b64    = base64encode(file("${path.module}/repos.yaml"))
  })

  tags = { Name = "Atlantis" }
}


resource "aws_eip" "atlantis" {
  instance = aws_instance.atlantis.id
  domain   = "vpc"
  tags     = { Name = "atlantis-eip" }
}


output "atlantis_url" {
  description = "Github webhook url"
  value       = "http://${aws_eip.atlantis.public_ip}:4141"
}

# output "ssh_address" {
#   value = aws_eip.atlantis.public_ip
# }
