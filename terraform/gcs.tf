resource "google_storage_bucket" "llm_models" {
  name          = "${local.project_id}-llm-models"
  location      = local.region
  force_destroy = false
  
  uniform_bucket_level_access = true
  versioning {
    enabled = true
  }
}

resource "google_storage_bucket_iam_member" "gke_bucket_access" {
  bucket = google_storage_bucket.llm_models.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.devops_compute_gke_user.email}"
}
