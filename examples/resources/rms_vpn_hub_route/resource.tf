resource "rms_vpn_hub" "main" {
  name         = "Production VPN"
  company_id   = 123456
  hub_zone     = "frankfurt-1"
}

resource "rms_vpn_hub_route" "network_a" {
  vpn_hub_id      = rms_vpn_hub.main.id
  vpn_hub_user_id = 1001
  ip_address      = "192.168.1.0"
  netmask         = "255.255.255.0"
  description     = "Network A subnet"
}
