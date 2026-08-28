# E2E Tests for Teltonika RMS Terraform Provider

This directory contains end-to-end examples demonstrating the full lifecycle of Teltonika RMS resources using Terraform.

## Prerequisites

1. Terraform >= 1.0
2. Valid RMS API token
3. Network access to RMS API endpoint

## Quick Start

### 1. Initialize Terraform

```bash
cd examples/e2e-tests
terraform init
```

### 2. Set Environment Variables

```bash
export RMS_TOKEN="your_api_token_here"
export RMS_BASE_URL="https://eu.rms.teltonika.lt/api"
```

Or create a `terraform.tfvars` file:

```hcl
rms_token    = "your_api_token_here"
rms_base_url = "https://eu.rms.teltonika.lt/api"
```

### 3. Plan

```bash
terraform plan -out=tfplan
```

Expected output shows resources to be created:
- `rms_email_configuration.main`
- `rms_alert_configuration.device_alert`
- `rms_role.admin`
- `rms_vpn_hub.main`
- `rms_vpn_hub_route.network_a`

### 4. Apply

```bash
terraform apply tfplan
```

Terraform will create all resources and display outputs:
- Email configuration ID
- Alert configuration ID
- Role ID
- VPN hub ID
- VPN hub route ID

### 5. Verify Resources

Check resources in RMS dashboard or via API:

```bash
# List email configurations
curl -H "Authorization: Bearer $RMS_TOKEN" \
  "$RMS_BASE_URL/email-configurations"

# List roles
curl -H "Authorization: Bearer $RMS_TOKEN" \
  "$RMS_BASE_URL/roles"
```

### 6. Update Resources (Optional)

Edit `main.tf` to modify resource attributes, then:

```bash
terraform plan
terraform apply
```

Example update - change role description:

```hcl
resource "rms_role" "admin" {
  title          = "Administrator"
  description    = "Updated admin description"  # Changed
  company_id     = 123456
  permission_ids = [1, 2, 3, 4, 5]
}
```

### 7. Destroy

```bash
terraform destroy
```

Terraform will delete all managed resources in reverse dependency order.

## Resource Details

### Email Configuration

Manages SMTP settings for email notifications.

**Required:**
- `company_id` - Company ID
- `from_name` - Sender name
- `from_email` - Sender email
- `smtp_host` - SMTP server hostname
- `smtp_port` - SMTP port (usually 587)
- `username` - SMTP username
- `password` - SMTP password

**Optional:**
- `use_tls` - Enable TLS (default: true)

### Alert Configuration

Manages alert rules for device monitoring.

**Required:**
- `device_id` - Device ID to monitor
- `alert_type_id` - Alert type identifier

**Optional:**
- `alert_subtype_id` - Alert subtype
- `action` - Action on alert (1 = email, etc.)
- `subject` - Email subject
- `message` - Email body
- `email` - Recipient email
- `smtp_config_id` - SMTP config to use

### Role

Manages user roles with permissions.

**Required:**
- `title` - Role name
- `company_id` - Company ID
- `permission_ids` - List of permission IDs

**Optional:**
- `description` - Role description

### VPN Hub

Manages VPN hub for secure connectivity.

**Required:**
- `name` - VPN hub name
- `company_id` - Company ID
- `hub_zone` - Server location (e.g., "frankfurt-1", "bahrain-1")

**Optional:**
- `description` - VPN hub description
- `vpn_type` - VPN type ("tap" or "tun")
- `tag_ids` - List of tag IDs

### VPN Hub Route

Manages routing rules for VPN hubs.

**Required:**
- `vpn_hub_id` - Parent VPN hub ID
- `vpn_hub_user_id` - VPN user ID
- `ip_address` - Route IP address
- `netmask` - Subnet mask

**Optional:**
- `description` - Route description

## Troubleshooting

### Authentication Errors

Ensure your API token is valid and has required permissions:
- `email-configurations` - Manage email configs
- `alerts-configurations` - Manage alerts
- `roles` - Manage roles
- `vpn` - Manage VPN hubs/routes

### Permission Denied

Check that the token's role includes necessary permissions. View available permissions:

```bash
curl -H "Authorization: Bearer $RMS_TOKEN" \
  "$RMS_BASE_URL/permissions"
```

### Network Issues

Verify connectivity to RMS API:

```bash
curl -I https://eu.rms.teltonika.lt/api/status
```

## Cleanup

To remove all resources and state:

```bash
terraform destroy -auto-approve
rm -rf .terraform* terraform.tfstate*
```
