---
page_title: "teltonika-rms-tags: Teltonika RMS Tags"
description: |-
  Retrieves a list of Teltonika RMS Tags.
---

# teltonika-rms-tags

Retrieves a list of all Teltonika RMS Tags.

## Example Usage

```hcl
data "teltonika-rms-tags" "all" {}

output "tag_names" {
  value = data.teltonika-rms-tags.all.tags[*].name
}
```

## Attribute Reference

The following attributes are exported:

- `id` - The identifier for this data source.
- `tags` - A list of tags with the following attributes:
  - `id` - The tag ID.
  - `name` - The tag name.
  - `color` - The tag color (hex code).
  - `company_id` - The company ID.
  - `device_count` - The number of devices.
