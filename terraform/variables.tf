locals {
  project_id = "dop-assignment-team1-staging"
  region     = "asia-southeast1"
  gke_zones = [
    "${local.region}-a",
  ]
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
    "managed-by" = "terraform"
    "project"    = "devops"
  }
}
