output "email_config_id" {
  description = "ID of the email configuration"
  value       = rms_email_configuration.main.id
}

output "alert_config_id" {
  description = "ID of the alert configuration"
  value       = rms_alert_configuration.device_alert.id
}

output "role_id" {
  description = "ID of the role"
  value       = rms_role.admin.id
}

output "vpn_hub_id" {
  description = "ID of the VPN hub"
  value       = rms_vpn_hub.main.id
}

output "vpn_hub_route_id" {
  description = "ID of the VPN hub route"
  value       = rms_vpn_hub_route.network_a.id
}
