variable "subnet_ids" {
  description = "List of subnet IDs for the Fargate task"
  type = list(string)
}

variable "cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
}

variable "task_definition_arn" {
  description = "ARN of the ECS task definition"
  type        = string
}

variable "security_group_id" {
  description = "ID of the security group for the Fargate task"
  type        = string
}

variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "source_file" {
  type        = string
  description = "Source file to of the function"
}

variable "handler" {
  description = "Handler for the Lambda function"
  type        = string
}

variable "runtime" {
  description = "Runtime for the Lambda function"
  type        = string
}

variable "timeout" {
  description = "Timeout for the Lambda function in seconds"
  type        = number
}

variable "memory_size" {
  description = "Memory size for the Lambda function in MB"
  type        = number
}