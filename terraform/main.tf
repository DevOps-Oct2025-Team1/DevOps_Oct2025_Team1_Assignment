# Cloud Storage Bucket for Application Artifacts
resource "google_artifact_registry_repository" "docker_repo" {
  provider = google-beta

  location      = var.region
  repository_id = "${var.environment}-docker-repo"
  description   = "Docker repository for ${var.environment}"
  format        = "DOCKER"

  labels = local.common_labels
}

# Cloud SQL (PostgreSQL) - Auth Database
resource "google_sql_database_instance" "auth_postgres" {
  name             = "${var.environment}-auth-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = var.environment == "production" ? "db-g1-small" : "db-f1-micro"

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.devops_vpc.id
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
    google_project_service.devops_api,
    google_compute_network.devops_vpc
  ]
}

resource "google_sql_database" "auth_db" {
  name     = "auth_db"
  instance = google_sql_database_instance.auth_postgres.name
}

# Cloud SQL (PostgreSQL) - Chats Database
resource "google_sql_database_instance" "chats_postgres" {
  name             = "${var.environment}-chats-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = var.environment == "production" ? "db-g1-small" : "db-f1-micro"

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.devops_vpc.id
    }

    backup_configuration {
      enabled = var.environment == "production"
    }
  }

  deletion_protection = var.enable_deletion_protection

  depends_on = [
    google_project_service.devops_api,
    google_compute_network.devops_vpc
  ]
}

resource "google_sql_database" "chats_db" {
  name     = "chats_db"
  instance = google_sql_database_instance.chats_postgres.name
}

# Pub/Sub Topic and Subscription
resource "google_pubsub_topic" "prompt_requests" {
  name = "${var.environment}-prompt-requests"

  labels = local.common_labels

  depends_on = [google_project_service.devops_api]
}

resource "google_pubsub_subscription" "prompt_requests" {
  name  = "${var.environment}-prompt-requests-sub"
  topic = google_pubsub_topic.prompt_requests.name

  labels = local.common_labels
}
