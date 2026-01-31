output "vpc_id" {
  value = module.network.vpc_id
}

output "dynamodb_users_table_name" {
  description = "Name of the users DynamoDB table."
  value       = module.dynamodb_users.table_name
}

output "dynamodb_users_table_arn" {
  description = "ARN of the users DynamoDB table."
  value       = module.dynamodb_users.table_arn
}
