module "dynamodb_users" {
  source     = "../../modules/dynamodb"
  table_name = "users"

  tags = {
    Environment = var.env
    Project     = var.project
  }
}

module "lambda_python" {
  source = "../../modules/lambda-python"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = "your-super-secret-jwt-for-dev"
  enable_build = var.enable_builds
}

module "lambda_java" {
  source = "../../modules/lambda-java"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = "your-super-secret-jwt-for-dev"
  enable_build = var.enable_builds
}

data "aws_caller_identity" "current" {}
