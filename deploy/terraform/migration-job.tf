# Cloud Run Job for database migrations.
# Runs goose up for all 7 service schemas before each deployment.
# Executed by GitHub Actions CI/CD via: gcloud run jobs execute myrmex-migrate --wait

resource "google_cloud_run_v2_job" "migrate" {
  name     = "myrmex-migrate"
  location = var.region

  template {
    template {
      service_account = google_service_account.myrmex_run.email

      vpc_access {
        connector = google_vpc_access_connector.myrmex.id
        egress    = "ALL_TRAFFIC"
      }

      containers {
        # Image updated by CI/CD on each deploy — initial placeholder
        image = "${var.region}-docker.pkg.dev/${var.project_id}/myrmex/migrate:latest"

        env {
          name = "DATABASE_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.secrets["DATABASE_URL"].secret_id
              version = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }

      # Migrations must complete within 10 minutes
      timeout = "600s"

      # Run only 1 instance — migrations are not parallel-safe
      max_retries = 1
    }
  }

  lifecycle {
    ignore_changes = [
      # CI/CD updates the image on every deploy; Terraform manages config only
      template[0].template[0].containers[0].image,
    ]
  }
}
