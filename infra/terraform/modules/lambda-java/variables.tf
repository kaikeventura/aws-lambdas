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

variable "enable_build" {
  description = "Enable or disable the build process. Set to false to skip building the artifact."
  type        = bool
  default     = true
}

variable "architecture" {
  description = "Lambda architecture (x86_64 or arm64)"
  type        = string
  default     = "x86_64"
}
