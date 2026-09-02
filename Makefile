# Terraform Provider Teltonika RMS Makefile

# Binary name
BINARY_NAME=terraform-provider-rms
TFPLUGINDOCS_VERSION=v0.25.0
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
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# Directories
CMD_DIR=cmd/terraform-provider-rms
INTERNAL_DIR=internal
TEST_DIR=tests

.PHONY: all build test lint fmt clean help install-tools pre-commit docs

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

## Run acceptance tests (mocked)
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 $(GOTEST) -v -race ./$(TEST_DIR)/acceptance/

## Run E2E acceptance tests against real API.
## These create and destroy real objects in the tenant the token belongs to.
## Required: TELTONIKA_RMS_TOKEN or RMS_ADMIN_TOKEN, plus RMS_PARENT_COMPANY_ID.
## Optional: TELTONIKA_RMS_BASE_URL, RMS_VPN_HUB_ZONE, RMS_VPN_HUB_USER_ID.
## Without a token every E2E test skips. TF_ACC gates them the rest of the time.
testacc-e2e:
	@echo "Running E2E acceptance tests against real API (creates and destroys real objects)..."
	TF_ACC=1 $(GOTEST) -v -race ./$(TEST_DIR)/acceptance/

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

## Install provider locally
install: build
	@echo "Installing provider locally..."
	mkdir -p ~/.terraform.d/plugins/localhost/moep90/rms/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)
	cp $(BIN_DIR)/$(BINARY_NAME) ~/.terraform.d/plugins/localhost/moep90/rms/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

## Generate documentation
# tfplugindocs builds the provider from the repository root, but package main
# lives in $(CMD_DIR), so export the schema via Terraform first and hand
# tfplugindocs the JSON. The throwaway workspace uses the hashicorp/ namespace
# because tfplugindocs looks the provider up under that key.
docs:
	@echo "Generating provider documentation..."
	$(eval DOCS_TMP := $(shell mktemp -d))
	mkdir -p $(DOCS_TMP)/plugins/registry.terraform.io/hashicorp/rms/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)
	$(GOBUILD) -o $(DOCS_TMP)/plugins/registry.terraform.io/hashicorp/rms/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)/terraform-provider-rms_v0.0.1 ./$(CMD_DIR)
	printf 'terraform {\n  required_providers {\n    rms = {\n      source  = "hashicorp/rms"\n      version = "0.0.1"\n    }\n  }\n}\n' > $(DOCS_TMP)/main.tf
	cd $(DOCS_TMP) && terraform init -plugin-dir=$(DOCS_TMP)/plugins -input=false > /dev/null
	cd $(DOCS_TMP) && terraform providers schema -json > $(DOCS_TMP)/schema.json
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$(TFPLUGINDOCS_VERSION) generate \
		--providers-schema $(DOCS_TMP)/schema.json \
		--provider-name rms \
		--rendered-provider-name "Teltonika RMS"
	rm -rf $(DOCS_TMP)

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
	@echo "  testacc     - Run acceptance tests against mocks"
	@echo "  testacc-e2e - Run E2E tests against real RMS (creates and destroys real objects)"
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
	@echo ""
	@echo "Release:"
	@echo "  release     - Prepare for release"
	@echo "  help        - Show this help message"
