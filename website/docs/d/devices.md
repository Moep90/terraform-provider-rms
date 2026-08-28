---
page_title: "teltonika-rms-devices: Teltonika RMS Devices"
description: |-
  Retrieves a list of Teltonika RMS Devices.
---

# teltonika-rms-devices

Retrieves a list of Teltonika RMS Devices with optional filtering.

## Example Usage

```hcl
data "teltonika-rms-devices" "all" {}

data "teltonika-rms-devices" "online" {
  company_id = 12345
  status     = "online"
}

output "device_names" {
  value = data.teltonika-rms-devices.all.devices[*].name
}
```

## Argument Reference

The following arguments are optional:

- `company_id` - (Optional) Filter by company ID.
- `status` - (Optional) Filter by status: online, offline, not_activated.

## Attribute Reference

The following attributes are exported:

- `id` - The identifier for this data source.
- `devices` - A list of devices with the following attributes:
  - `id` - The device ID.
  - `name` - The device name.
  - `serial` - The device serial number.
  - `mac` - The MAC address.
  - `imei` - The IMEI number.
  - `device_series` - The device series.
  - `status` - The device status.
  - `firmware` - The firmware version.
