output "lambda_arn" {
  description = "The ARN of the Java Lambda function"
  value       = aws_lambda_function.java_lambda.arn
}

output "lambda_name" {
  description = "The name of the Java Lambda function"
  value       = aws_lambda_function.java_lambda.function_name
}
