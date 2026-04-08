# Outputs for GCE staging VM
output "staging_gce_ip" {
  description = "Static IP of the GCE staging VM"
  value       = var.staging_gce_enabled ? google_compute_address.staging[0].address : ""
}

output "staging_backup_bucket" {
  description = "GCS bucket name for staging backups"
  value       = var.staging_gce_enabled ? google_storage_bucket.staging_backups[0].name : ""
}
