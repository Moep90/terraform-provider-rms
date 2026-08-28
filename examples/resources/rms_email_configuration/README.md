# Email Configuration Resource

Manages SMTP email configuration for RMS notifications.

## Example Usage

```hcl
resource "rms_email_configuration" "main" {
  company_id  = 123456
  from_name   = "RMS Notifications"
  from_email  = "notifications@example.com"
  smtp_host   = "smtp.example.com"
  smtp_port   = 587
  username    = "notifications@example.com"
  password    = "smtp_password_here"
  use_tls     = true
}
```

## Argument Reference

### Required

- `company_id` - (Int) The company ID
- `from_name` - (String) Sender display name
- `from_email` - (String) Sender email address
- `smtp_host` - (String) SMTP server hostname
- `smtp_port` - (Int) SMTP port number
- `username` - (String) SMTP authentication username
- `password` - (String) SMTP authentication password

### Optional

- `use_tls` - (Bool) Enable TLS encryption (default: true)

## Import

Email configurations can be imported using the ID:

```bash
terraform import rms_email_configuration.main 1
```
