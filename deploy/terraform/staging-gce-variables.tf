# Variables for GCE-based staging VM (all gated behind staging_gce_enabled)
variable "staging_gce_enabled" {
  description = "Enable GCE-based staging VM"
  type        = bool
  default     = false
}

variable "staging_gce_machine_type" {
  description = "GCE machine type for staging VM"
  type        = string
  default     = "e2-medium"
}

variable "staging_gce_zone" {
  description = "GCE zone for staging VM"
  type        = string
  default     = "asia-southeast1-b"
}

variable "staging_gce_disk_size_gb" {
  description = "Boot disk size in GB"
  type        = number
  default     = 20
}

variable "staging_domain" {
  description = "Domain for staging (e.g., staging.internalsystem.org)"
  type        = string
  default     = ""
}

variable "staging_ssh_public_key" {
  description = "SSH public key for deploy user (GitHub Actions CD)"
  type        = string
  default     = ""
}
