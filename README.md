# Terraform Provider for Teltonika RMS

[![CI](https://github.com/YOUR_USERNAME/terraform-provider-teltonika-rms/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/terraform-provider-teltonika-rms/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/terraform-provider-teltonika-rms)](https://goreportcard.com/report/github.com/YOUR_USERNAME/terraform-provider-teltonika-rms)
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

## Installation

### Using Terraform

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    teltonika-rms = {
      source  = "YOUR_USERNAME/teltonika-rms"
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
    teltonika-rms = {
      source  = "YOUR_USERNAME/teltonika-rms"
      version = ">= 0.1.0"
    }
  }
}
```

## Configuration

Configure the provider with your Teltonika RMS API token:

```hcl
provider "teltonika-rms" {
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
resource "teltonika-rms_company" "main" {
  company_name = "My Company"
  parent_id    = null
}

resource "teltonika-rms_company" "subsidiary" {
  company_name = "Subsidiary Company"
  parent_id    = teltonika-rms_company.main.id
}
```

### Managing Devices

```hcl
resource "teltonika-rms_device" "router" {
  name               = "Office Router"
  device_series      = "rut"
  serial             = "0123456789"
  mac                = "00:11:22:33:44:55"
  company_id         = teltonika-rms_company.main.id
  auto_credit_enable = true
  password           = "device-password"
}
```

### Managing Tags

```hcl
resource "teltonika-rms_tag" "production" {
  name       = "Production"
  color      = "#00ff00"
  company_id = teltonika-rms_company.main.id
}

resource "teltonika-rms_tag" "development" {
  name       = "Development"
  color      = "#ff0000"
  company_id = teltonika-rms_company.main.id
}
```

### Using Data Sources

```hcl
data "teltonika-rms_companies" "all" {}

data "teltonika-rms_devices" "office" {
  company_id = teltonika-rms_company.main.id
  status     = "online"
}

data "teltonika-rms_tags" "all" {}
```

## Resources

- `teltonika-rms_company` - Manages a Teltonika RMS Company
- `teltonika-rms_device` - Manages a Teltonika RMS Device
- `teltonika-rms_tag` - Manages a Teltonika RMS Tag
- `teltonika-rms_user` - Manages a Teltonika RMS User
- `teltonika-rms_invitation` - Manages a Teltonika RMS User Invitation

## Data Sources

- `teltonika-rms_companies` - Retrieves all companies
- `teltonika-rms_company` - Retrieves a single company
- `teltonika-rms_devices` - Retrieves all devices with optional filtering
- `teltonika-rms_device` - Retrieves a single device
- `teltonika-rms_tags` - Retrieves all tags
- `teltonika-rms_users` - Retrieves all users
- `teltonika-rms_invitations` - Retrieves all invitations

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.5+ or OpenTofu 1.6+

### Building the Provider

```bash
git clone https://github.com/YOUR_USERNAME/terraform-provider-teltonika-rms
cd terraform-provider-teltonika-rms
go build -o terraform-provider-teltonika-rms ./cmd/terraform-provider-teltonika-rms
```

### Testing

```bash
# Run unit tests
go test -v ./...

# Run acceptance tests (requires real API credentials)
export TELTONIKA_RMS_TOKEN="your-api-token"
go test -v ./tests/acc/...
```

### Running with Terraform

```bash
# Install the provider locally
mkdir -p ~/.terraform.d/plugins/localhost/YOUR_USERNAME/teltonika-rms/1.0.0/linux_amd64
cp terraform-provider-teltonika-rms ~/.terraform.d/plugins/localhost/YOUR_USERNAME/teltonika-rms/1.0.0/linux_amd64/

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
