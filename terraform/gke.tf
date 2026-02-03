resource "google_container_cluster" "devops-gke" {
  name                     = "devops-gke"
  location                 = local.region
  network                  = google_compute_network.devops-vpc.id
  subnetwork               = google_compute_subnetwork.devops-subnet-private.id
  networking_mode          = "VPC_NATIVE"
  deletion_protection      = false
  remove_default_node_pool = true
  initial_node_count       = 1

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"
      display_name = "Allow All"
    }
  }
}

resource "google_container_node_pool" "devops-node-pool" {
  name     = "devops-node-pool"
  cluster  = google_container_cluster.devops-gke.name
  location = local.region

  autoscaling {
    min_node_count = 1
    max_node_count = 5
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = "e2-standard-4"
    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
}

resource "kubernetes_service" "frontend" {
  metadata {
    name      = "frontend"
    namespace = "default"
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "http"
      port        = 80
      target_port = 80
      protocol    = "TCP"
    }

    selector = {
      app = "frontend"
    }
  }
}

resource "kubernetes_service" "api-gateway" {
  metadata {
    name      = "api-gateway"
    namespace = "default"
  }

  spec {
    type = "LoadBalancer"

    port {
      port        = 8080
      target_port = 8080
    }

    selector = {
      app = "api-gateway"
    }
  }
}

