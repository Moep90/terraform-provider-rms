---
page_title: "Teltonika RMS Provider"
description: |-
  The Teltonika RMS provider provides resources to interact with Teltonika RMS.
---

# Teltonika RMS Provider

The Teltonika RMS provider provides resources to interact with [Teltonika RMS](https://rms.teltonika-networks.com/).

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
terraform {
  required_providers {
    teltonika_rms = {
      source = "teltonika-rms/teltonika_rms"
      version = ">= 0.1.0"
    }
  }
}

provider "teltonika_rms" {
  token = var.teltonika_token
}
```

## Argument Reference

The following arguments are supported:

- `token` - (Optional) API token for authentication. Can also be set via the `TELTONIKA_RMS_TOKEN` environment variable.
- `base_url` - (Optional) Base URL for the Teltonika RMS API. Defaults to `https://rms.teltonika-networks.com/api`. Can also be set via the `TELTONIKA_RMS_BASE_URL` environment variable.
- `timeout` - (Optional) Request timeout in seconds. Defaults to `30`.
- `max_retry` - (Optional) Maximum number of retries for failed requests. Defaults to `3`.
