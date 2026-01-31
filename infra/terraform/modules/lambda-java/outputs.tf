output "lambda_arn" {
  description = "The ARN of the Java Lambda function"
  value       = aws_lambda_function.java_lambda.arn
}

output "lambda_invoke_arn" {
  description = "The Invoke ARN of the Java Lambda function (required for API Gateway)"
  value       = aws_lambda_function.java_lambda.invoke_arn
}

output "lambda_name" {
  description = "The name of the Java Lambda function"
  value       = aws_lambda_function.java_lambda.function_name
}
