resource "google_compute_subnetwork" "devops_subnet_private" {
  name                     = "devops-subnet-private"
  region                   = local.region
  network                  = google_compute_network.devops_vpc.id
  private_ip_google_access = true
  stack_type               = "IPV4_ONLY"
  ip_cidr_range            = var.vpc_cidr

  secondary_ip_range {
    range_name    = var.subnet_pods_range
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = var.subnet_services_range
    ip_cidr_range = var.services_cidr
  }
}


