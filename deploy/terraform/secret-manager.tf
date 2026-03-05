# Secret Manager — empty placeholder secrets; values must be set externally via gcloud CLI
# gcloud secrets versions add SECRET_NAME --data-file=<(echo -n "value")
locals {
  secrets = [
    "DATABASE_URL",
    "REDIS_ADDR",
    "JWT_SECRET",
    "MESSAGING_BACKEND",
    "GCP_PROJECT_ID",
    "OAUTH_GOOGLE_CLIENT_ID",
    "OAUTH_GOOGLE_CLIENT_SECRET",
    "OAUTH_MICROSOFT_CLIENT_ID",
    "OAUTH_MICROSOFT_CLIENT_SECRET",
    "OAUTH_MICROSOFT_TENANT_ID",
    "SMTP_HOST",
    "SMTP_PORT",
    "SMTP_USERNAME",
    "SMTP_PASSWORD",
    "SMTP_FROM_EMAIL",
    "SMTP_FROM_NAME",
    "LLM_API_KEY",
    "LLM_PROVIDER",
    "LLM_MODEL",
  ]
}

resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(local.secrets)
  secret_id = each.value

  replication {
    auto {}
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }

  # Prevent accidental destruction of secrets in production
  lifecycle {
    prevent_destroy = true
  }

  depends_on = [google_project_service.apis]
}
