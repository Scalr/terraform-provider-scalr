variable "vpc_id" {
  description = "VPC ID where the ECS cluster will be created"
  type        = string
}

variable "allow_all_ingress" {
  description = "Whether to allow all ingress traffic"
  type        = bool
  default     = false
}

variable "security_group_name" {
  type        = string
  description = "AWS Security Group name"
}

variable "cluster_name" {
  type        = string
  description = "Name of the ECS cluster"
}

variable "task_name" {
  type        = string
  description = "Name of the ECS task"
}

variable "image" {
  type        = string
  description = "ECS container image"
}

variable "scalr_url" {}
variable "scalr_agent_token" {}

variable "limit_cpu" {
  type        = number
  description = "The hard limit for the cpu unit used by the task"
}

variable "limit_memory" {
  type        = number
  description = "The hard limit for the memory used by the task"
  default     = 2048
}
