---
page_title: "teltonika-rms-users: Teltonika RMS Users"
description: |-
  Retrieves a list of Teltonika RMS Users.
---

# teltonika-rms-users

Retrieves a list of all Teltonika RMS Users.

## Example Usage

```hcl
data "teltonika-rms-users" "all" {}

output "usernames" {
  value = data.teltonika-rms-users.all.users[*].username
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
