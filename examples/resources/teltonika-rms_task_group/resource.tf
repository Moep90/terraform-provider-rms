resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_task_group" "maintenance" {
  name        = "Maintenance Tasks"
  description = "Group of maintenance-related tasks"
  company_id  = rms_company.main.id
}
