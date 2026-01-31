module "network" {
  source = "../../modules/network"

  project = var.project
  env     = var.env

  vpc_cidr = var.vpc_cidr
}

module "dynamodb_users" {
  source     = "../../modules/dynamodb"
  table_name = "users-dev"

  tags = {
    Environment = var.env
    Project     = var.project
  }
}

data "aws_caller_identity" "current" {}
