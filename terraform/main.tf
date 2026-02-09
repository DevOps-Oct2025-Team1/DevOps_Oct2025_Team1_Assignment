# =============================================================================
# Enable Required APIs
# =============================================================================
resource "google_project_service" "apis" {
  for_each = toset(local.apis)

  project = local.project_id
  service = each.value

  disable_on_destroy = false
}

# =============================================================================
# VPC Network
# =============================================================================
resource "google_compute_network" "vpc" {
  name                    = "${var.environment}-vpc"
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"

  depends_on = [google_project_service.apis]
}

# =============================================================================
# Subnet
# =============================================================================
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.environment}-subnet"
  ip_cidr_range = var.vpc_cidr
  region        = var.region
  network       = google_compute_network.vpc.id

  private_ip_google_access = true

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = var.services_cidr
  }
}

# =============================================================================
# Cloud Router for NAT
# =============================================================================
resource "google_compute_router" "router" {
  name    = "${var.environment}-router"
  region  = var.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.environment}-nat"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

# =============================================================================
# Firewall Rules
# =============================================================================
resource "google_compute_firewall" "iap_access" {
  name    = "${var.environment}-allow-iap"
  network = google_compute_network.vpc.name

  allow {
    protocol = "tcp"
    ports    = ["22", "3389"]
  }

  source_ranges = [var.ip_range_iap]
}

# =============================================================================
# GKE Cluster
# =============================================================================
resource "google_container_cluster" "primary" {
  name     = "${var.environment}-gke-cluster"
  location = var.region
  project  = local.project_id

  network    = google_compute_network.vpc.id
  subnetwork = google_compute_subnetwork.subnet.id

  # Private cluster configuration
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = var.master_cidr

    master_global_access_config {
      enabled = true
    }
  }

  # IP allocation policy
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # Master authorized networks (allow IAP)
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = var.ip_range_iap
      display_name = "IAP"
    }
  }

  # Release channel
  release_channel {
    channel = "REGULAR"
  }

  # Network policy
  network_policy {
    enabled = true
  }

  # Enable workload identity
  workload_identity_config {
    workload_pool = "${local.project_id}.svc.id.goog"
  }

  # Initial node pool (will be deleted after creation)
  initial_node_count = 1

  node_config {
    machine_type = var.gke_machine_type
    disk_size_gb = var.gke_disk_size

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = local.common_labels
  }

  # Remove default node pool after cluster creation
  remove_default_node_pool = true

  depends_on = [
    google_project_service.apis,
    google_compute_subnetwork.subnet
  ]
}

# =============================================================================
# GKE Node Pool
# =============================================================================
resource "google_container_node_pool" "primary" {
  name       = "${var.environment}-node-pool"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = var.gke_node_count

  node_locations = var.gke_zones

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
    disk_size_gb = var.gke_disk_size

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = local.common_labels

    tags = ["${var.environment}-gke-node"]
  }

  depends_on = [google_container_cluster.primary]
}

# =============================================================================
# Cloud Storage Bucket for Terraform State
# =============================================================================
resource "google_storage_bucket" "terraform_state" {
  name          = "${local.project_id}-terraform-state"
  location      = var.region
  force_destroy = !var.enable_deletion_protection

  versioning {
    enabled = var.bucket_versioning
  }

  uniform_bucket_level_access = true

  labels = local.common_labels

  lifecycle {
    prevent_destroy = true
  }
}

# =============================================================================
# Cloud Storage Bucket for Application Artifacts
# =============================================================================
resource "google_artifact_registry_repository" "docker_repo" {
  provider = google-beta

  location      = var.region
  repository_id = "${var.environment}-docker-repo"
  description   = "Docker repository for ${var.environment}"
  format        = "DOCKER"

  labels = local.common_labels
}

# =============================================================================
# Cloud SQL (PostgreSQL) - Auth Database
# =============================================================================
resource "google_sql_database_instance" "auth_postgres" {
  name             = "${var.environment}-auth-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = var.environment == "production" ? "db-g1-small" : "db-f1-micro"

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.vpc.id
    }

    backup_configuration {
      enabled = var.environment == "production"
    }

    maintenance_window {
      day  = 7
      hour = 3
    }
  }

  deletion_protection = var.enable_deletion_protection

  depends_on = [
    google_project_service.apis,
    google_compute_network.vpc
  ]
}

resource "google_sql_database" "auth_db" {
  name     = "auth_db"
  instance = google_sql_database_instance.auth_postgres.name
}

# =============================================================================
# Cloud SQL (PostgreSQL) - Chats Database
# =============================================================================
resource "google_sql_database_instance" "chats_postgres" {
  name             = "${var.environment}-chats-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = var.environment == "production" ? "db-g1-small" : "db-f1-micro"

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.vpc.id
    }

    backup_configuration {
      enabled = var.environment == "production"
    }
  }

  deletion_protection = var.enable_deletion_protection

  depends_on = [
    google_project_service.apis,
    google_compute_network.vpc
  ]
}

resource "google_sql_database" "chats_db" {
  name     = "chats_db"
  instance = google_sql_database_instance.chats_postgres.name
}

# =============================================================================
# Pub/Sub Topic and Subscription
# =============================================================================
resource "google_pubsub_topic" "prompt_requests" {
  name = "${var.environment}-prompt-requests"

  labels = local.common_labels

  depends_on = [google_project_service.apis]
}

resource "google_pubsub_subscription" "prompt_requests" {
  name  = "${var.environment}-prompt-requests-sub"
  topic = google_pubsub_topic.prompt_requests.name

  labels = local.common_labels
}

# =============================================================================
# Service Accounts
# =============================================================================
resource "google_service_account" "gke_nodes" {
  account_id   = "${var.environment}-gke-nodes"
  display_name = "GKE Node Service Account"
  description  = "Service account for GKE nodes in ${var.environment}"
}

resource "google_project_iam_member" "gke_nodes_roles" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/monitoring.viewer",
    "roles/stackdriver.resourceMetadata.writer",
    "roles/autoscaling.metricsWriter",
    "roles/artifactregistry.reader"
  ])

  project = local.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}
