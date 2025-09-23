output "function_name" {
  description = "The name of the Lambda function"
  value       = aws_lambda_function.scalr_webhook.function_name
}

output "function_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.scalr_webhook.arn
}

output "invoke_arn" {
  description = "The ARN to invoke the Lambda function"
  value       = aws_lambda_function.scalr_webhook.invoke_arn
}
