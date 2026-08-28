# Migration Guide - Teltonika RMS Terraform Provider

## Upgrading from Previous Versions

### Version 0.1.0 (Current)

This is the initial release. No migration needed.

## Migrating from Other Providers

### From Teltonika RMS CLI

If you're moving from the Teltonika RMS CLI tools:

```bash
# Old CLI approach
rms-cli device create --name "My Device" --serial "12345"

# New Terraform approach
resource "rms_device" "example" {
  name          = "My Device"
  device_series = "rut"
  serial        = "12345"
  company_id    = 12345
}
```

### From Manual RMS Management

If you're moving from manual RMS web interface management:

1. **Import existing resources**
```bash
# Import existing company
terraform import rms_company.main 12345

# Import existing device
terraform import rms_device.router 67890
```

2. **Generate configuration**
```bash
terraform plan -out=tfplan
terraform show -json tfplan > plan.json
```

## State Migration

### Moving State to Remote Backend

```bash
# Initialize with local state
terraform init

# Configure remote backend
terraform backend s3 \
  -bucket=my-terraform-state \
  -key=teltonika-rms.tfstate \
  -region=us-east-1

# Migrate state
terraform init -migrate-state
```

### Importing Existing Resources

```hcl
# Example: Import a company
terraform import rms_company.main 12345

# Example: Import a device
terraform import rms_device.router 67890

# Example: Import a tag
terraform import rms_tag.production 11111
```

## Breaking Changes

### Version 0.1.0

No breaking changes in initial release.

## Migration Checklist

- [ ] Backup existing RMS configuration
- [ ] Review provider documentation
- [ ] Set up authentication (API token)
- [ ] Import existing resources
- [ ] Test in non-production environment
- [ ] Plan migration window
- [ ] Execute migration
- [ ] Verify all resources
- [ ] Update documentation
