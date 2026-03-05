# Service account used by all Cloud Run services
resource "google_service_account" "myrmex_run" {
  account_id   = "myrmex-run"
  display_name = "Myrmex Cloud Run Service Account"
  description  = "Identity for all Myrmex Cloud Run services"
}

# IAM bindings for the Cloud Run service account
locals {
  run_sa_roles = [
    "roles/cloudsql.client",          # Connect to Cloud SQL via private IP
    "roles/redis.editor",             # Read/write Memorystore Redis
    "roles/pubsub.publisher",         # Publish events to Pub/Sub topics
    "roles/pubsub.subscriber",        # Pull and ack Pub/Sub subscriptions
    "roles/secretmanager.secretAccessor", # Read secret versions at runtime
    "roles/artifactregistry.reader",  # Pull Docker images from Artifact Registry
  ]
}

resource "google_project_iam_member" "run_sa_roles" {
  for_each = toset(local.run_sa_roles)

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.myrmex_run.email}"
}

# Allow Cloud Run to invoke other internal Cloud Run services (gRPC module calls)
resource "google_project_iam_member" "run_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.myrmex_run.email}"
}
