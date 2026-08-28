# Contributing to Terraform Provider Teltonika RMS

Thank you for your interest in contributing to the Teltonika RMS Terraform Provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to create a welcoming environment for everyone.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Terraform 1.5+ or OpenTofu 1.6+
- Git
- A GitHub account

### Setup Development Environment

```bash
# Fork the repository
git clone https://github.com/Moep90/terraform-provider-rms.git
cd terraform-provider-rms

# Install dependencies
go mod tidy

# Run tests
go test -v ./...

# Build the provider
go build -o terraform-provider-rms ./cmd/terraform-provider-rms
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Follow Conventional Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

Example:
```bash
git commit -m "feat: add support for device monitoring status"
```

### 3. Write Tests

- Add unit tests for new functionality
- Maintain or improve code coverage
- Run all tests before submitting:

```bash
go test -v -race -cover ./...
```

### 4. Follow Go Best Practices

- Run `go fmt` and `go vet` before committing:
```bash
go fmt ./...
go vet ./...
```

- Use `golangci-lint` for additional linting:
```bash
golangci-lint run
```

### 5. Update Documentation

- Regenerate resource/data source documentation with `make docs`
- Add examples in `examples/`
- Update README.md if needed

## Pull Request Process

### 1. Create a Pull Request

- Push your branch to your fork
- Create a PR against the `main` branch
- Fill out the PR template completely

### 2. PR Requirements

- All CI checks must pass
- Code review from at least one maintainer
- No merge conflicts
- Squash commits into meaningful units

### 3. Code Review

- Be responsive to review feedback
- Address comments promptly
- Explain design decisions in comments

## Release Process

Releases follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality
- **PATCH** version for backwards-compatible bug fixes

Releases are automated via GitHub Actions and semantic-release.

## Questions?

- Open an issue for questions
- Join our discussions on GitHub Discussions

## Recognition

Contributors will be recognized in:
- Git commit history
- CHANGELOG.md
- README.md contributors section

Thank you for contributing!
