# Staging Cloud Run services — mirrors production with staging- prefix
# All services use min_instances=0 to minimize staging costs.
# Staging uses DATABASE_URL_STAGING and JWT_SECRET_STAGING secrets.

locals {
  # Staging common secrets (same as prod except DB + JWT)
  staging_common_secrets = [
    { name = "DATABASE_URL",      secret = "DATABASE_URL_STAGING" },
    { name = "REDIS_ADDR",        secret = "REDIS_ADDR" },
    { name = "JWT_SECRET",        secret = "JWT_SECRET_STAGING" },
    { name = "MESSAGING_BACKEND", secret = "MESSAGING_BACKEND" },
    { name = "GCP_PROJECT_ID",    secret = "GCP_PROJECT_ID" },
  ]

  # Module gRPC services: name → port mapping
  staging_module_services = {
    module-hr          = { port = 50052, memory = "256Mi", cpu = "1" }
    module-subject     = { port = 50053, memory = "256Mi", cpu = "1" }
    module-timetable   = { port = 50054, memory = "512Mi", cpu = "2" }
    module-student     = { port = 50055, memory = "256Mi", cpu = "1" }
    module-analytics   = { port = 8055,  memory = "256Mi", cpu = "1" }
    module-notification = { port = 8056, memory = "256Mi", cpu = "1" }
  }
}

# ---------------------------------------------------------------------------
# staging-frontend
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "staging_frontend" {
  name     = "staging-frontend"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.myrmex_run.email
    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }
    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
    containers {
      image = "${local.image_base}/frontend:${local.tag}"
      ports {
        container_port = 3000
        name           = "http1"
      }
      resources {
        limits = { memory = "256Mi", cpu = "1" }
      }
      env {
        name  = "VITE_API_BASE_URL"
        value = "" # set post-deploy to staging-core URL
      }
    }
  }

  depends_on = [google_project_iam_member.run_sa_roles, google_artifact_registry_repository.myrmex]
}

resource "google_cloud_run_v2_service_iam_member" "staging_frontend_public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.staging_frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# ---------------------------------------------------------------------------
# staging-core
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "staging_core" {
  name     = "staging-core"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.myrmex_run.email
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }
    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
    containers {
      image = "${local.image_base}/core:${local.tag}"
      ports {
        container_port = 8080
        name           = "h2c"
      }
      resources {
        limits = { memory = "512Mi", cpu = "1" }
      }

      # Module gRPC addresses pointing to staging- services
      env { name = "SERVER_GRPC_HR_ADDR",           value = "staging-module-hr-${var.project_id}-${var.region}.a.run.app:443" }
      env { name = "SERVER_GRPC_SUBJECT_ADDR",       value = "staging-module-subject-${var.project_id}-${var.region}.a.run.app:443" }
      env { name = "SERVER_GRPC_TIMETABLE_ADDR",     value = "staging-module-timetable-${var.project_id}-${var.region}.a.run.app:443" }
      env { name = "SERVER_GRPC_STUDENT_ADDR",       value = "staging-module-student-${var.project_id}-${var.region}.a.run.app:443" }
      env { name = "SERVER_GRPC_ANALYTICS_ADDR",     value = "staging-module-analytics-${var.project_id}-${var.region}.a.run.app:443" }
      env { name = "SERVER_GRPC_NOTIFICATION_ADDR",  value = "staging-module-notification-${var.project_id}-${var.region}.a.run.app:443" }

      dynamic "env" {
        for_each = local.staging_common_secrets
        content {
          name = env.value.name
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.secrets[env.value.secret].secret_id
              version = "latest"
            }
          }
        }
      }

      # OAuth secrets (needed for HCMUS SSO testing in UAT)
      dynamic "env" {
        for_each = [
          { name = "OAUTH_GOOGLE_CLIENT_ID",        secret = "OAUTH_GOOGLE_CLIENT_ID" },
          { name = "OAUTH_GOOGLE_CLIENT_SECRET",    secret = "OAUTH_GOOGLE_CLIENT_SECRET" },
          { name = "OAUTH_MICROSOFT_CLIENT_ID",     secret = "OAUTH_MICROSOFT_CLIENT_ID" },
          { name = "OAUTH_MICROSOFT_CLIENT_SECRET", secret = "OAUTH_MICROSOFT_CLIENT_SECRET" },
          { name = "OAUTH_MICROSOFT_TENANT_ID",     secret = "OAUTH_MICROSOFT_TENANT_ID" },
        ]
        content {
          name = env.value.name
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.secrets[env.value.secret].secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_iam_member.run_sa_roles, google_sql_database_instance.postgres_staging]
}

resource "google_cloud_run_v2_service_iam_member" "staging_core_public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.staging_core.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# ---------------------------------------------------------------------------
# Staging module services — gRPC, internal ingress, for_each
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "staging_modules" {
  for_each = local.staging_module_services

  name     = "staging-${each.key}"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }
    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
    containers {
      image = "${local.image_base}/${each.key}:${local.tag}"
      ports {
        container_port = each.value.port
        name           = each.value.port == 8055 || each.value.port == 8056 ? "http1" : "h2c"
      }
      resources {
        limits = { memory = each.value.memory, cpu = each.value.cpu }
      }

      dynamic "env" {
        for_each = local.staging_common_secrets
        content {
          name = env.value.name
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.secrets[env.value.secret].secret_id
              version = "latest"
            }
          }
        }
      }

      # SMTP secrets for staging-module-notification
      dynamic "env" {
        for_each = each.key == "module-notification" ? [
          { name = "SMTP_HOST",       secret = "SMTP_HOST" },
          { name = "SMTP_PORT",       secret = "SMTP_PORT" },
          { name = "SMTP_USERNAME",   secret = "SMTP_USERNAME" },
          { name = "SMTP_PASSWORD",   secret = "SMTP_PASSWORD" },
          { name = "SMTP_FROM_EMAIL", secret = "SMTP_FROM_EMAIL" },
          { name = "SMTP_FROM_NAME",  secret = "SMTP_FROM_NAME" },
        ] : []
        content {
          name = env.value.name
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.secrets[env.value.secret].secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_iam_member.run_sa_roles]
}
