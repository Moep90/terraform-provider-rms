data "rms_roles" "all" {}

output "role_titles" {
  value = [for r in data.rms_roles.all.roles : r.title]
}
