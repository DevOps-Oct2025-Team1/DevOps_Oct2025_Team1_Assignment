resource "google_compute_network" "devops_vpc" {
  name                    = "devops-vpc"
  routing_mode            = "REGIONAL"
  auto_create_subnetworks = false
}

