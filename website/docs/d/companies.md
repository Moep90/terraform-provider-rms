---
page_title: "teltonika_rms_companies: Teltonika RMS Companies"
description: |-
  Retrieves a list of Teltonika RMS Companies.
---

# teltonika_rms_companies

Retrieves a list of all Teltonika RMS Companies.

## Example Usage

```hcl
data "teltonika_rms_companies" "all" {}

output "company_names" {
  value = data.teltonika_rms_companies.all.companies[*].company_name
}
```

## Attribute Reference

The following attributes are exported:

- `id` - The identifier for this data source.
- `companies` - A list of companies with the following attributes:
  - `id` - The company ID.
  - `company_name` - The company name.
  - `parent_id` - The parent company ID.
  - `device_count` - The number of devices.
