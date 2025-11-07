# Makefile for aws-ssm

# Version information
VERSION ?= 0.1.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS := -ldflags "\
	-X github.com/aws-ssm/pkg/version.Version=$(VERSION) \
	-X github.com/aws-ssm/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/aws-ssm/pkg/version.BuildDate=$(BUILD_DATE)"

# Binary name
BINARY_NAME := aws-ssm

# Build directory
BUILD_DIR := dist

# Platforms to build for
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all
all: clean test build

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build for current platform
	@echo "Building $(BINARY_NAME) for current platform..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

.PHONY: build-all
build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(call build_platform,$(platform)))
	@echo "All builds complete!"

define build_platform
	$(eval OS := $(word 1,$(subst /, ,$(1))))
	$(eval ARCH := $(word 2,$(subst /, ,$(1))))
	$(eval OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.exe,))
	@echo "Building for $(OS)/$(ARCH)..."
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(OUTPUT) .
endef

.PHONY: release
release: clean test build-all checksums archives ## Build release artifacts
	@echo "Release artifacts created in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

.PHONY: archives
archives: ## Create compressed archives for release
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
	for file in $(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			case "$$file" in \
				*.exe) \
					zip "$${file%.exe}.zip" "$$file" && rm "$$file" ;; \
				*) \
					tar -czf "$$file.tar.gz" "$$file" && rm "$$file" ;; \
			esac; \
		fi; \
	done
	@echo "Archives created!"

.PHONY: checksums
checksums: ## Generate SHA256 checksums
	@echo "Generating checksums..."
	@cd $(BUILD_DIR) && \
	if command -v shasum >/dev/null 2>&1; then \
		shasum -a 256 $(BINARY_NAME)-* > checksums.txt 2>/dev/null || true; \
	elif command -v sha256sum >/dev/null 2>&1; then \
		sha256sum $(BINARY_NAME)-* > checksums.txt 2>/dev/null || true; \
	fi
	@echo "Checksums generated: $(BUILD_DIR)/checksums.txt"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete!"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted!"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete!"

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "Modules tidied!"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

.PHONY: install
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

.PHONY: uninstall
uninstall: ## Uninstall binary from GOPATH/bin
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Uninstalled!"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded!"

.PHONY: verify
verify: fmt vet lint test ## Run all verification checks
	@echo "All verification checks passed!"

.PHONY: version
version: ## Display version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

.PHONY: run
run: build ## Build and run
	@./$(BINARY_NAME) $(ARGS)

.PHONY: dev
dev: ## Run in development mode (with hot reload if available)
	@go run $(LDFLAGS) . $(ARGS)

# Release preparation
.PHONY: pre-release
pre-release: verify ## Verify everything is ready for release
	@echo "Checking release readiness..."
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set"; exit 1; fi
	@if git diff-index --quiet HEAD --; then \
		echo "✓ Working directory is clean"; \
	else \
		echo "✗ Working directory has uncommitted changes"; \
		exit 1; \
	fi
	@if git tag | grep -q "^v$(VERSION)$$"; then \
		echo "✗ Tag v$(VERSION) already exists"; \
		exit 1; \
	else \
		echo "✓ Tag v$(VERSION) is available"; \
	fi
	@echo "✓ Ready for release!"

.PHONY: tag
tag: ## Create and push git tag
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set"; exit 1; fi
	@echo "Creating tag v$(VERSION)..."
	@git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tag created. Push with: git push origin v$(VERSION)"

.PHONY: docker-build
docker-build: ## Build Docker image (if Dockerfile exists)
	@if [ -f Dockerfile ]; then \
		docker build -t aws-ssm:$(VERSION) .; \
	else \
		echo "Dockerfile not found"; \
	fi

