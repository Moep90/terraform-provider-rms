---
page_title: "rms_task: Teltonika RMS Task"
description: |-
  Manages a Teltonika RMS Task.
---

# rms_task

Manages a Teltonika RMS Task. Tasks are used to send commands or configurations to devices.

## Example Usage

```hcl
resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_task" "reboot_all" {
  name        = "Reboot All Devices"
  description = "Scheduled reboot for all devices"
  task_type   = "reboot"
  company_id  = rms_company.main.id
  payload     = "{\"command\":\"reboot\"}"
}
```

### Sequential Task Group Example

```hcl
resource "rms_company" "main" {
  company_name = "My Company"
}

resource "rms_task_group" "maintenance" {
  name        = "Maintenance Tasks"
  description = "Group of maintenance-related tasks"
  company_id  = rms_company.main.id
}

resource "rms_task" "reboot_first" {
  name          = "Reboot First Device"
  description   = "Sequential reboot task"
  task_type     = "reboot"
  company_id    = rms_company.main.id
  task_group_id = rms_task_group.maintenance.id
  payload       = "{\"command\":\"reboot\",\"sequential\":true}"
  scheduled_at  = "2024-06-01T00:00:00Z"
}

resource "rms_task" "config_update" {
  name          = "Config Update"
  description   = "Sequential config update task"
  task_type     = "config_update"
  company_id    = rms_company.main.id
  task_group_id = rms_task_group.maintenance.id
  payload       = "{\"config\":\"updated\"}"
  scheduled_at  = "2024-06-02T00:00:00Z"
}

resource "rms_task" "firmware_upgrade" {
  name          = "Firmware Upgrade"
  description   = "Sequential firmware upgrade task"
  task_type     = "firmware_upgrade"
  company_id    = rms_company.main.id
  task_group_id = rms_task_group.maintenance.id
  payload       = "{\"firmware\":\"v2.0\"}"
  scheduled_at  = "2024-06-03T00:00:00Z"
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) The name of the task.
- `task_type` - (Required) The type of task (e.g., `reboot`, `config_update`, `firmware_upgrade`).
- `company_id` - (Required) The company ID that owns this task.

The following arguments are optional:

- `description` - (Optional) Description of the task.
- `task_group_id` - (Optional) The task group ID this task belongs to.
- `payload` - (Optional) JSON payload containing task-specific parameters.
- `scheduled_at` - (Optional) Scheduled execution time in ISO8601 format.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier for the task.
- `status` - The current status of the task (`pending`, `running`, `completed`, `failed`).
- `created_at` - Creation timestamp in ISO8601 format.
- `updated_at` - Last update timestamp in ISO8601 format.

## Import

Tasks can be imported using their ID:

```bash
terraform import rms_task.reboot_all 12345
```
