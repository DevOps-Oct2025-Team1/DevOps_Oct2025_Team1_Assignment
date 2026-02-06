resource "google_storage_bucket" "llm_models" {
  name          = "${local.project_id}-llm-models-staging"
  location      = local.region
  force_destroy = false
  
  uniform_bucket_level_access = true
  versioning {
    enabled = true
  }
}
