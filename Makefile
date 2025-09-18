# bffgen Makefile

# Variables
BINARY_NAME=bffgen
VERSION?=dev
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_SHA=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT_SHA)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/bffgen
	@echo "✅ Build completed: $(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME); \
		if [ "$$os" = "windows" ]; then output_name=$(BINARY_NAME).exe; fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o dist/$(BINARY_NAME)-$$os-$$arch$$([ "$$os" = "windows" ] && echo .exe) ./cmd/bffgen; \
	done
	@cd dist && sha256sum * > checksums.txt
	@echo "✅ Multi-platform build completed"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...
	go test -race ./...
	@echo "✅ Tests completed"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	go vet ./...
	go fmt ./...
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code is not formatted"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "✅ Linting completed"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	@echo "✅ Clean completed"

# Install locally
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/bffgen
	@echo "✅ Installation completed"

# Create a release tag
.PHONY: tag
tag:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "❌ VERSION must be set (e.g., make tag VERSION=v0.1.0)"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) created and pushed"

# Prepare release (build, test, lint)
.PHONY: release-prep
release-prep: clean test lint build-all
	@echo "✅ Release preparation completed"

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT_SHA)"

# Development server
.PHONY: dev
dev:
	@echo "Starting development server..."
	go run ./cmd/bffgen

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install locally"
	@echo "  tag          - Create and push a git tag (requires VERSION=v0.1.0)"
	@echo "  release-prep - Prepare release (build, test, lint)"
	@echo "  version      - Show version information"
	@echo "  dev          - Start development server"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make tag VERSION=v0.1.0"
	@echo "  make release-prep"
