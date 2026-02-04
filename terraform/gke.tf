resource "google_container_cluster" "devops-gke" {
  name                     = "devops-gke"
  location                 = "${local.region}-a"
  network                  = google_compute_network.devops-vpc.id
  subnetwork               = google_compute_subnetwork.devops-subnet-private.id
  networking_mode          = "VPC_NATIVE"
  deletion_protection      = false
  remove_default_node_pool = true
  initial_node_count       = 1

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "35.235.240.0/20"
      display_name = "IAP"
    }
  }

  workload_identity_config {
    workload_pool = "${local.project_id}.svc.id.goog"
  }
}

resource "google_container_node_pool" "devops-node-pool" {
  name     = "devops-node-pool"
  cluster  = google_container_cluster.devops-gke.name
  location = local.region
  node_count         = 3

  autoscaling {
    min_node_count = 2
    max_node_count = 4
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type    = "e2-standard-4"
    disk_size_gb    = 30
    service_account = google_service_account.devops_compute_gke_user.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
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

resource "kubernetes_service" "api_gateway" {
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

resource "kubernetes_service_account_v1" "default" {
  metadata {
    name      = "default"
    namespace = "default"
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.devops_compute_gke_user.email
    }
  }

  automount_service_account_token = true
}

