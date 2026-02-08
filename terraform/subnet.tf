resource "google_compute_subnetwork" "devops_subnet_private" {
  name                     = "devops-subnet-private"
  region                   = local.region
  network                  = google_compute_network.devops_vpc.id
  private_ip_google_access = true
  stack_type               = "IPV4_ONLY"
  ip_cidr_range            = var.vpc_cidr

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = var.services_cidr
  }
}


