# CI/CD service account and Workload Identity Federation for GitHub Actions keyless auth.
# This allows GitHub Actions to authenticate to GCP without storing JSON key files.

# Service account used by GitHub Actions CI/CD pipeline
resource "google_service_account" "cicd" {
  account_id   = "myrmex-cicd"
  display_name = "Myrmex CI/CD (GitHub Actions)"
  description  = "Used by GitHub Actions to build/push images and deploy to Cloud Run"
}

# Roles needed by the CI/CD pipeline
locals {
  cicd_roles = [
    "roles/run.admin",                    # deploy Cloud Run services + jobs
    "roles/artifactregistry.writer",      # push Docker images
    "roles/secretmanager.secretAccessor", # read secrets (for migration job setup)
    "roles/iam.serviceAccountUser",       # act as myrmex-run service account
    "roles/cloudbuild.builds.editor",     # submit Cloud Build jobs (optional)
  ]
}

resource "google_project_iam_member" "cicd" {
  for_each = toset(local.cicd_roles)
  project  = var.project_id
  role     = each.value
  member   = "serviceAccount:${google_service_account.cicd.email}"
}

# Workload Identity Pool — the namespace for external identity providers
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-actions"
  display_name              = "GitHub Actions"
  description               = "WIF pool for GitHub Actions keyless authentication"
}

# Workload Identity Provider — maps GitHub OIDC tokens to GCP identities
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "GitHub OIDC"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  # Map GitHub token claims to GCP attributes for condition matching
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.repository" = "assertion.repository"
  }

  # Only allow tokens from the Myrmex repository
  attribute_condition = "assertion.repository == 'HuynhHoangPhuc/myrmex'"
}

# Allow GitHub Actions (from the Myrmex repo) to impersonate the cicd service account
resource "google_service_account_iam_member" "cicd_wif_binding" {
  service_account_id = google_service_account.cicd.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/HuynhHoangPhuc/myrmex"
}

# Outputs — copy these values into GitHub repository secrets
output "wif_provider" {
  description = "Value for GCP_WIF_PROVIDER GitHub secret"
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "cicd_sa_email" {
  description = "Value for GCP_SA_EMAIL GitHub secret"
  value       = google_service_account.cicd.email
}
