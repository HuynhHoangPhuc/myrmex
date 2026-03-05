# Monitoring resources: uptime checks and alert policies for Myrmex ERP

resource "google_monitoring_uptime_check_config" "core_health" {
  display_name = "Myrmex Core API Health"
  timeout      = "10s"
  period       = "300s"

  http_check {
    path         = "/health"
    port         = 443
    use_ssl      = true
    validate_ssl = true
  }

  monitored_resource {
    type = "uptime_url"
    labels = {
      project_id = var.project_id
      host       = "core-placeholder.run.app" # updated after first deploy
    }
  }
}

resource "google_monitoring_alert_policy" "high_error_rate" {
  display_name = "Myrmex High Error Rate"
  combiner     = "OR"

  conditions {
    display_name = "5xx error rate > 5%"
    condition_threshold {
      filter          = "resource.type=\"cloud_run_revision\" AND metric.type=\"run.googleapis.com/request_count\" AND metric.labels.response_code_class=\"5xx\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.05
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = [] # Add email/Slack channel IDs after setup

  alert_strategy {
    auto_close = "604800s"
  }
}

resource "google_monitoring_alert_policy" "cloudsql_connections" {
  display_name = "Myrmex Cloud SQL High Connections"
  combiner     = "OR"

  conditions {
    display_name = "DB connections > 20"
    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND metric.type=\"cloudsql.googleapis.com/database/postgresql/num_backends\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 20
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = []

  alert_strategy {
    auto_close = "604800s"
  }
}
