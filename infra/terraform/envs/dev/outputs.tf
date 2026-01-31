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

# --- Python API Outputs ---
output "python_api_url" {
  description = "URL for Python API"
  value       = module.api_gateway_python.base_url
}

output "python_api_key" {
  description = "API Key for Python API"
  value       = module.api_gateway_python.api_key
  sensitive   = true
}

# --- Java API Outputs ---
output "java_api_url" {
  description = "URL for Java API"
  value       = module.api_gateway_java.base_url
}

output "java_api_key" {
  description = "API Key for Java API"
  value       = module.api_gateway_java.api_key
  sensitive   = true
}
