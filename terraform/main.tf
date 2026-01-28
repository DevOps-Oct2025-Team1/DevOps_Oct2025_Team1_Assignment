terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.16.0"
    }
  }
}

provider "google" {
  project = "devops-project-484907"
  region  = "us-west1"
  zone    = "us-west1-a"
}

resource "google_compute_network" "terraform_vpc_network" {
  name = "terraform-network"
}

resource "google_compute_instance_template" "terraform_instance_template" {
  name         = "my-instance-template"
  machine_type = "e2-standard-4"

  disk {
    source_image = "debian-cloud/debian-11"
    disk_size_gb = 250
  }

  network_interface {
    network = google_compute_network.terraform_vpc_network.self_link
  }
}

resource "google_compute_target_pool" "terraform_target_pool" {
  name   = "my-target-pool"
  region = "us-west1"
}

resource "google_compute_region_instance_group_manager" "terraform_CRIGM" {
  name   = "my-region-igm"
  region = "us-west1"
  
  target_size = 2

  version {
    instance_template = google_compute_instance_template.terraform_instance_template.id
    name              = "primary"
  }

  target_pools       = [google_compute_target_pool.terraform_target_pool.id]
  base_instance_name = "foobar"
}

resource "google_compute_region_autoscaler" "terraform_autoscaler" {
  name   = "my-region-autoscaler"
  region = "us-west1"
  target = google_compute_region_instance_group_manager.terraform_CRIGM.id

  autoscaling_policy {
    max_replicas    = 3
    min_replicas    = 1
    cooldown_period = 60

    cpu_utilization {
      target = 0.8
    }
  }
}
