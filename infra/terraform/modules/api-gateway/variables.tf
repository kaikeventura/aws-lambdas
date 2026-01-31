variable "project" {
  type = string
}

variable "env" {
  type = string
}

variable "api_name" {
  description = "Suffix for the API Name (e.g. 'python' or 'java')"
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Invoke ARN of the target Lambda"
  type        = string
}

variable "lambda_function_name" {
  description = "Function Name of the target Lambda"
  type        = string
}
