# Staging environment overrides — apply with:
# terraform apply -var-file=staging.tfvars
environment             = "staging"
db_instance_tier        = "db-f1-micro"
redis_memory_size_gb    = 1
core_min_instances      = 0
notification_min_instances = 0
module_min_instances    = 0
frontend_min_instances  = 0
docker_image_tag        = "latest"
