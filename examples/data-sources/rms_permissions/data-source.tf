# RMS exposes permissions per role, not as a global catalogue. Role 2 is the
# built-in Administrator role, which holds every permission.
data "rms_permissions" "all" {
  role_id = 2
}

output "permission_names" {
  value = [for p in data.rms_permissions.all.permissions : p.name]
}
