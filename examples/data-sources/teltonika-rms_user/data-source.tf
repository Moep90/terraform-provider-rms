data "rms_users" "all" {}

output "usernames" {
  value = data.rms_users.all.users[*].username
}
