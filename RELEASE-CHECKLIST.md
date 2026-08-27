# Pre-Publication Checklist

This document outlines the steps needed before publicly releasing the Terraform Provider for Teltonika RMS.

## ✅ Completed Items

- [x] Provider implementation (5 resources, 7 data sources)
- [x] Unit tests with 62.5% code coverage
- [x] CI/CD pipeline for Terraform 1.5+ and OpenTofu
- [x] Complete documentation (README, website/docs)
- [x] Examples for all resources and data sources
- [x] LICENSE file (MIT)
- [x] SECURITY.md - Security policy
- [x] CODE_OF_CONDUCT.md - Code of Conduct
- [x] CONTRIBUTING.md - Contribution guidelines
- [x] Issue templates (bug report, feature request, question)
- [x] Pull Request template
- [x] Dependabot configuration
- [x] Renovate configuration
- [x] Code owners file
- [x] CHANGELOG.md
- [x] Pre-commit hooks configuration
- [x] Go linter configuration (.golangci.yml)
- [x] GoReleaser configuration
- [x] Semantic Release configuration

## ⚠️ Items Requiring Your Input

### 1. Replace Placeholders

Replace all instances of `YOUR_USERNAME` with your actual GitHub username:

```bash
# In README.md
sed -i 's/YOUR_USERNAME/YOUR_ACTUAL_USERNAME/g' README.md

# In .github/dependabot.yml
sed -i 's/YOUR_USERNAME/YOUR_ACTUAL_USERNAME/g' .github/dependabot.yml

# In .github/CODEOWNERS
sed -i 's/YOUR_USERNAME/YOUR_ACTUAL_USERNAME/g' .github/CODEOWNERS

# In CHANGELOG.md
sed -i 's/YOUR_USERNAME/YOUR_ACTUAL_USERNAME/g' CHANGELOG.md

# In .github/workflows/ci.yml
sed -i 's/YOUR_USERNAME/YOUR_ACTUAL_USERNAME/g' .github/workflows/ci.yml
```

### 2. Update LICENSE File

Edit the LICENSE file and replace `[YEAR]` and `[YOUR NAME]` with actual values:

```
MIT License

Copyright (c) 2024 [Your Full Name]
```

### 3. Create GitHub Repository

```bash
# Create repository on GitHub (via UI or CLI)
gh repo create YOUR_USERNAME/terraform-provider-teltonika-rms --public --description "Terraform/OpenTofu provider for Teltonika RMS"

# Add remote and push
git remote add origin https://github.com/YOUR_USERNAME/terraform-provider-teltonika-rms.git
git branch -M main
git push -u origin main
```

### 4. Create Initial Release

```bash
# Create initial tag
git tag -a v0.1.0 -m "Initial release of Teltonika RMS provider"

# Push tag
git push origin v0.1.0

# Create GitHub release
gh release create v0.1.0 \
  --title "v0.1.0 - Initial Release" \
  --notes "Initial release of the Teltonika RMS Terraform/OpenTofu provider. See CHANGELOG.md for details."
```

### 5. Enable GitHub Features

On GitHub repository settings:
- [ ] Enable Issues
- [ ] Enable Discussions (optional)
- [ ] Enable Projects (optional)
- [ ] Enable GitHub Pages (for documentation, optional)
- [ ] Add topic tags: `terraform`, `opentofu`, `terraform-provider`, `teltonika`, `rms`, `iot`

### 6. Configure Branch Protection

On GitHub repository settings > Branches:
- [ ] Add branch protection rule for `main`
- [ ] Require pull request reviews before merging
- [ ] Require status checks to pass before merging
- [ ] Require branches to be up to date before merging
- [ ] Enable "Include administrators"

### 7. Set Up Secrets (Optional)

If you want to enable automated releases:
- [ ] Add `GITHUB_TOKEN` secret (automatically available)
- [ ] Add `GPG_KEY_ID` and `GPG_PRIVATE_KEY` for signing releases (optional)

## 📋 Post-Publication Tasks

### For OpenTofu Registry

1. Fork to OpenTofu provider registry format
2. Submit to OpenTofu registry following their guidelines
3. Update documentation with OpenTofu installation instructions

### For Terraform Registry

1. Read [HashiCorp's provider submission guidelines](https://registry.terraform.io/docs/providers/development)
2. Submit a provider submission request via GitHub issue to HashiCorp
3. Wait for approval (can take several weeks)
4. Once approved, follow their onboarding process

## 🔒 Security Checklist

- [ ] No secrets or credentials in code or history
- [ ] All dependencies reviewed for vulnerabilities
- [ ] Security policy in place and visible
- [ ] Dependabot and Renovate enabled
- [ ] Security contact method defined

## 📖 Documentation Checklist

- [ ] README.md is complete and clear
- [ ] All resources have documentation
- [ ] All data sources have documentation
- [ ] Examples are provided and tested
- [ ] CONTRIBUTING.md is clear
- [ ] SECURITY.md is complete
- [ ] CODE_OF_CONDUCT.md is in place
- [ ] CHANGELOG.md is maintained

## 🧪 Testing Checklist

- [ ] All unit tests pass
- [ ] CI pipeline passes on all branches
- [ ] Provider builds successfully
- [ ] `go mod tidy` produces no changes
- [ ] `go fmt` produces no changes
- [ ] `golangci-lint` passes without errors

## ✨ Final Review

Before going public, ask yourself:

1. Would I be comfortable if a potential employer saw this code?
2. Is there any sensitive information that could leak?
3. Is the code quality professional and maintainable?
4. Is the documentation clear enough for someone unfamiliar with the project?
5. Are the license and legal aspects properly addressed?

If you can answer "yes" to all these questions, you're ready to publish!

## 🚀 Ready to Publish

Once all items above are completed, you're ready to make your repository public and start building your open-source community!
