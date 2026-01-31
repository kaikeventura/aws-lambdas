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

module "lambda_python" {
  source = "../../modules/lambda-python"

  table_name = module.dynamodb_users.table_name
  table_arn  = module.dynamodb_users.table_arn
  jwt_secret = "your-super-secret-jwt-for-dev"
}

data "aws_caller_identity" "current" {}
