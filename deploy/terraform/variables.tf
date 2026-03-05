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

variable "module_min_instances" {
  description = "Minimum Cloud Run instances for module services (hr, subject, timetable, student, analytics)"
  type        = number
  default     = 1
}

variable "frontend_min_instances" {
  description = "Minimum Cloud Run instances for frontend service"
  type        = number
  default     = 1
}

variable "alert_email" {
  description = "Email address for monitoring alert notifications"
  type        = string
  default     = ""
}

variable "alert_slack_webhook_url" {
  description = "Slack webhook URL for monitoring alert notifications (optional)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "docker_image_tag" {
  description = "Docker image tag to deploy across all services"
  type        = string
  default     = "latest"
}
