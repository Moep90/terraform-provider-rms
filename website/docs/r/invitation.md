---
page_title: "teltonika_rms_invitation: Teltonika RMS Invitation"
description: |-
  Manages a Teltonika RMS User Invitation.
---

# teltonika_rms_invitation

Manages a Teltonika RMS User Invitation.

## Example Usage

```hcl
resource "teltonika_rms_invitation" "new_user" {
  email      = "newuser@example.com"
  role       = "user"
  company_id = 12345
}
```

## Argument Reference

The following arguments are required:

- `email` - (Required) The email address to invite.
- `role` - (Required) The role for the invited user.
- `company_id` - (Required) The company ID to invite the user to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the invitation.
- `created_at` - The creation timestamp.

## Import

Invitations can be imported using their ID:

```bash
terraform import teltonika_rms_invitation.new_user 12345
```
