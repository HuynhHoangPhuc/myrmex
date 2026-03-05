# Memorystore Redis 7 — used for session caching and pub/sub fanout
resource "google_redis_instance" "redis" {
  name           = "myrmex-redis"
  tier           = "BASIC"
  memory_size_gb = var.redis_memory_size_gb
  region         = var.region

  redis_version = "REDIS_7_0"

  # Private VPC access — no public IP
  authorized_network = google_compute_network.myrmex.id

  # Persistence disabled for BASIC tier cache use-case
  persistence_config {
    persistence_mode = "DISABLED"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }

  depends_on = [
    google_project_service.apis,
    google_service_networking_connection.private_vpc_connection,
  ]
}
