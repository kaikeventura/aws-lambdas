output "lambda_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.python_lambda.arn
}
