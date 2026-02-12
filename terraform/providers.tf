terraform {
  required_version = ">= 1.3"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.16.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
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

  backend "gcs" {
    # Bucket name provided via -backend-config in CI/CD
    prefix = "terraform/state"
  }
}

provider "google" {
  project = local.project_id
  region  = local.region
}

provider "google-beta" {
  project = local.project_id
  region  = local.region
}

data "google_client_config" "default" {}
