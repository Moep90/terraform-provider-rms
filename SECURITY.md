# Security Policy

## Supported Versions

We release security patches for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take the security of the Teltonika RMS Terraform Provider seriously. If you believe you've found a security vulnerability, please report it responsibly.

### How to Report

**GitHub Security Advisory (Preferred)**: 
1. Go to the [Security tab](https://github.com/teltonika-rms/terraform-provider-teltonika-rms/security) of this repository
2. Click "Report a vulnerability"
3. Provide detailed information about the vulnerability

### What to Include

When reporting a vulnerability, please include:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested mitigations
- Whether you've disclosed this to anyone else

## Response Timeline

After submitting a security report, you can expect:

1. **Acknowledgment**: Within 48-72 hours
2. **Initial Assessment**: Within 5-10 business days
3. **Fix Development**: Timeline depends on severity
4. **Public Disclosure**: Coordinated disclosure, typically within 90 days

## Security Best Practices

### For Users

1. **Use HTTPS**: Always use HTTPS for API connections
2. **Token Management**: 
   - Store tokens securely using environment variables or secret management tools
   - Never commit tokens to version control
   - Rotate tokens regularly
3. **Principle of Least Privilege**: Use tokens with minimal required permissions
4. **Monitor Usage**: Regularly audit API usage and access logs

### For Contributors

1. **No Secrets in Code**: Never commit API keys, tokens, or credentials
2. **Use Environment Variables**: For testing, use environment variables for sensitive data
3. **Review Dependencies**: Keep dependencies up to date
4. **Security Reviews**: Participate in security code reviews

## Vulnerability Disclosure Policy

We follow a coordinated disclosure process:

1. Reporter submits vulnerability report
2. Maintainers verify and assess the vulnerability
3. Fix is developed and tested
4. Patched version is released
5. Security advisory is published with appropriate details

## Third-Party Dependencies

We use Dependabot to automatically monitor and update dependencies. All dependencies are reviewed for security vulnerabilities before updates are merged.

## Contact

For security questions or vulnerability reports, please use GitHub Security Advisories as described above.

## Acknowledgments

We thank the security community for responsible disclosure of vulnerabilities. Contributors will be acknowledged (with their permission) in our security advisories.
