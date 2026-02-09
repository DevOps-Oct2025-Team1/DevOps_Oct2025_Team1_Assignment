locals {
  project_id = var.project_id
  region     = var.region
  gke_zones  = var.gke_zones

  apis = [
    "compute.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com",
    "serviceusage.googleapis.com",
  ]

  terraform_admin_roles = [
    "roles/compute.networkAdmin",
    "roles/container.admin",
    "roles/iam.serviceAccountAdmin",
    "roles/iam.serviceAccountUser",
    "roles/resourcemanager.projectIamAdmin",
    "roles/serviceusage.serviceUsageAdmin",
  ]

  common_labels = {
    "managed-by"  = "terraform"
    "project"     = "devops"
    "environment" = var.environment
  }

  ip = var.ip_range_iap
}
