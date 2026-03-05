# Cloud Run v2 services — 8 services total (1 frontend + 1 core gateway + 6 modules)
# All services use the myrmex-run service account and VPC connector for private access.

locals {
  image_base = "${var.region}-docker.pkg.dev/${var.project_id}/myrmex"
  tag        = var.docker_image_tag

  # Common secret env vars injected into every service
  common_secrets = [
    { name = "DATABASE_URL",      secret = "DATABASE_URL" },
    { name = "REDIS_ADDR",        secret = "REDIS_ADDR" },
    { name = "JWT_SECRET",        secret = "JWT_SECRET" },
    { name = "MESSAGING_BACKEND", secret = "MESSAGING_BACKEND" },
    { name = "GCP_PROJECT_ID",    secret = "GCP_PROJECT_ID" },
  ]
}

# ---------------------------------------------------------------------------
# frontend — React SPA served via Node/nginx container
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "frontend" {
  name     = "frontend"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
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
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      env {
        name  = "VITE_API_BASE_URL"
        value = "" # populated post-deploy via CI referencing core_url output
      }
    }
  }

  depends_on = [
    google_project_iam_member.run_sa_roles,
    google_artifact_registry_repository.myrmex,
  ]
}

# Allow unauthenticated public access to frontend
resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# ---------------------------------------------------------------------------
# core — Gin HTTP gateway + WebSocket + gRPC-gateway (HTTP/2)
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "core" {
  name     = "core"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = var.core_min_instances
      max_instance_count = 20
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/core:${local.tag}"

      ports {
        container_port = 8080
        name           = "h2c" # HTTP/2 cleartext for WebSocket + gRPC-gateway
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }

      # gRPC addresses for each module (internal Cloud Run URLs)
      env {
        name  = "SERVER_GRPC_HR_ADDR"
        value = "module-hr-${var.project_id}-${var.region}.a.run.app:443"
      }
      env {
        name  = "SERVER_GRPC_SUBJECT_ADDR"
        value = "module-subject-${var.project_id}-${var.region}.a.run.app:443"
      }
      env {
        name  = "SERVER_GRPC_TIMETABLE_ADDR"
        value = "module-timetable-${var.project_id}-${var.region}.a.run.app:443"
      }
      env {
        name  = "SERVER_GRPC_STUDENT_ADDR"
        value = "module-student-${var.project_id}-${var.region}.a.run.app:443"
      }
      env {
        name  = "SERVER_GRPC_ANALYTICS_ADDR"
        value = "module-analytics-${var.project_id}-${var.region}.a.run.app:443"
      }
      env {
        name  = "SERVER_GRPC_NOTIFICATION_ADDR"
        value = "module-notification-${var.project_id}-${var.region}.a.run.app:443"
      }

      # Secrets
      dynamic "env" {
        for_each = local.common_secrets
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

      dynamic "env" {
        for_each = [
          { name = "OAUTH_GOOGLE_CLIENT_ID",       secret = "OAUTH_GOOGLE_CLIENT_ID" },
          { name = "OAUTH_GOOGLE_CLIENT_SECRET",   secret = "OAUTH_GOOGLE_CLIENT_SECRET" },
          { name = "OAUTH_MICROSOFT_CLIENT_ID",    secret = "OAUTH_MICROSOFT_CLIENT_ID" },
          { name = "OAUTH_MICROSOFT_CLIENT_SECRET", secret = "OAUTH_MICROSOFT_CLIENT_SECRET" },
          { name = "OAUTH_MICROSOFT_TENANT_ID",    secret = "OAUTH_MICROSOFT_TENANT_ID" },
          { name = "LLM_API_KEY",                  secret = "LLM_API_KEY" },
          { name = "LLM_PROVIDER",                 secret = "LLM_PROVIDER" },
          { name = "LLM_MODEL",                    secret = "LLM_MODEL" },
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

  depends_on = [
    google_project_iam_member.run_sa_roles,
    google_sql_database_instance.postgres,
    google_redis_instance.redis,
  ]
}

# Allow unauthenticated public access to core API
resource "google_cloud_run_v2_service_iam_member" "core_public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.core.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# ---------------------------------------------------------------------------
# Internal module services — gRPC, internal ingress only
# ---------------------------------------------------------------------------

# module-hr (port 50052)
resource "google_cloud_run_v2_service" "module_hr" {
  name     = "module-hr"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-hr:${local.tag}"

      ports {
        container_port = 50052
        name           = "h2c"
      }

      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

# module-subject (port 50053)
resource "google_cloud_run_v2_service" "module_subject" {
  name     = "module-subject"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-subject:${local.tag}"

      ports {
        container_port = 50053
        name           = "h2c"
      }

      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

# module-timetable (port 50054) — CSP solver, needs more memory
resource "google_cloud_run_v2_service" "module_timetable" {
  name     = "module-timetable"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-timetable:${local.tag}"

      ports {
        container_port = 50054
        name           = "h2c"
      }

      resources {
        limits = {
          memory = "512Mi" # CSP backtracking + AC-3 solver is memory-intensive
          cpu    = "2"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

# module-student (port 50055)
resource "google_cloud_run_v2_service" "module_student" {
  name     = "module-student"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-student:${local.tag}"

      ports {
        container_port = 50055
        name           = "h2c"
      }

      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

# module-analytics (port 8055, HTTP/1)
resource "google_cloud_run_v2_service" "module_analytics" {
  name     = "module-analytics"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-analytics:${local.tag}"

      ports {
        container_port = 8055
        name           = "http1"
      }

      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

# module-notification (port 8056, HTTP/1) — min 1 instance for reliable delivery
resource "google_cloud_run_v2_service" "module_notification" {
  name     = "module-notification"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.myrmex_run.email

    scaling {
      min_instance_count = var.notification_min_instances
      max_instance_count = 10
    }

    vpc_access {
      connector = google_vpc_access_connector.myrmex.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "${local.image_base}/module-notification:${local.tag}"

      ports {
        container_port = 8056
        name           = "http1"
      }

      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1"
        }
      }

      dynamic "env" {
        for_each = local.common_secrets
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

      dynamic "env" {
        for_each = [
          { name = "SMTP_HOST",       secret = "SMTP_HOST" },
          { name = "SMTP_PORT",       secret = "SMTP_PORT" },
          { name = "SMTP_USERNAME",   secret = "SMTP_USERNAME" },
          { name = "SMTP_PASSWORD",   secret = "SMTP_PASSWORD" },
          { name = "SMTP_FROM_EMAIL", secret = "SMTP_FROM_EMAIL" },
          { name = "SMTP_FROM_NAME",  secret = "SMTP_FROM_NAME" },
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

  depends_on = [google_project_iam_member.run_sa_roles]
}
