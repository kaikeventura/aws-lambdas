variable "table_name" {
  description = "The name of the DynamoDB table"
  type        = string
}

variable "table_arn" {
  description = "The ARN of the DynamoDB table"
  type        = string
}

variable "jwt_secret" {
  description = "The JWT secret key"
  type        = string
}
