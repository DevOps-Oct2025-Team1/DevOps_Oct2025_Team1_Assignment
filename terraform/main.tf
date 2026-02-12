# Cloud Storage Bucket for Application Artifacts
resource "google_artifact_registry_repository" "docker_repo" {
  provider = google-beta

  location      = var.region
  repository_id = "${var.environment}-docker-repo"
  description   = "Docker repository for ${var.environment}"
  format        = "DOCKER"

  labels     = local.common_labels
  depends_on = [google_project_service.devops_api]
}

# Pub/Sub Topic and Subscription
resource "google_pubsub_topic" "prompt_requests" {
  name = "${var.environment}-prompt-requests"

  labels = local.common_labels

  depends_on = [google_project_service.devops_api]
}

resource "google_pubsub_subscription" "prompt_requests" {
  name  = "${var.environment}-prompt-requests-sub"
  topic = google_pubsub_topic.prompt_requests.name

  labels = local.common_labels
}
