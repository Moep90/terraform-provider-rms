---
page_title: "rms_device_esim_bootstrap: Teltonika RMS Device E-SIM Bootstrap"
description: |-
  Retrieves the E-SIM bootstrap status for a Teltonika RMS device.
---

# rms_device_esim_bootstrap

Retrieves the E-SIM bootstrap status for a Teltonika RMS device. Useful for TRB devices with eSIM capabilities.

## Example Usage

```hcl
data "rms_device_esim_bootstrap" "example" {
  device_id = 12345
}

output "esim_status" {
  value = data.rms_device_esim_bootstrap.example.esim_bootstrap
}
```

## Argument Reference

The following arguments are required:

- `device_id` - (Required) The ID of the device to check E-SIM bootstrap status for.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The identifier for this data source.
- `esim_bootstrap` - The E-SIM bootstrap status (e.g., 'enabled', 'disabled', 'pending').
- `status` - The overall status of the E-SIM bootstrap check.
- `message` - Additional message about the E-SIM bootstrap status.
