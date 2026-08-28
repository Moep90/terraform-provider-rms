---
page_title: "rms_devices_export: Teltonika RMS Devices Export"
description: |-
  Exports device information in CSV format from Teltonika RMS.
---

# rms_devices_export

Exports device information in CSV format from Teltonika RMS. This data source retrieves all devices and returns them as raw CSV data.

## Example Usage

```hcl
data "rms_devices_export" "all" {}

output "devices_csv" {
  value = data.rms_devices_export.all.csv_data
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The identifier for this data source.
- `csv_data` - The raw CSV data containing device information with columns: id, name, serial, device_series, status, mac, imei, company_id, monitoring_enable, etc.
