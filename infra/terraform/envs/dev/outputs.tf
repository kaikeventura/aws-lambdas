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

output "lambda_go_arn" {
  description = "ARN of the Go Lambda function."
  value       = module.lambda_go.lambda_arn
}

output "lambda_rust_arn" {
  description = "ARN of the Rust Lambda function."
  value       = module.lambda_rust.lambda_arn
}

# --- Python API Outputs ---
output "python_api_url" {
  description = "URL for Python API"
  value       = module.api_gateway_python.base_url
}

output "python_api_key" {
  description = "API Key for Python API"
  value       = module.api_gateway_python.api_key
  sensitive   = false
}

# --- Java API Outputs ---
output "java_api_url" {
  description = "URL for Java API"
  value       = module.api_gateway_java.base_url
}

output "java_api_key" {
  description = "API Key for Java API"
  value       = module.api_gateway_java.api_key
  sensitive   = false
}

# --- Go API Outputs ---
output "go_api_url" {
  description = "URL for Go API"
  value       = module.api_gateway_go.base_url
}

output "go_api_key" {
  description = "API Key for Go API"
  value       = module.api_gateway_go.api_key
  sensitive   = false
}

# --- Rust API Outputs ---
output "rust_api_url" {
  description = "URL for Rust API"
  value       = module.api_gateway_rust.base_url
}

output "rust_api_key" {
  description = "API Key for Rust API"
  value       = module.api_gateway_rust.api_key
  sensitive   = false
}
