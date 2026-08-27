---
page_title: "teltonika_rms_tag: Teltonika RMS Tag"
description: |-
  Manages a Teltonika RMS Tag.
---

# teltonika_rms_tag

Manages a Teltonika RMS Tag. Tags can be used to organize and filter devices.

## Example Usage

```hcl
resource "teltonika_rms_tag" "production" {
  name       = "Production"
  color      = "#00ff00"
  company_id = 12345
}

resource "teltonika_rms_tag" "development" {
  name       = "Development"
  color      = "#ff0000"
  company_id = 12345
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) The name of the tag.
- `company_id` - (Required) The company ID to assign the tag to.

The following arguments are optional:

- `color` - (Optional) The color of the tag (hex code).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the tag.
- `device_count` - The number of devices assigned to this tag.

## Import

Tags can be imported using their ID:

```bash
terraform import teltonika_rms_tag.production 12345
```
