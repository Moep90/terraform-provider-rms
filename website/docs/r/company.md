---
page_title: "teltonika_rms_company: Teltonika RMS Company"
description: |-
  Manages a Teltonika RMS Company.
---

# teltonika_rms_company

Manages a Teltonika RMS Company. Companies can be hierarchical with parent-child relationships.

## Example Usage

```hcl
resource "teltonika_rms_company" "main" {
  company_name = "Main Company"
}

resource "teltonika_rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = teltonika_rms_company.main.id
}
```

## Argument Reference

The following arguments are required:

- `company_name` - (Required) The name of the company.

The following arguments are optional:

- `parent_id` - (Optional) The parent company ID. If set, this company becomes a subsidiary of the parent.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the company.
- `device_count` - The number of devices associated with this company.
- `created_at` - The timestamp when the company was created.

## Import

Companies can be imported using their ID:

```bash
terraform import teltonika_rms_company.main 12345
```
