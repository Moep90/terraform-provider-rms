resource "rms_alert_configuration" "main" {
  device_id        = 789012345
  alert_type_id    = 1
  alert_subtype_id = 2
  action           = 1
  subject          = "Device Alert"
  message          = "Alert from device"
  email            = "admin@example.com"
}
