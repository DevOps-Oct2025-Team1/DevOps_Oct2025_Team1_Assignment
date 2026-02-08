variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region for resources"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production, etc.)"
  type        = string
}

variable "gke_zones" {
  description = "List of zones for GKE nodes"
  type        = list(string)
}

variable "gke_node_count" {
  description = "Initial node count for GKE node pool"
  type        = number
}

variable "gke_min_node_count" {
  description = "Minimum node count for GKE autoscaling"
  type        = number
}

variable "gke_max_node_count" {
  description = "Maximum node count for GKE autoscaling"
  type        = number
}

variable "gke_machine_type" {
  description = "Machine type for GKE nodes"
  type        = string
}

variable "gke_disk_size" {
  description = "Disk size in GB for GKE nodes"
  type        = number
}

variable "cluster_min_cpu" {
  description = "Minimum CPU for cluster autoscaling"
  type        = number
}

variable "cluster_max_cpu" {
  description = "Maximum CPU for cluster autoscaling"
  type        = number
}

variable "cluster_min_memory" {
  description = "Minimum memory for cluster autoscaling"
  type        = number
}

variable "cluster_max_memory" {
  description = "Maximum memory for cluster autoscaling"
  type        = number
}

variable "vpc_cidr" {
  description = "CIDR range for the VPC subnet"
  type        = string
}

variable "pods_cidr" {
  description = "CIDR range for GKE pods"
  type        = string
}

variable "services_cidr" {
  description = "CIDR range for GKE services"
  type        = string
}

variable "master_cidr" {
  description = "CIDR range for GKE master"
  type        = string
}

variable "ip_range_iap" {
  description = "CIDR range for IAP (Identity-Aware Proxy)"
  type        = string
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection for GKE cluster"
  type        = bool
}

variable "bucket_versioning" {
  description = "Enable versioning for GCS buckets"
  type        = bool
}
