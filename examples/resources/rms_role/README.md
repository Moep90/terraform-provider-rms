# Role Resource

Manages user roles with permissions in RMS.

## Example Usage

```hcl
resource "rms_role" "admin" {
  title          = "Administrator"
  description    = "Full access role"
  company_id     = 123456
  permission_ids = [1, 2, 3, 4, 5]
}

resource "rms_role" "viewer" {
  title          = "Viewer"
  description    = "Read-only access"
  company_id     = 123456
  permission_ids = [1]
}
```

## Argument Reference

### Required

- `title` - (String) Role name/title
- `company_id` - (Int) Company ID this role belongs to
- `permission_ids` - (Set of Int) List of permission IDs assigned to this role

### Optional

- `description` - (String) Role description

## Import

Roles can be imported using the ID:

```bash
terraform import rms_role.admin 1
```
