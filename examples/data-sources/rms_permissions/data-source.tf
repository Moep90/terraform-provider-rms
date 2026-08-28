data "rms_permissions" "all" {}

output "permission_names" {
  value = [for p in data.rms_permissions.all.permissions : p.name]
}
