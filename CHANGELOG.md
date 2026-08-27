# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of the Teltonika RMS Terraform Provider

## [0.1.0] - 2024-01-01

### Added
- Provider framework with authentication support
- API client with retry logic and error handling
- Resource: `teltonika-rms_company` - Manage RMS companies
- Resource: `teltonika-rms_device` - Manage devices (RUT, TRB, etc.)
- Resource: `teltonika-rms_tag` - Manage tags for organizing devices
- Resource: `teltonika-rms_user` - Manage users
- Resource: `teltonika-rms_invitation` - Manage user invitations
- Data source: `teltonika-rms_companies` - List all companies
- Data source: `teltonika-rms_company` - Get single company
- Data source: `teltonika-rms_devices` - List devices with filters
- Data source: `teltonika-rms_device` - Get single device
- Data source: `teltonika-rms_tags` - List all tags
- Data source: `teltonika-rms_users` - List all users
- Data source: `teltonika-rms_invitations` - List all invitations
- Comprehensive documentation
- Unit tests with 62.5% code coverage
- CI/CD pipeline for Terraform 1.5+ and OpenTofu
- GitHub Actions workflows
- Security policy and vulnerability reporting process
- Contributing guidelines
- Code of Conduct

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- Implemented secure token handling
- Added security policy documentation

[Unreleased]: https://github.com/teltonika-rms/terraform-provider-teltonika-rms/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/teltonika-rms/terraform-provider-teltonika-rms/releases/tag/v0.1.0
