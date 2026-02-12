resource "google_storage_bucket" "llm_models" {
  name          = "${local.project_id}-llm-models-${var.environment}"
  location      = local.region
  force_destroy = false

  uniform_bucket_level_access = true
  versioning {
    enabled = var.bucket_versioning
  }

  depends_on = [google_project_service.devops_api]
}

# Terraform State Storage Bucket
resource "google_storage_bucket" "terraform_state" {
  name          = "${local.project_id}-terraform-state"
  location      = local.region
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  # Enable object versioning for state locking
  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age        = 90
      with_state = "ARCHIVED"
    }
  }

  depends_on = [google_project_service.devops_api]
}
