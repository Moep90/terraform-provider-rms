# VPN Hub Route Resource

Manages routing rules for VPN hubs in RMS.

## Example Usage

```hcl
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
```

## Argument Reference

### Required

- `vpn_hub_id` - (Int) Parent VPN hub ID
- `vpn_hub_user_id` - (Int) VPN user ID
- `ip_address` - (String) Route IP address
- `netmask` - (String) Subnet mask

### Optional

- `description` - (String) Route description

## Import

VPN hub routes can be imported using the composite ID (format: `vpn_hub_id:vpn_hub_user_id:ip_address`):

```bash
terraform import rms_vpn_hub_route.network_a 1:1001:192.168.1.0
```
