output "lambda_arn" {
  value = aws_lambda_function.dotnet_lambda.arn
}

output "invoke_arn" {
  value = aws_lambda_function.dotnet_lambda.invoke_arn
}

output "function_name" {
  value = aws_lambda_function.dotnet_lambda.function_name
}
