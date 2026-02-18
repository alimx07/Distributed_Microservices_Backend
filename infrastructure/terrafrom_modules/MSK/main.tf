terraform {
  backend "s3" {}
}

locals {
  prefix = "tf-dmb-${var.environment}"

  default_tags = merge(var.default_tags, {
    Name = "${local.prefix}-msk"
  })
}


module "sg" {
  source = "../SG"
  vpc_id      = var.vpc_id
  security_groups = [
    {
      name        = "${local.prefix}-msk-sg"
      description = "MSK Kafka access"
      ingress_rules = [
        {
          cidr_block  = var.private_cidr
          from_port   = 9092
          to_port     = 9098
          ip_protocol = "tcp"
          description = "Kafka brokers"
        }
      ]
      egress_rules = [
        {
          cidr_block  = "0.0.0.0/0"
          ip_protocol = "-1"
          description = "Allow all outbound"
        }
      ]
    }
  ]
}



resource "aws_msk_configuration" "this" {
  name              = "${local.prefix}-config"
  kafka_versions    = [var.kafka_version]

  server_properties = <<-EOF
    auto.create.topics.enable=true
    default.replication.factor=${var.number_of_broker_nodes > 2 ? 3 : var.number_of_broker_nodes}
    min.insync.replicas=${var.number_of_broker_nodes > 2 ? 2 : 1}
    num.partitions=3
    log.retention.hours=10
  EOF
}


resource "aws_msk_cluster" "this" {
  cluster_name           = "${local.prefix}-kafka"
  kafka_version          = var.kafka_version
  number_of_broker_nodes = var.number_of_broker_nodes

  configuration_info {
    arn      = aws_msk_configuration.this.arn
    revision = aws_msk_configuration.this.latest_revision
  }

  broker_node_group_info {
    instance_type   = var.instance_type
    client_subnets  = var.subnet_ids
    security_groups = module.sg.security_group_ids

    storage_info {
      ebs_storage_info {
        volume_size = var.volume_size
      }
    }
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS_PLAINTEXT"
      in_cluster    = false
    }
  }

  tags = local.default_tags
}


resource "aws_mskconnect_custom_plugin" "debezium" {
  name         = "${local.prefix}-debezium-plugin"
  content_type = "ZIP"

  location {
    s3 {
      bucket_arn = data.aws_s3_bucket.s3.arn
      file_key   = aws_s3_object.debezium_plugin.key
    }
  }
}


resource "aws_s3_object" "debezium_plugin" {
  bucket = data.aws_s3_bucket.s3.id
  key    = "debezium-connector-postgres.zip"
  source = var.debezium_plugin_zip_path
  etag   = filemd5(var.debezium_plugin_zip_path)
}

# IAM Role for MSK Connect
resource "aws_iam_role" "msk_connect" {
  name = "${local.prefix}-msk-connect-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "kafkaconnect.amazonaws.com"
        }
      }
    ]
  })

  tags = local.default_tags
}

resource "aws_iam_role_policy" "msk_connect" {
  name = "${local.prefix}-msk-connect-policy"
  role = aws_iam_role.msk_connect.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kafka-cluster:Connect",
          "kafka-cluster:DescribeCluster",
          "kafka-cluster:ReadData",
          "kafka-cluster:WriteData",
          "kafka-cluster:CreateTopic",
          "kafka-cluster:DescribeTopic",
          "kafka-cluster:AlterGroup",
          "kafka-cluster:DescribeGroup",
          "kafkaconnect:CreateCustomPlugin"
        ]
        Resource = "${aws_msk_cluster.this.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject"
        ]
        Resource = [
          data.aws_s3_bucket.s3.arn,
          "${data.aws_s3_bucket.s3.arn}/*"
        ]
      }
    ]
  })
}



resource "aws_mskconnect_connector" "debezium" {
  name = "${local.prefix}-debezium-postgres"

  kafkaconnect_version = "2.7.1"

  capacity {
    provisioned_capacity {
      mcu_count    = 1
      worker_count = 1
    }
  }

  # Add Secrets
  connector_configuration = {
    "connector.class"                    = "io.debezium.connector.postgresql.PostgresConnector"
    "plugin.name"                        = "pgoutput"
    "publication.name"                   = "my_pub"
    "slot.name"                          = "logical_slot"
    "database.hostname"                  = var.db_primary_host
    "database.port"                      = "5432"
    "database.user"                      = var.db_connect_user
    "database.password"                  = var.db_connect_password
    "database.dbname"                    = var.db_name
    "topic.prefix"                       = "post_service"
    "key.converter"                      = "org.apache.kafka.connect.json.JsonConverter"
    "key.converter.schemas.enable"       = "false"
    "value.converter"                    = "org.apache.kafka.connect.json.JsonConverter"
    "value.converter.schemas.enable"     = "false"
    "tombstones.on.delete"              = "false"
    "transforms"                         = "unwrap"
    "transforms.unwrap.type"             = "io.debezium.transforms.ExtractNewRecordState"
    "tasks.max"       = "1"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = aws_msk_cluster.this.bootstrap_brokers
      vpc {
        subnets         = var.subnet_ids
        security_groups = module.sg.security_group_ids
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "PLAINTEXT"
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.debezium.arn
      revision = aws_mskconnect_custom_plugin.debezium.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.msk_connect.arn
}



data "aws_s3_bucket" "s3" {
  bucket = var.bucket_name
}