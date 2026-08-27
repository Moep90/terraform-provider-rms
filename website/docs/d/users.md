---
page_title: "teltonika_rms_users: Teltonika RMS Users"
description: |-
  Retrieves a list of Teltonika RMS Users.
---

# teltonika_rms_users

Retrieves a list of all Teltonika RMS Users.

## Example Usage

```hcl
data "teltonika_rms_users" "all" {}

output "usernames" {
  value = data.teltonika_rms_users.all.users[*].username
}
```

## Attribute Reference

The following attributes are exported:

- `id` - The identifier for this data source.
- `users` - A list of users with the following attributes:
  - `id` - The user ID.
  - `username` - The username.
  - `email` - The email address.
  - `role` - The user role.
  - `company_id` - The company ID.
