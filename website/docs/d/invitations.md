---
page_title: "teltonika_rms_invitations: Teltonika RMS Invitations"
description: |-
  Retrieves a list of Teltonika RMS Invitations.
---

# teltonika_rms_invitations

Retrieves a list of all Teltonika RMS User Invitations.

## Example Usage

```hcl
data "teltonika_rms_invitations" "all" {}

output "invitation_emails" {
  value = data.teltonika_rms_invitations.all.invitations[*].email
}
```

## Attribute Reference

The following attributes are exported:

- `id` - The identifier for this data source.
- `invitations` - A list of invitations with the following attributes:
  - `id` - The invitation ID.
  - `email` - The email address.
  - `role` - The role.
  - `company_id` - The company ID.
  - `created_at` - The creation timestamp.
