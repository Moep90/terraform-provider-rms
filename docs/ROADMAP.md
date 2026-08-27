# Roadmap - Teltonika RMS Terraform Provider

## Current Status (v0.1.0)

### ✅ Completed
- Core provider framework
- 5 Resources: company, device, tag, user, invitation
- 7 Data Sources: companies, company, devices, device, tags, users, invitations
- CI/CD pipeline
- Documentation
- Unit tests

## Planned Features

### v0.2.0 - Enhanced Device Management
- [ ] Device configuration management
- [ ] Remote access configurations
- [ ] Device monitoring settings
- [ ] Firmware management

### v0.3.0 - Advanced Features
- [ ] Company hierarchy management
- [ ] Bulk device operations
- [ ] Tag assignment to devices
- [ ] User role management

### v0.4.0 - Integration Features
- [ ] Event/Alert management
- [ ] Dashboard configurations
- [ ] Reporting resources
- [ ] Integration with monitoring tools

### v0.5.0 - Enterprise Features
- [ ] Multi-company support
- [ ] Advanced access controls
- [ ] Audit logging
- [ ] Compliance features

## Future Considerations

### Long-term Goals
- Full RMS API coverage
- Acceptance tests with real API
- OpenTofu registry publication
- Terraform registry publication
- SDK for custom extensions

### Community Contributions
We welcome contributions for:
- New resources
- Bug fixes
- Documentation improvements
- Test coverage
- Performance improvements

## Deprecation Policy

- Features are deprecated with at least 2 minor versions notice
- Security fixes backported to last 2 minor versions
- Major version changes include migration guides

## Getting Involved

Want to help shape this roadmap?
- Open an issue with feature requests
- Submit pull requests
- Join discussions on GitHub
- Contribute documentation
