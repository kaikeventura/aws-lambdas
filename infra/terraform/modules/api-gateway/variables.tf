variable "project" {
  type = string
}

variable "env" {
  type = string
}

variable "python_lambda_invoke_arn" {
  description = "Invoke ARN for Python Lambda"
  type        = string
}

variable "python_lambda_function_name" {
  description = "Function Name for Python Lambda"
  type        = string
}

variable "java_lambda_invoke_arn" {
  description = "Invoke ARN for Java Lambda"
  type        = string
}

variable "java_lambda_function_name" {
  description = "Function Name for Java Lambda"
  type        = string
}
