data "teltonika-rms_users" "all" {}

output "usernames" {
  value = data.teltonika-rms_users.all.users[*].username
}
