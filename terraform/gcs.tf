resource "google_storage_bucket" "llm_models" {
  name          = "${local.project_id}-llm-models-${var.environment}"
  location      = local.region
  force_destroy = false

  uniform_bucket_level_access = true
  versioning {
    enabled = var.bucket_versioning
  }
  
  depends_on = [ google_project_service.devops_api ]
}
