# Cloud SQL PostgreSQL 16 — shared database for all Myrmex modules (schema-per-module)
resource "google_sql_database_instance" "postgres" {
  name             = "myrmex-postgres"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier = var.db_instance_tier

    # Private IP only — no public IP exposure; SSL required for all connections
    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.myrmex.id
      enable_private_path_for_google_cloud_services = true
      ssl_mode                                      = "ENCRYPTED_ONLY"
    }

    # Daily automated backups with 7-day retention
    backup_configuration {
      enabled                        = true
      start_time                     = "02:00" # 02:00 UTC = 09:00 ICT
      point_in_time_recovery_enabled = true
      backup_retention_settings {
        retained_backups = 7
        retention_unit   = "COUNT"
      }
    }

    # Maintenance window — Sunday 03:00 UTC
    maintenance_window {
      day          = 7
      hour         = 3
      update_track = "stable"
    }

    database_flags {
      name  = "max_connections"
      value = "200"
    }

    insights_config {
      query_insights_enabled = true
    }
  }

  # Prevent accidental destruction of production database
  lifecycle {
    prevent_destroy = true
  }

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

# Application database
resource "google_sql_database" "myrmex" {
  name     = "myrmex"
  instance = google_sql_database_instance.postgres.name
}

# ---------------------------------------------------------------------------
# Staging Cloud SQL — separate instance, no prevent_destroy, cheaper config
# ---------------------------------------------------------------------------
resource "google_sql_database_instance" "postgres_staging" {
  name             = "myrmex-postgres-staging"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.myrmex.id
      enable_private_path_for_google_cloud_services = true
      ssl_mode                                      = "ENCRYPTED_ONLY"
    }

    backup_configuration {
      enabled = false # no PITR for staging
    }

    database_flags {
      name  = "max_connections"
      value = "50"
    }
  }

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

resource "google_sql_database" "myrmex_staging" {
  name     = "myrmex"
  instance = google_sql_database_instance.postgres_staging.name
}
