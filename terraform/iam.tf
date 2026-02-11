resource "google_service_account_iam_member" "workload_identity_binding" {
  service_account_id = google_service_account.devops_compute_gke_user.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${local.project_id}.svc.id.goog[default/default]"

  depends_on = [google_container_cluster.devops_gke]
}

resource "google_service_account" "devops_compute_gke_user" {
  account_id   = "devops-compute-gke-user"
  display_name = "DevOps GKE Compute Service Account"
}

resource "google_service_account" "devops_github_actions" {
  account_id   = "devops-github-actions"
  display_name = "DevOps GitHub Actions Service Account"
}

resource "google_project_iam_member" "devops_gke_user_binding" {
  project = local.project_id
  role    = "roles/container.nodeServiceAccount"
  member  = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}

resource "google_storage_bucket_iam_member" "gke_llm_models_viewer" {
  bucket = google_storage_bucket.llm_models.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${var.project_id}.svc.id.goog[default/llm-service-account]"
}

resource "google_storage_bucket_iam_member" "gke_llm_models_downloader" {
  bucket = google_storage_bucket.llm_models.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.project_id}.svc.id.goog[default/llm-downloader-service-account]"
}

resource "google_pubsub_topic_iam_member" "gke_prompt_manager" {
  topic  = google_pubsub_topic.prompt_requests.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${var.project_id}.svc.id.goog[default/prompt-manager-service-account]"
}

resource "google_pubsub_subscription_iam_member" "gke_prompt_manager" {
  subscription = google_pubsub_subscription.prompt_requests.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${var.project_id}.svc.id.goog[default/prompt-manager-service-account]"
}

resource "google_project_iam_member" "devops_gke_artifact_registry_reader" {
  project = local.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}

resource "google_project_iam_member" "devops_github_actions_artifact_registry_writer" {
  project = local.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.devops_github_actions.email}"
}

# IAM permissions for Terraform state bucket access
resource "google_storage_bucket_iam_member" "github_actions_terraform_state_admin" {
  bucket = google_storage_bucket.terraform_state.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.devops_github_actions.email}"
}

resource "google_project_iam_member" "github_actions_compute_admin" {
  project = local.project_id
  role    = "roles/compute.admin"
  member  = "serviceAccount:${google_service_account.devops_github_actions.email}"
}

resource "google_project_iam_member" "github_actions_container_admin" {
  project = local.project_id
  role    = "roles/container.admin"
  member  = "serviceAccount:${google_service_account.devops_github_actions.email}"
}

resource "google_project_iam_member" "github_actions_iam_admin" {
  project = local.project_id
  role    = "roles/resourcemanager.projectIamAdmin"
  member  = "serviceAccount:${google_service_account.devops_github_actions.email}"
}
