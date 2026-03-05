output "frontend_url" {
  description = "Public URL of the frontend Cloud Run service"
  value       = google_cloud_run_v2_service.frontend.uri
}

output "core_url" {
  description = "Public URL of the core API / gateway Cloud Run service"
  value       = google_cloud_run_v2_service.core.uri
}

output "database_connection_name" {
  description = "Cloud SQL connection name (used for Cloud SQL Auth Proxy)"
  value       = google_sql_database_instance.postgres.connection_name
}

output "redis_host" {
  description = "Memorystore Redis host IP (private, accessible via VPC connector)"
  value       = google_redis_instance.redis.host
}

output "artifact_registry_url" {
  description = "Full Artifact Registry Docker repository URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.myrmex.repository_id}"
}
