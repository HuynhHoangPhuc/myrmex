# Cloud Run custom domain mappings — requires DNS CNAME records pointing to ghs.googlehosted.com
# Set domain variables in terraform.tfvars after coordinating with HCMUS IT.
# GCP auto-provisions managed SSL certificates once DNS propagates (~10 min).

variable "frontend_domain" {
  description = "Custom domain for the frontend (e.g. myrmex.hcmus.edu.vn). Leave empty to skip."
  type        = string
  default     = ""
}

variable "api_domain" {
  description = "Custom domain for the core API (e.g. api.myrmex.hcmus.edu.vn). Leave empty to skip."
  type        = string
  default     = ""
}

resource "google_cloud_run_domain_mapping" "frontend" {
  count    = var.frontend_domain != "" ? 1 : 0
  name     = var.frontend_domain
  location = var.region

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.frontend.name
  }
}

resource "google_cloud_run_domain_mapping" "api" {
  count    = var.api_domain != "" ? 1 : 0
  name     = var.api_domain
  location = var.region

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.core.name
  }
}

# After applying, retrieve DNS records to configure at your DNS provider:
# terraform output domain_mapping_records
output "frontend_domain_dns_records" {
  description = "DNS records to add for the frontend custom domain"
  value       = length(google_cloud_run_domain_mapping.frontend) > 0 ? google_cloud_run_domain_mapping.frontend[0].status[*].resource_records : []
}

output "api_domain_dns_records" {
  description = "DNS records to add for the API custom domain"
  value       = length(google_cloud_run_domain_mapping.api) > 0 ? google_cloud_run_domain_mapping.api[0].status[*].resource_records : []
}
