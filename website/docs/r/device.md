---
page_title: "teltonika_rms_device: Teltonika RMS Device"
description: |-
  Manages a Teltonika RMS Device.
---

# teltonika_rms_device

Manages a Teltonika RMS Device. Supports RUT, TRB, and other device series.

## Example Usage

```hcl
resource "teltonika_rms_company" "main" {
  company_name = "My Company"
}

resource "teltonika_rms_device" "router" {
  name               = "Office Router"
  device_series      = "rut"
  serial             = "0123456789"
  mac                = "00:11:22:33:44:55"
  company_id         = teltonika_rms_company.main.id
  auto_credit_enable = true
  password           = "device-password"
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) The name of the device.
- `device_series` - (Required) The device series: rut, trb, etc.
- `serial` - (Required) The device serial number.
- `company_id` - (Required) The company ID to assign the device to.

The following arguments are optional:

- `mac` - (Optional) The MAC address (required for RUT devices).
- `imei` - (Optional) The IMEI (required for TRB devices).
- `auto_credit_enable` - (Optional) Whether to automatically enable credits for the device. Defaults to `true`.
- `password` - (Optional) The device password for initial validation.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the device.
- `status` - The device status: online, offline, not_activated.
- `firmware` - The device firmware version.
- `created_at` - The timestamp when the device was added.

## Import

Devices can be imported using their ID:

```bash
terraform import teltonika_rms_device.router 12345
```
