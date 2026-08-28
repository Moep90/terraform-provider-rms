# Teltonika RMS Permissions and Roles

This document describes how to work with RMS permissions and roles using the Terraform provider.

## Available Permissions

The RMS API provides 151 permissions across different categories. You can retrieve all available permissions using the `rms_permissions` datasource.

### Permission Categories

**Device Management:**
- `view_devices`, `create_devices`, `update_devices`, `delete_devices`
- `move_devices`
- `view_device_tasks`, `create_device_tasks`, `update_device_tasks`, `delete_device_tasks`
- `execute_device_actions`, `cancel_device_actions`
- `view_device_logs`, `view_device_alerts`

**User & Company Management:**
- `view_users`, `update_users`, `delete_users`, `invite_users`
- `view_companies`, `create_companies`, `update_companies`, `delete_companies`
- `hierarchy_access`

**Role & Permission Management:**
- `view_roles`, `create_roles`, `update_roles`, `delete_roles`
- `view_permissions`

**VPN & Network:**
- `view_vpn_hubs`, `create_vpn_hubs`, `update_vpn_hubs`, `delete_vpn_hubs`
- `view_device_vpn`, `create_device_vpn`, `update_device_vpn`, `delete_device_vpn`

**Monitoring & Alerts:**
- `view_device_alerts`, `create_device_alert_configurations`, `update_device_alert_configurations`
- `view_device_dashboard`

And many more...

## Data Sources

### rms_permissions

Retrieves all available RMS permissions.

```hcl
data "rms_permissions" "all" {}

output "available_permissions" {
  value = data.rms_permissions.all.permissions
}

output "permission_names" {
  value = [for p in data.rms_permissions.all.permissions : p.name]
}
```

### rms_roles

Retrieves all roles for your company.

```hcl
data "rms_roles" "all" {}

output "role_titles" {
  value = [for r in data.rms_roles.all.roles : r.title]
}

output "admin_role_id" {
  value = [for r in data.rms_roles.all.roles if r.title == "Admin"][0].id
}
```

## Usage Examples

### Creating a Role with Specific Permissions

```hcl
# Get all device management permissions
data "rms_permissions" "device_perms" {
}

locals {
  device_permission_ids = [
    for perm in data.rms_permissions.device_perms.permissions
    if can(regex("^view_devices|^create_devices|^update_devices|^delete_devices", perm.name))
  ]
}

resource "rms_role" "device_manager" {
  title          = "Device Manager"
  description    = "Can manage devices and tasks"
  company_id     = 123
  permission_ids = [for perm in local.device_permission_ids : perm.id]
}
```

### Using Role Data Source to Assign Permissions

```hcl
data "rms_roles" "existing" {}

# Find the admin role
locals {
  admin_role = [for role in data.rms_roles.existing.roles if role.title == "Admin"][0]
}

output "admin_role_details" {
  value = {
    id            = local.admin_role.id
    title         = local.admin_role.title
    description   = local.admin_role.description
    permission_ids = local.admin_role.permission_ids
  }
}
```

### Creating a Custom Role Based on Existing Role

```hcl
data "rms_roles" "templates" {}

locals {
  viewer_role = [for role in data.rms_roles.templates.roles if role.title == "Viewer"][0]
}

resource "rms_role" "readonly_admin" {
  title          = "Read-Only Admin"
  description    = "Admin with read-only access"
  company_id     = 123
  permission_ids = local.viewer_role.permission_ids
}
```
