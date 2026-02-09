output "lambda_arn" {
  value = aws_lambda_function.go_lambda.arn
}

output "invoke_arn" {
  value = aws_lambda_function.go_lambda.invoke_arn
}

output "function_name" {
  value = aws_lambda_function.go_lambda.function_name
}
