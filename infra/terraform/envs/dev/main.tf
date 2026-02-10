module "dynamodb_users" {
  source     = "../../modules/dynamodb"
  table_name = "users"

  tags = {
    Environment = var.env
    Project     = var.project
  }
}

variable "jwt_secret_value" {
  type      = string
  default   = "a-super-long-and-secure-secret-for-dev-environment-that-is-safe"
  sensitive = true
}

module "lambda_python" {
  source = "../../modules/lambda-python"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = var.jwt_secret_value
  enable_build = var.enable_builds
}

module "lambda_java" {
  source = "../../modules/lambda-java"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = var.jwt_secret_value
  enable_build = var.enable_builds
}

module "lambda_go" {
  source = "../../modules/lambda-go"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = var.jwt_secret_value
  enable_build = var.enable_builds
}

module "lambda_rust" {
  source = "../../modules/lambda-rust"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = var.jwt_secret_value
  enable_build = var.enable_builds
}

module "lambda_nodejs" {
  source = "../../modules/lambda-nodejs"

  table_name   = module.dynamodb_users.table_name
  table_arn    = module.dynamodb_users.table_arn
  jwt_secret   = var.jwt_secret_value
  enable_build = var.enable_builds
}

module "api_gateway_python" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env
  api_name = "python"

  lambda_invoke_arn    = module.lambda_python.lambda_invoke_arn
  lambda_function_name = module.lambda_python.lambda_function_name
}

module "api_gateway_java" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env
  api_name = "java"

  lambda_invoke_arn    = module.lambda_java.lambda_invoke_arn
  lambda_function_name = module.lambda_java.lambda_name
}

module "api_gateway_go" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env
  api_name = "go"

  lambda_invoke_arn    = module.lambda_go.invoke_arn
  lambda_function_name = module.lambda_go.function_name
}

module "api_gateway_rust" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env
  api_name = "rust"

  lambda_invoke_arn    = module.lambda_rust.invoke_arn
  lambda_function_name = module.lambda_rust.function_name
}

module "api_gateway_nodejs" {
  source = "../../modules/api-gateway"

  project = var.project
  env     = var.env
  api_name = "nodejs"

  lambda_invoke_arn    = module.lambda_nodejs.invoke_arn
  lambda_function_name = module.lambda_nodejs.function_name
}

data "aws_caller_identity" "current" {}
