resource "google_container_cluster" "devops_gke" {
  name                     = "devops-gke"
  location                 = local.region
  network                  = google_compute_network.devops_vpc.id
  subnetwork               = google_compute_subnetwork.devops_subnet_private.id
  networking_mode          = "VPC_NATIVE"
  deletion_protection      = var.enable_deletion_protection
  remove_default_node_pool = false
  initial_node_count       = var.gke_node_count
  node_locations           = local.gke_zones

  ip_allocation_policy {
    cluster_secondary_range_name  = var.subnet_pods_range
    services_secondary_range_name = var.subnet_services_range
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = var.master_cidr
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = local.ip
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
      minimum       = var.cluster_min_cpu
      maximum       = var.cluster_max_cpu
    }
    resource_limits {
      resource_type = "memory"
      minimum       = var.cluster_min_memory
      maximum       = var.cluster_max_memory
    }
  }
}

resource "google_container_node_pool" "devops_node_pool" {
  name       = "devops-node-pool"
  cluster    = google_container_cluster.devops_gke.name
  location   = local.region
  node_count = var.gke_node_count

  autoscaling {
    min_node_count = var.gke_min_node_count
    max_node_count = var.gke_max_node_count
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = var.gke_machine_type
    disk_type    = "pd-standard"
    disk_size_gb = var.gke_disk_size

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
