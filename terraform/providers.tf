terraform {
  required_version = ">= 1.3"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.16.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.24.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.5.2"
    }
  }
}

provider "google" {
  project = local.project_id
  region  = local.region
}

provider "kubernetes" {
  host  = "https://${google_container_cluster.devops-gke.endpoint}"
  token = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(
    google_container_cluster.devops-gke.master_auth[0].cluster_ca_certificate
  )
}

data "google_client_config" "default" {}
