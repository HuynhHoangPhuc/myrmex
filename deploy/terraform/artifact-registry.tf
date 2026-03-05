# Artifact Registry — Docker repository for all Myrmex service images
resource "google_artifact_registry_repository" "myrmex" {
  repository_id = "myrmex"
  format        = "DOCKER"
  location      = var.region
  description   = "Myrmex ERP Docker images (core, modules, frontend)"

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }

  depends_on = [google_project_service.apis]
}
