resource "rms_company" "main" {
  company_name = "My Company"
}

# Task group to bundle related tasks
resource "rms_task_group" "deployment" {
  name        = "Production Deployment"
  description = "Sequential deployment tasks for production environment"
  company_id  = rms_company.main.id
}

# Task 1: Pre-deployment checks
resource "rms_task" "pre_check" {
  name        = "Pre-Deployment Health Check"
  description = "Verify all devices are online before deployment"
  task_type   = "health_check"
  company_id  = rms_company.main.id
  task_group_id = rms_task_group.deployment.id
  payload     = jsonencode({
    check_type = "connectivity"
    timeout    = 300
  })
  scheduled_at = "2024-01-15T08:00:00Z"
}

# Task 2: Configuration update (runs after pre-check)
resource "rms_task" "config_update" {
  name        = "Configuration Update"
  description = "Push new configuration to all devices"
  task_type   = "config_update"
  company_id  = rms_company.main.id
  task_group_id = rms_task_group.deployment.id
  payload     = jsonencode({
    config_file = "firmware_v2.3.tar.gz"
    apply_mode  = "staged"
    staged_percentage = 25
  })
  scheduled_at = "2024-01-15T09:00:00Z"
}

# Task 3: Post-deployment verification
resource "rms_task" "post_check" {
  name        = "Post-Deployment Verification"
  description = "Verify devices are running correctly after update"
  task_type   = "health_check"
  company_id  = rms_company.main.id
  task_group_id = rms_task_group.deployment.id
  payload     = jsonencode({
    check_type = "firmware_version"
    expected_version = "v2.3"
  })
  scheduled_at = "2024-01-15T10:00:00Z"
}
