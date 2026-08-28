resource "rms_email_configuration" "main" {
  company_id = 123456
  from_name  = "RMS Notifications"
  from_email = "notifications@example.com"
  smtp_host  = "smtp.example.com"
  smtp_port  = 587
  username   = "notifications@example.com"
  password   = "smtp_password_here"
  use_tls    = true
}
