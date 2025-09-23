variable "environment" {
  type = string
  description = "API Gateway Environment name"
}

variable "lambda_invoke_arn" {
  description = "The ARN of the Lambda function to invoke"
  type        = string
}

variable "lambda_function_name" {
  description = "The name of the Lambda function to invoke"
  type        = string
}

variable "name" {
  type = string
  description = "Name of the API gateway"
}

variable "additional_allowed_ips" {
  description = "Additional IP addresses to allow (in CIDR notation)"
  type        = list(string)
  default     = []
}

variable "allow_all_ingress" {
  description = "Whether to allow all ingress traffic (overrides IP whitelist)"
  type        = bool
  default     = false
}

