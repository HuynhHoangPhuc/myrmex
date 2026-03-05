# Pub/Sub topics — NATS JetStream events bridged to GCP Pub/Sub for inter-module messaging
locals {
  topics = [
    "hr-events",
    "subject-events",
    "student-events",
    "timetable-events",
    "audit-events",
    "notification-events",
    "core-events",
  ]
}

resource "google_pubsub_topic" "topics" {
  for_each = toset(local.topics)

  name = each.value

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }

  depends_on = [google_project_service.apis]
}

# Pull subscriptions for consumers

# audit-consumer: persists audit events to partitioned audit_logs table
resource "google_pubsub_subscription" "audit_consumer" {
  name  = "audit-consumer"
  topic = google_pubsub_topic.topics["audit-events"].name

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# notification-consumer: processes core-events to dispatch in-app / email notifications
resource "google_pubsub_subscription" "notification_consumer" {
  name  = "notification-consumer"
  topic = google_pubsub_topic.topics["core-events"].name

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# analytics-consumer: processes student-events for analytics aggregation
resource "google_pubsub_subscription" "analytics_consumer" {
  name  = "analytics-consumer"
  topic = google_pubsub_topic.topics["student-events"].name

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# analytics-hr-consumer: processes hr-events for analytics aggregation
resource "google_pubsub_subscription" "analytics_hr_consumer" {
  name  = "analytics-hr-consumer"
  topic = google_pubsub_topic.topics["hr-events"].name

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}
