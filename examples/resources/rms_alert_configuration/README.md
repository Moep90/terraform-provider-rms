# Alert Configuration Resource

Manages alert rules for device monitoring in RMS.

## Example Usage

```hcl
resource "rms_alert_configuration" "device_alert" {
  device_id        = 789012345
  alert_type_id    = 1
  alert_subtype_id = 2
  action           = 1
  subject          = "Device Alert"
  message          = "Alert from device"
  email            = "admin@example.com"
}
```

## Argument Reference

### Required

- `device_id` - (Int) The device ID to monitor
- `alert_type_id` - (Int) The alert type identifier

### Optional

- `alert_subtype_id` - (Int) The alert subtype identifier
- `action` - (Int) Action on alert trigger (1 = send email, etc.)
- `subject` - (String) Email subject line
- `message` - (String) Email body content
- `email` - (String) Recipient email address
- `smtp_config_id` - (Int) SMTP configuration to use

## Import

Alert configurations can be imported using the ID:

```bash
terraform import rms_alert_configuration.device_alert 1
```
