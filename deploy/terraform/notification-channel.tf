# Monitoring notification channels — email and optional Slack webhook
# Channels are wired into alert policies in monitoring.tf
# Set alert_email and alert_slack_webhook_url in terraform.tfvars to activate

resource "google_monitoring_notification_channel" "email" {
  count        = var.alert_email != "" ? 1 : 0
  display_name = "Myrmex Ops Email"
  type         = "email"

  labels = {
    email_address = var.alert_email
  }
}

resource "google_monitoring_notification_channel" "slack" {
  count        = var.alert_slack_webhook_url != "" ? 1 : 0
  display_name = "Myrmex Ops Slack"
  type         = "slack"

  sensitive_labels {
    auth_token = var.alert_slack_webhook_url
  }
}

# Aggregated list of all active notification channel IDs for use in alert policies
locals {
  notification_channel_ids = concat(
    [for ch in google_monitoring_notification_channel.email : ch.id],
    [for ch in google_monitoring_notification_channel.slack : ch.id],
  )
}
