# Terraform Provider Teltonika RMS Makefile

# Binary name
BINARY_NAME=terraform-provider-teltonika-rms
BIN_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOIMPORTS=goimports

# Build parameters
VERSION?=0.1.0
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Directories
CMD_DIR=cmd/terraform-provider-teltonika-rms
INTERNAL_DIR=internal
TEST_DIR=tests

.PHONY: all build test lint fmt clean help install-tools pre-commit release-check

# Default target
all: build

## Build the provider
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

## Build with debug symbols
build-debug:
	@echo "Building $(BINARY_NAME) with debug symbols..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -gcflags="all=-N -l" -o $(BIN_DIR)/$(BINARY_NAME)-debug ./$(CMD_DIR)

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./... -skip TestAcc

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Run only unit tests (no acceptance tests)
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./$(INTERNAL_DIR)/api/...

## Run linting
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

## Format code
fmt:
	@echo "Formatting Go files..."
	$(GOFMT) ./...

## Verify code formatting
fmt-check:
	@echo "Checking Go formatting..."
	@if [ -n "$$(go fmt ./...)" ]; then \
		echo "Go files are not properly formatted. Run 'make fmt'"; \
		exit 1; \
	fi

## Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## Run all checks
check: fmt-check vet lint
	@echo "All checks passed!"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BIN_DIR)/$(BINARY_NAME)
	rm -f $(BIN_DIR)/$(BINARY_NAME)-debug
	rm -f coverage.out coverage.html
	rm -rf dist/ build/

## Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/google/addlicense@latest

## Run pre-commit hooks
pre-commit:
	@echo "Running pre-commit hooks..."
	pre-commit run --all-files

## Run release checklist
release-check:
	@echo "Running release checklist..."
	./scripts/release-check.sh

## Install provider locally
install: build
	@echo "Installing provider locally..."
	mkdir -p ~/.terraform.d/plugins/localhost/teltonika-rms/teltonika-rms/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)
	cp $(BIN_DIR)/$(BINARY_NAME) ~/.terraform.d/plugins/localhost/teltonika-rms/teltonika-rms/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

## Generate documentation
docs:
	@echo "Generating documentation..."
	# This would use tfplugindocs if installed

## Release (creates tag and triggers CI)
release:
	@echo "Preparing release $(VERSION)..."
	@echo "1. Ensure all tests pass"
	@echo "2. Update CHANGELOG.md"
	@echo "3. Run: git tag -a v$(VERSION) -m 'Release v$(VERSION)'"
	@echo "4. Run: git push origin v$(VERSION)"
	@echo "Semantic release will handle the rest!"

## Show help
help:
	@echo "Terraform Provider Teltonika RMS - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Commands:"
	@echo "  build       - Build the provider binary"
	@echo "  build-debug - Build with debug symbols"
	@echo "  install     - Install provider locally"
	@echo ""
	@echo "Test Commands:"
	@echo "  test        - Run all tests"
	@echo "  test-unit   - Run only unit tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint        - Run golangci-lint"
	@echo "  fmt         - Format Go files"
	@echo "  fmt-check   - Check if files are formatted"
	@echo "  vet         - Run go vet"
	@echo "  check       - Run all quality checks"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean       - Remove build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  pre-commit  - Run pre-commit hooks"
	@echo "  release-check - Run release checklist"
	@echo ""
	@echo "Release:"
	@echo "  release     - Prepare for release"
	@echo "  help        - Show this help message"
