variable "project_id" {
  description = "GCP project ID — required, no default"
  type        = string
}

variable "region" {
  description = "GCP region for all resources"
  type        = string
  default     = "asia-southeast1"
}

variable "environment" {
  description = "Deployment environment label"
  type        = string
  default     = "production"
}

variable "db_instance_tier" {
  description = "Cloud SQL instance machine tier"
  type        = string
  default     = "db-f1-micro"
}

variable "redis_memory_size_gb" {
  description = "Memorystore Redis memory size in GB"
  type        = number
  default     = 1
}

variable "core_min_instances" {
  description = "Minimum Cloud Run instances for core service (keeps WS connections warm)"
  type        = number
  default     = 1
}

variable "notification_min_instances" {
  description = "Minimum Cloud Run instances for notification service"
  type        = number
  default     = 1
}

variable "docker_image_tag" {
  description = "Docker image tag to deploy across all services"
  type        = string
  default     = "latest"
}
