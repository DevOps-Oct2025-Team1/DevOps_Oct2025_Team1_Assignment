resource "google_compute_subnetwork" "devops-subnet-public" {
  name                     = "devops-subnet-public"
  region                   = local.region
  network                  = google_compute_network.devops-vpc.id
  private_ip_google_access = true
  stack_type               = "IPV4_ONLY"
  ip_cidr_range            = "10.10.0.0/16"
}

resource "google_compute_subnetwork" "devops-subnet-private" {
  name                     = "devops-subnet-private"
  region                   = local.region
  network                  = google_compute_network.devops-vpc.id
  private_ip_google_access = true
  stack_type               = "IPV4_ONLY"
  ip_cidr_range            = "10.20.0.0/16"
}


