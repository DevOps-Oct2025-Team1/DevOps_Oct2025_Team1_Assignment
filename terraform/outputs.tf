output "cluster_name" {
  description = "The name of the GKE cluster"
  value       = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  description = "The endpoint of the GKE cluster"
  value       = google_container_cluster.primary.endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "The CA certificate of the GKE cluster"
  value       = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "vpc_id" {
  description = "The ID of the VPC network"
  value       = google_compute_network.vpc.id
}

output "subnet_id" {
  description = "The ID of the subnet"
  value       = google_compute_subnetwork.subnet.id
}

output "auth_database_instance" {
  description = "The connection name of the auth Cloud SQL instance"
  value       = google_sql_database_instance.auth_postgres.connection_name
}

output "auth_database_name" {
  description = "The name of the auth database"
  value       = google_sql_database.auth_db.name
}

output "chats_database_instance" {
  description = "The connection name of the chats Cloud SQL instance"
  value       = google_sql_database_instance.chats_postgres.connection_name
}

output "chats_database_name" {
  description = "The name of the chats database"
  value       = google_sql_database.chats_db.name
}

output "artifact_repository_id" {
  description = "The ID of the Artifact Registry repository"
  value       = google_artifact_registry_repository.docker_repo.repository_id
}

output "pubsub_topic" {
  description = "The name of the Pub/Sub topic"
  value       = google_pubsub_topic.prompt_requests.name
}

output "pubsub_subscription" {
  description = "The name of the Pub/Sub subscription"
  value       = google_pubsub_subscription.prompt_requests.name
}

output "region" {
  description = "The GCP region"
  value       = var.region
}

output "project_id" {
  description = "The GCP project ID"
  value       = local.project_id
}
