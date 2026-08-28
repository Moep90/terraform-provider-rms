data "rms_invitations" "all" {}

output "invitation_emails" {
  value = data.rms_invitations.all.invitations[*].email
}
