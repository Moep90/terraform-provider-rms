---
page_title: "rms_task_group: Teltonika RMS Task Group"
description: |-
  Manages a Teltonika RMS Task Group.
---

# rms_task_group

Manages a Teltonika RMS Task Group. Task groups organize related tasks for batch operations.

## Example Usage

```hcl
resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_task_group" "maintenance" {
  name        = "Maintenance Tasks"
  description = "Group of maintenance-related tasks"
  company_id  = rms_company.main.id
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) The name of the task group.
- `company_id` - (Required) The company ID that owns this task group.

The following arguments are optional:

- `description` - (Optional) Description of the task group.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the task group.
- `status` - The current status of the task group (`active`, `paused`, `completed`).
- `task_count` - Number of tasks in this group.
- `created_at` - Creation timestamp in ISO8601 format.
- `updated_at` - Last update timestamp in ISO8601 format.

## Import

Task groups can be imported using their ID:

```bash
terraform import rms_task_group.maintenance 12345
```
