---
page_title: "teltonika-rms-user: Teltonika RMS User"
description: |-
  Manages a Teltonika RMS User.
---

# teltonika-rms-user

Manages a Teltonika RMS User.

## Example Usage

```hcl
resource "teltonika-rms-user" "admin" {
  username   = "admin"
  email      = "admin@example.com"
  role       = "admin"
  company_id = 12345
}
```

## Argument Reference

The following arguments are required:

- `username` - (Required) The username.
- `email` - (Required) The email address.
- `role` - (Required) The user role.
- `company_id` - (Required) The company ID.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the user.

## Import

Users can be imported using their ID:

```bash
terraform import teltonika-rms-user.admin 12345
```
