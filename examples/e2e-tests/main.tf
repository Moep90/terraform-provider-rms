terraform {
  required_providers {
    rms = {
      source  = "registry.example.com/teltonika-rms/rms"
      version = ">= 0.1.0"
    }
  }
}

provider "rms" {
  token    = var.rms_token
  base_url = var.rms_base_url
}

variable "rms_token" {
  description = "RMS API token"
  type        = string
  sensitive   = true
}

variable "rms_base_url" {
  description = "RMS API base URL"
  type        = string
  default     = "https://eu.rms.teltonika.lt/api"
}

# Email Configuration
resource "rms_email_configuration" "main" {
  company_id = 123456
  from_name  = "RMS Notifications"
  from_email = "notifications@example.com"
  smtp_host  = "smtp.example.com"
  smtp_port  = 587
  username   = "notifications@example.com"
  password   = "smtp_password_here"
  use_tls    = true
}

# Alert Configuration
resource "rms_alert_configuration" "device_alert" {
  device_id        = 789012345
  alert_type_id    = 1
  alert_subtype_id = 2
  action           = 1
  subject          = "Device Alert"
  message          = "Alert from device"
  email            = "admin@example.com"
}

# Role
resource "rms_role" "admin" {
  title          = "Administrator"
  description    = "Full access role"
  company_id     = 123456
  permission_ids = [1, 2, 3, 4, 5]
}

# VPN Hub
resource "rms_vpn_hub" "main" {
  name        = "Production VPN"
  description = "Main VPN hub for production devices"
  company_id  = 123456
  hub_zone    = "frankfurt-1"
  vpn_type    = "tun"
  tag_ids     = [10, 20]
}

# VPN Hub Route
resource "rms_vpn_hub_route" "network_a" {
  vpn_hub_id      = rms_vpn_hub.main.id
  vpn_hub_user_id = 1001
  ip_address      = "192.168.1.0"
  netmask         = "255.255.255.0"
  description     = "Network A subnet"
}

output "email_config_id" {
  value = rms_email_configuration.main.id
}

output "alert_config_id" {
  value = rms_alert_configuration.device_alert.id
}

output "role_id" {
  value = rms_role.admin.id
}

output "vpn_hub_id" {
  value = rms_vpn_hub.main.id
}

output "vpn_hub_route_id" {
  value = rms_vpn_hub_route.network_a.id
}
