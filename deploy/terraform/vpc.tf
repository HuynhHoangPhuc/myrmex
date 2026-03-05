# VPC network for Myrmex — private connectivity between Cloud Run, Cloud SQL, and Redis
resource "google_compute_network" "myrmex" {
  name                    = "myrmex-vpc"
  auto_create_subnetworks = true # use auto-mode subnets per region

  depends_on = [google_project_service.apis]
}

# Serverless VPC Access connector — allows Cloud Run to reach private VPC resources
# (Cloud SQL private IP, Memorystore Redis)
resource "google_vpc_access_connector" "myrmex" {
  name          = "myrmex-connector"
  region        = var.region
  network       = google_compute_network.myrmex.name
  ip_cidr_range = "10.8.0.0/28" # /28 = 14 usable IPs, sufficient for connector

  min_instances = 2
  max_instances = 10
  machine_type  = "f1-micro"

  depends_on = [google_project_service.apis]
}

# Private services access — required for Cloud SQL private IP and Memorystore
resource "google_compute_global_address" "private_services_range" {
  name          = "myrmex-private-services-range"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.myrmex.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.myrmex.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_services_range.name]

  depends_on = [google_project_service.apis]
}
