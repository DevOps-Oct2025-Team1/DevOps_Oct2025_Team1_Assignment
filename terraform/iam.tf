resource "google_service_account_iam_member" "workload_identity_binding" {
  service_account_id = google_service_account.devops_compute_gke_user.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${local.project_id}.svc.id.goog[default/default]"

  depends_on = [google_container_cluster.devops-gke]
}

resource "google_service_account" "devops_compute_gke_user" {
  account_id   = "devops-compute-gke-user"
  display_name = "DevOps GKE Compute Service Account"
}

resource "google_project_iam_member" "devops_gke_user_binding" {
  project = local.project_id
  role    = "roles/container.nodeServiceAccount"
  member  = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}

resource "google_storage_bucket_iam_member" "gke_bucket_access" {
  bucket = google_storage_bucket.llm_models.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}

resource "google_project_iam_member" "devops_gke_artifact_registry_writer" {
  project = local.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}

