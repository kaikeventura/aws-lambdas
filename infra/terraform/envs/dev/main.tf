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

module "api_gateway" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env

  python_lambda_invoke_arn    = module.lambda_python.lambda_invoke_arn
  python_lambda_function_name = module.lambda_python.lambda_function_name

  java_lambda_invoke_arn    = module.lambda_java.lambda_invoke_arn
  java_lambda_function_name = module.lambda_java.lambda_name
}

data "aws_caller_identity" "current" {}
