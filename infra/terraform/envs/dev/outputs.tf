output "dynamodb_users_table_name" {
  description = "Name of the users DynamoDB table."
  value       = module.dynamodb_users.table_name
}

output "dynamodb_users_table_arn" {
  description = "ARN of the users DynamoDB table."
  value       = module.dynamodb_users.table_arn
}

output "lambda_python_arn" {
  description = "ARN of the Python Lambda function."
  value       = module.lambda_python.lambda_arn
}

output "lambda_java_arn" {
  description = "ARN of the Java Lambda function."
  value       = module.lambda_java.lambda_arn
}

output "api_gateway_url" {
  description = "Base URL for the API Gateway"
  value       = module.api_gateway.base_url
}

output "api_key" {
  description = "API Key for accessing the API"
  value       = module.api_gateway.api_key
  sensitive   = true
}
