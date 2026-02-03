resource "google_compute_router" "devops-router" {
  name    = "devops-router"
  region  = local.region
  network = google_compute_network.devops-vpc.id
}

resource "google_compute_router_nat" "devops-nat" {
  name   = "devops-nat"
  region = local.region
  router = google_compute_router.devops-router.name

  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  nat_ips                            = [google_compute_address.devops-nat.self_link]

  subnetwork {
    name                    = google_compute_subnetwork.devops-subnet-private.self_link
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
}

resource "google_compute_address" "devops-nat" {
  name         = "devops-nat"
  address_type = "EXTERNAL"
  network_tier = "STANDARD"
  region       = local.region
}
