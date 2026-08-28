---
page_title: "Permissions and Roles"
subcategory: "Guides"
---

# Teltonika RMS Permissions and Roles

How to look up RMS permissions and roles and use them to build `rms_role`.

RMS has no global permissions endpoint. Permissions are reachable only through
a role, at `/roles/{id}/permissions`, so `rms_permissions` takes a `role_id`.
The built-in Administrator role (id `2`) holds every permission, which makes it
the practical choice for discovering IDs.

## Data Sources

### rms_permissions

```hcl
data "rms_permissions" "all" {
  role_id = 2
}

output "permission_names" {
  value = [for p in data.rms_permissions.all.permissions : p.name]
}
```

Each permission exposes `id`, `name` (a slug such as `view_pending_device_actions`),
`title`, `description` and `category`. The list is sorted by `name`.

### rms_roles

Retrieves the roles visible to the token, sorted by `id`. `permission_ids` is
populated per role from the sub-resource, which costs one extra request per
role.

```hcl
data "rms_roles" "all" {}

output "role_titles" {
  value = [for r in data.rms_roles.all.roles : r.title]
}
```

Roles carry both a `name` slug and a human readable `title`. `company_id` is
null on the built-in roles, which belong to no company.

## Usage Examples

### Creating a role with selected permissions

```hcl
data "rms_permissions" "all" {
  role_id = 2
}

locals {
  device_permissions = [
    "view_devices",
    "create_devices",
    "update_devices",
    "delete_devices",
  ]

  device_permission_ids = [
    for p in data.rms_permissions.all.permissions : p.id
    if contains(local.device_permissions, p.name)
  ]
}

resource "rms_role" "device_manager" {
  title          = "Device Manager"
  description    = "Can manage devices"
  company_id     = rms_company.main.id
  permission_ids = local.device_permission_ids
}
```

Selecting by name rather than by hardcoded id keeps the config readable and
survives id changes between tenants.

### Cloning a role

```hcl
data "rms_roles" "templates" {}

locals {
  viewer_role = one([for r in data.rms_roles.templates.roles : r if r.title == "Advanced guest"])
}

resource "rms_role" "readonly_admin" {
  title          = "Read-Only Admin"
  description    = "Admin with read-only access"
  company_id     = rms_company.main.id
  permission_ids = local.viewer_role.permission_ids
}
```

`one()` returns null when no role matches, instead of failing on an index into
an empty list.
