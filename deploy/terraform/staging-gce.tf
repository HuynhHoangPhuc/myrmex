# GCE staging VM — single VM running Docker Compose with Caddy HTTPS
# All resources gated behind var.staging_gce_enabled

# Service account for staging VM (GCS backup access)
resource "google_service_account" "staging_gce" {
  count        = var.staging_gce_enabled ? 1 : 0
  account_id   = "myrmex-staging-gce"
  display_name = "Myrmex Staging GCE VM"
}

# GCS write permission for backups
resource "google_project_iam_member" "staging_gce_gcs" {
  count   = var.staging_gce_enabled ? 1 : 0
  project = var.project_id
  role    = "roles/storage.objectCreator"
  member  = "serviceAccount:${google_service_account.staging_gce[0].email}"
}

# Static external IP
resource "google_compute_address" "staging" {
  count  = var.staging_gce_enabled ? 1 : 0
  name   = "myrmex-staging-ip"
  region = var.region
}

# Firewall — HTTP, HTTPS (public)
resource "google_compute_firewall" "staging_allow_web" {
  count   = var.staging_gce_enabled ? 1 : 0
  name    = "myrmex-staging-allow-web"
  network = google_compute_network.myrmex.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["myrmex-staging"]
}

# Firewall — SSH via IAP only (Identity-Aware Proxy)
resource "google_compute_firewall" "staging_allow_ssh_iap" {
  count   = var.staging_gce_enabled ? 1 : 0
  name    = "myrmex-staging-allow-ssh-iap"
  network = google_compute_network.myrmex.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"] # Google IAP range
  target_tags   = ["myrmex-staging"]
}

# GCE instance
resource "google_compute_instance" "staging" {
  count        = var.staging_gce_enabled ? 1 : 0
  name         = "myrmex-staging"
  machine_type = var.staging_gce_machine_type
  zone         = var.staging_gce_zone
  tags         = ["myrmex-staging"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2404-lts-amd64"
      size  = var.staging_gce_disk_size_gb
      type  = "pd-balanced"
    }
  }

  network_interface {
    network = google_compute_network.myrmex.name
    access_config {
      nat_ip = google_compute_address.staging[0].address
    }
  }

  service_account {
    email  = google_service_account.staging_gce[0].email
    scopes = ["storage-rw"]
  }

  metadata_startup_script = file("${path.module}/scripts/staging-gce-startup.sh")

  metadata = {
    ssh-keys = var.staging_ssh_public_key != "" ? "deploy:${var.staging_ssh_public_key}" : null
  }
}

# GCS bucket for backups
resource "google_storage_bucket" "staging_backups" {
  count    = var.staging_gce_enabled ? 1 : 0
  name     = "${var.project_id}-staging-backups"
  location = var.region

  lifecycle_rule {
    condition { age = 30 }
    action { type = "Delete" }
  }

  versioning {
    enabled = true
  }

  uniform_bucket_level_access = true
}
