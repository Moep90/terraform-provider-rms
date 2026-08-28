# FAQ - Teltonika RMS Terraform Provider

## General Questions

### What is this provider?
This is a community-maintained Terraform/OpenTofu provider for managing Teltonika RMS resources. It is not officially supported by Teltonika Networks.

### Which Teltonika RMS API version does this use?
This provider uses the Teltonika RMS API v3 (currently in BETA).

### Can I use this with OpenTofu?
Yes! This provider is compatible with both Terraform >= 1.5.0 and OpenTofu >= 1.6.0.

## Authentication

### How do I authenticate?
You can authenticate using:
- Environment variable: `TELTONIKA_RMS_TOKEN`
- Provider configuration: `token` argument

### Where do I get an API token?
Generate a Personal Access Token (PAT) from your Teltonika RMS account settings.

### Can I use OAuth?
Currently, this provider only supports Personal Access Tokens. OAuth support may be added in future versions.

## Resources

### Can I manage devices?
Yes, the `rms_device` resource supports creating and managing devices.

### Do tags work with devices?
Yes, you can create tags and assign them to devices.

### Can I manage users?
Yes, the `rms_user` resource allows user management.

## Troubleshooting

### "Unauthorized" error
Check that your API token is valid and has the necessary permissions.

### "Rate limited" error
The API has rate limits. Reduce the frequency of API calls or contact Teltonika for higher limits.

### Device not appearing after creation
Devices may take a few moments to appear in RMS. Wait a moment and refresh.

## Best Practices

### Should I store tokens in Terraform state?
No. Use environment variables or a secrets manager to avoid storing tokens in state files.

### How do I handle hierarchical companies?
Use the `parent_id` attribute on the `rms_company` resource.

### Can I import existing resources?
Yes, all resources support import using their ID.

## Contributing

### How do I report a bug?
Open an issue on GitHub with detailed reproduction steps.

### How do I contribute code?
See our [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.
