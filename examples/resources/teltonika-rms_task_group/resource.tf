resource "teltonika-rms_company" "main" {
  company_name = "My Company"
}

resource "teltonika-rms_task_group" "maintenance" {
  name        = "Maintenance Tasks"
  description = "Group of maintenance-related tasks"
  company_id  = teltonika-rms_company.main.id
}
