variable "project" {
  description = "Nome do projeto"
  type        = string
  default     = "aws-lambdas"
}

variable "env" {
  description = "Ambiente (dev, prod, etc)"
  type        = string
  default     = "dev"
}

variable "region" {
  description = "Região AWS"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "CIDR da VPC"
  type        = string
  default     = "10.10.0.0/16"
}
