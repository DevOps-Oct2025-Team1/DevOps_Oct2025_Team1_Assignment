project_id  = "dop-assignment-team1-prod"
project_number = "106380445880"
region      = "us-central1"
environment = "production"

gke_zones          = ["us-central1-a"]
gke_node_count     = 1
gke_min_node_count = 1
gke_max_node_count = 3
gke_machine_type   = "e2-standard-4"
gke_disk_size      = 30

cluster_min_cpu    = 8
cluster_max_cpu    = 12
cluster_min_memory = 32
cluster_max_memory = 48

subnet_pods_range     = "pods"
subnet_services_range = "services"
vpc_cidr      = "10.20.0.0/16"
pods_cidr     = "10.21.0.0/16"
services_cidr = "10.22.0.0/20"
master_cidr   = "172.16.0.0/28"

ip_range_iap = "35.235.240.0/20"

enable_deletion_protection = true
bucket_versioning          = true
