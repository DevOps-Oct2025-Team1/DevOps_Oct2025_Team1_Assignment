resource "google_container_cluster" "devops-gke" {
  name                     = "devops-gke"
  location                 = local.region
  network                  = google_compute_network.devops-vpc.id
  subnetwork               = google_compute_subnetwork.devops-subnet-private.id
  networking_mode          = "VPC_NATIVE"
  deletion_protection      = false
  remove_default_node_pool = true
  initial_node_count       = 1
  node_locations           = local.gke_zones

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

  cluster_autoscaling {
    enabled = true
    resource_limits {
      resource_type = "cpu"
      minimum       = 1
      maximum       = 6
    }
    resource_limits {
      resource_type = "memory"
      minimum       = 2
      maximum       = 12
    }
  }
}

resource "google_container_node_pool" "devops-node-pool" {
  name       = "devops-node-pool"
  cluster    = google_container_cluster.devops-gke.name
  location   = local.region
  node_count = 1

  autoscaling {
    min_node_count = 1
    max_node_count = 2
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = "e2-medium"
    disk_type    = "pd-standard"
    disk_size_gb = 30

    service_account = google_service_account.devops_compute_gke_user.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    labels = merge(local.common_labels, {
      "node-pool" = "devops-node-pool"
    })
  }

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

output "cluster_name" {
  value = google_container_cluster.devops-gke.name
}

output "cluster_endpoint" {
  value     = google_container_cluster.devops-gke.endpoint
  sensitive = true
}

output "get_credentials_command" {
  value = "gcloud container clusters get-credentials ${google_container_cluster.devops-gke.name} --region ${local.region} --project ${local.project_id}"
}