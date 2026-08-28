data "teltonika-rms_invitations" "all" {}

output "invitation_emails" {
  value = data.teltonika-rms_invitations.all.invitations[*].email
}
