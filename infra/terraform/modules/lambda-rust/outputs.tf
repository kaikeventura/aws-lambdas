output "lambda_arn" {
  value = aws_lambda_function.rust_lambda.arn
}

output "invoke_arn" {
  value = aws_lambda_function.rust_lambda.invoke_arn
}

output "function_name" {
  value = aws_lambda_function.rust_lambda.function_name
}
