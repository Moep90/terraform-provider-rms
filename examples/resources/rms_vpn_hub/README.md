# VPN Hub Resource

Manages VPN hubs for secure multi-site connectivity in RMS.

## Example Usage

```hcl
resource "rms_vpn_hub" "main" {
  name         = "Production VPN"
  description  = "Main VPN hub for production devices"
  company_id   = 123456
  hub_zone     = "frankfurt-1"
  vpn_type     = "tun"
  tag_ids      = [10, 20]
}
```

## Argument Reference

### Required

- `name` - (String) VPN hub name
- `company_id` - (Int) Company ID this VPN hub belongs to
- `hub_zone` - (String) Server location (e.g., "frankfurt-1", "bahrain-1")

### Optional

- `description` - (String) VPN hub description
- `vpn_type` - (String) VPN type: "tap" or "tun"
- `tag_ids` - (Set of Int) List of tag IDs assigned to this VPN hub

## Import

VPN hubs can be imported using the ID:

```bash
terraform import rms_vpn_hub.main 1
```
