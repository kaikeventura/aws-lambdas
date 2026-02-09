variable "table_name" {
  description = "DynamoDB table name"
  type        = string
}

variable "table_arn" {
  description = "DynamoDB table ARN"
  type        = string
}

variable "jwt_secret" {
  description = "JWT Secret"
  type        = string
}

variable "enable_build" {
  description = "Enable build step"
  type        = bool
  default     = true
}
