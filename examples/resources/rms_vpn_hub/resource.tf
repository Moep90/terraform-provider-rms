resource "rms_vpn_hub" "main" {
  name        = "Production VPN"
  description = "Main VPN hub for production devices"
  company_id  = 123456
  hub_zone    = "frankfurt-1"
  vpn_type    = "tun"
  tag_ids     = [10, 20]
}
