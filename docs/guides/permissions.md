---
page_title: "Permissions and Roles"
subcategory: "Guides"
---

# Teltonika RMS Permissions and Roles

How to look up RMS permissions and roles and use them to build `rms_role`.

## Data Sources

### rms_permissions

Retrieves the available RMS permissions. Each entry carries the `id` that
`rms_role.permission_ids` expects, plus the `name` used to select it.

```hcl
data "rms_permissions" "all" {}

output "permission_names" {
  value = [for p in data.rms_permissions.all.permissions : p.name]
}
```

The catalogue is returned sorted by `name`. `id` is null if the API reports no
numeric identifier for a permission, in which case that permission cannot be
assigned to a role.

### rms_roles

Retrieves the roles visible to the token's company, sorted by `id`.

```hcl
data "rms_roles" "all" {}

output "role_titles" {
  value = [for r in data.rms_roles.all.roles : r.title]
}

output "admin_role_id" {
  value = one([for r in data.rms_roles.all.roles : r.id if r.title == "Admin"])
}
```

## Usage Examples

### Creating a role with selected permissions

```hcl
data "rms_permissions" "all" {}

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

### Reading an existing role

```hcl
data "rms_roles" "existing" {}

locals {
  admin_role = one([for r in data.rms_roles.existing.roles : r if r.title == "Admin"])
}

output "admin_role" {
  value = {
    id             = local.admin_role.id
    title          = local.admin_role.title
    description    = local.admin_role.description
    permission_ids = local.admin_role.permission_ids
  }
}
```

`one()` returns null when no role matches, instead of failing on an index into
an empty list.

### Cloning a role

```hcl
data "rms_roles" "templates" {}

locals {
  viewer_role = one([for r in data.rms_roles.templates.roles : r if r.title == "Viewer"])
}

resource "rms_role" "readonly_admin" {
  title          = "Read-Only Admin"
  description    = "Admin with read-only access"
  company_id     = rms_company.main.id
  permission_ids = local.viewer_role.permission_ids
}
```
