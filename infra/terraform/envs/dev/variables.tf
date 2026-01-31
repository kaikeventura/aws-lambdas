variable "project" {
  description = "Project name"
  type        = string
  default     = "aws-lambdas"
}

variable "env" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "enable_builds" {
  description = "Enable builds for Lambda functions"
  type        = bool
  default     = true
}
