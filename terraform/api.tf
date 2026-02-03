resource "google_project_service" "devops-api" {
  for_each = toset(local.apis)
  service  = each.key

  disable_on_destroy = false
}
