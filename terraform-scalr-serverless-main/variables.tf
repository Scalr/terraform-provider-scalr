variable "api_gateway_name" {
  type    = string
  default = "scalr-agent-pool-api"
}

variable "api_gateway_environment" {
  type    = string
  default = "prod"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "allow_all_ingress" {
  description = "Whether to allow all ingress traffic (overrides IP whitelist)"
  type        = bool
  default     = false
}

variable "ecs_security_group_name" {
  type        = string
  description = "AWS Security Group name"
  default     = "scalr-agent-ecs-tasks"
}

variable "ecs_cluster_name" {
  type        = string
  description = "Name of the ECS cluster"
  default     = "scalr-agent-pool-cluster"
}

variable "ecs_task_name" {
  type        = string
  description = "Name of the ECS task"
  default     = "scalr-agent-run"
}

variable "ecs_limit_cpu" {
  type        = number
  description = "The hard limit for the cpu unit used by the task"
  default     = 1024
}

variable "ecs_limit_memory" {
  type        = number
  description = "The hard limit for the memory used by the task"
  default     = 2048
}

variable "ecs_image" {
  type        = string
  description = "ECS container image"
  default     = "scalr/agent:latest"
}


variable "lambda_function_name" {
  type        = string
  description = "The name of the role to execute the function"
  default     = "scalr-agent"
}

variable "lambda_handler" {
  type        = string
  description = "Lambda handler"
  default     = "lambda_function.lambda_handler"
}

variable "lambda_runtime" {
  type        = string
  description = "Lambda handler runtime"
  default     = "python3.11"
}

variable "lambda_timeout" {
  type        = number
  description = "Lambda timeout"
  default     = 30
}

variable "lambda_memory_size" {
  type        = number
  description = "Lambda memory size"
  default     = 128
}

variable "vpc_name" {
  type = string
  description = "VPC Network name"
  default = "scalr-agent"
}

variable "scalr_hostname" {
  type = string
  description = "host name of Scalr instance"
}

variable "scalr_token" {
  type = string
  description = "Scalr token"
}

