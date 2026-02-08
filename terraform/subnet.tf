resource "google_compute_subnetwork" "devops-subnet-private" {
  name                     = "devops-subnet-private"
  region                   = local.region
  network                  = google_compute_network.devops_vpc.id
  private_ip_google_access = true
  stack_type               = "IPV4_ONLY"
  ip_cidr_range            = "10.20.0.0/16"

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.21.0.0/16"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.22.0.0/20"
  }
}


