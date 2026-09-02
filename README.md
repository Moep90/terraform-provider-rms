# Terraform Provider for Teltonika RMS

[![CI](https://github.com/Moep90/terraform-provider-rms/actions/workflows/ci.yml/badge.svg)](https://github.com/Moep90/terraform-provider-rms/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Moep90/terraform-provider-rms)](https://goreportcard.com/report/github.com/Moep90/terraform-provider-rms)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Disclaimer**: This is a community-maintained Terraform/OpenTofu provider for Teltonika RMS. It is not officially supported by Teltonika Networks.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5.0
- [OpenTofu](https://opentofu.org/) >= 1.6.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Description

This provider allows you to manage Teltonika RMS resources using Terraform or OpenTofu. It provides resources for:

- Companies (with hierarchical support)
- Devices (RUT, TRB, and other device series)
- Tags (for organizing devices)
- Users
- User Invitations
- Tasks (commands and configurations for devices)
- Task Groups (organize related tasks for batch operations)

## Installation

### Using Terraform

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    rms = {
      source  = "moep90/rms"
      version = ">= 0.1.0"
    }
  }
}
```

### Using OpenTofu

Add the provider to your OpenTofu configuration:

```hcl
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    rms = {
      source  = "moep90/rms"
      version = ">= 0.1.0"
    }
  }
}
```

## Configuration

Configure the provider with your Teltonika RMS API token:

```hcl
provider "rms" {
  token = var.teltonika_token
  # Optional: base_url defaults to https://rms.teltonika-networks.com/api
  # base_url = "https://rms.teltonika-networks.com/api"
  # Optional: timeout in seconds (default: 30)
  # timeout = 30
  # Optional: max_retry (default: 3)
  # max_retry = 3
}
```

Or use environment variables:

```bash
export TELTONIKA_RMS_TOKEN="your-api-token"
```

## Example Usage

### Managing Companies

```hcl
resource "rms_company" "main" {
  company_name = "My Company"
  parent_id    = null
}

resource "rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = rms_company.main.id
}
```

### Managing Devices

```hcl
resource "rms_device" "router" {
  name               = "Office Router"
  device_series      = "rut"
  serial             = "0123456789"
  mac                = "00:11:22:33:44:55"
  company_id         = rms_company.main.id
  auto_credit_enable = true
  password           = "device-password"
}
```

### Managing Tags

```hcl
resource "rms_tag" "production" {
  name       = "Production"
  color      = "#00ff00"
  company_id = rms_company.main.id
}

resource "rms_tag" "development" {
  name       = "Development"
  color      = "#ff0000"
  company_id = rms_company.main.id
}
```

### Using Data Sources

```hcl
data "rms_companies" "all" {}

data "rms_devices" "office" {
  company_id = rms_company.main.id
  status     = "online"
}

data "rms_tags" "all" {}
```

## Resources

- `rms_company` - Manages a Teltonika RMS Company
- `rms_device` - Manages a Teltonika RMS Device
- `rms_tag` - Manages a Teltonika RMS Tag
- `rms_user` - Manages a Teltonika RMS User
- `rms_invitation` - Manages a Teltonika RMS User Invitation
- `rms_task` - Manages a Teltonika RMS Task (commands/configurations for devices)
- `rms_task_group` - Manages a Teltonika RMS Task Group (organizes related tasks)
- `rms_role` - Manages an access control role
- `rms_device_tags` - Manages the tag assignments of a device
- `rms_alert_configuration` - Manages a device alert configuration
- `rms_email_configuration` - Manages an SMTP email configuration
- `rms_vpn_hub` - Manages a VPN hub
- `rms_vpn_hub_route` - Manages a VPN hub route

## Data Sources

- `rms_companies` - Retrieves all companies
- `rms_company` - Retrieves a single company
- `rms_devices` - Retrieves all devices with optional filtering
- `rms_device` - Retrieves a single device
- `rms_tags` - Retrieves all tags
- `rms_users` - Retrieves all users
- `rms_invitations` - Retrieves all invitations
- `rms_permissions` - Retrieves all available permissions
- `rms_roles` - Retrieves all roles
- `rms_device_esim_bootstrap` - Retrieves eSIM bootstrap details for a device
- `rms_devices_export` - Retrieves all devices as CSV

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.5+ or OpenTofu 1.6+

### Building the Provider

```bash
git clone https://github.com/Moep90/terraform-provider-rms
cd terraform-provider-rms
go build -o terraform-provider-rms ./cmd/terraform-provider-rms
```

### Testing

```bash
# Run unit tests
go test -v ./...

# Run acceptance tests against mocked RMS endpoints
make testacc
```

### E2E tests

`make testacc-e2e` runs the suite against a real RMS tenant. Each E2E test
applies a configuration, confirms the object through an API read, destroys it
and confirms it is gone. It creates and deletes real objects in the tenant the
token belongs to. Objects are named with a per-run prefix (`tfe2e-<timestamp>`),
so a run that aborts mid-apply leaves identifiable debris.

Required:

- `TELTONIKA_RMS_TOKEN` or `RMS_ADMIN_TOKEN`: API token, read with the same
  precedence as the provider. Without one every E2E test skips.
- `RMS_PARENT_COMPANY_ID`: company the test objects are created under.
- `TF_ACC=1`: set by `make testacc-e2e`. Terraform's standard gate for tests
  that touch real infrastructure, so `go test ./...` never reaches RMS.

Optional:

- `TELTONIKA_RMS_BASE_URL`: API host, defaults to the provider default.
- `RMS_VPN_HUB_ZONE`: zone for the VPN hub test, defaults to `frankfurt-1`.
- `RMS_VPN_HUB_USER_ID`: hub user for the VPN hub route test.

```bash
export RMS_ADMIN_TOKEN="your-api-token"
export RMS_PARENT_COMPANY_ID="your-company-id"
make testacc-e2e
```

`rms_vpn_hub` and `rms_vpn_hub_route` are expected to fail: their create parses
an `id` the asynchronous RMS response does not carry. `rms_task` has no create
operation at all, so its E2E test asserts the apply-time rejection instead of a
lifecycle.

### Running with Terraform

```bash
# Install the provider locally
mkdir -p ~/.terraform.d/plugins/localhost/moep90/rms/1.0.0/linux_amd64
cp terraform-provider-rms ~/.terraform.d/plugins/localhost/moep90/rms/1.0.0/linux_amd64/

# Create a terraform.tfvars file with your token
echo 'teltonika_token = "your-token"' > terraform.tfvars

# Initialize and apply
terraform init
terraform apply
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Security

Please see our [Security Policy](SECURITY.md) for reporting vulnerabilities.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the GitHub repository.

## Acknowledgments

This provider was created to help manage Teltonika RMS resources through Infrastructure as Code. Special thanks to the Terraform and OpenTofu communities for their excellent tooling and documentation.
