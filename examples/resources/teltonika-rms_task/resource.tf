resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_task" "reboot_all" {
  name        = "Reboot All Devices"
  description = "Scheduled reboot for all devices"
  task_type   = "reboot"
  company_id  = rms_company.main.id
  payload     = "{\"command\":\"reboot\"}"
}
