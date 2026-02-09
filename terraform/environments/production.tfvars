project_id  = "dop-assignment-team1-production"
region      = "asia-southeast1"
environment = "production"

gke_zones          = ["asia-southeast1-a", "asia-southeast1-b", "asia-southeast1-c"]
gke_node_count     = 3
gke_min_node_count = 3
gke_max_node_count = 10
gke_machine_type   = "e2-standard-4"
gke_disk_size      = 100

cluster_min_cpu    = 24
cluster_max_cpu    = 100
cluster_min_memory = 96
cluster_max_memory = 400

vpc_cidr      = "10.0.0.0/16"
pods_cidr     = "10.1.0.0/16"
services_cidr = "10.2.0.0/20"
master_cidr   = "172.16.0.0/28"

ip_range_iap = "35.235.240.0/20"

enable_deletion_protection = true
bucket_versioning          = true
