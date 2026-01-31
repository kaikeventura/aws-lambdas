output "base_url" {
  description = "Base URL for API Gateway"
  value       = aws_api_gateway_stage.dev.invoke_url
}

output "api_key" {
  description = "API Key value"
  value       = aws_api_gateway_api_key.mykey.value
  sensitive   = true
}
