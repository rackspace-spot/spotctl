# Makefile for spotctl

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=spotctl
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN_PATH=.
DIST=dist

# Build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Ldflags
LDFLAGS=-ldflags "-X github.com/rackspace-spot/spotctl/internal/version.Version=$(VERSION) \
	-X github.com/rackspace-spot/spotctl/internal/version.Commit=$(GIT_COMMIT) \
	-X github.com/rackspace-spot/spotctl/internal/version.BuildDate=$(BUILD_DATE)"

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v $(MAIN_PATH)

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run race tests
.PHONY: test-race
test-race:
	$(GOTEST) -race -short ./...

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Vet code
.PHONY: vet
vet:
	$(GOCMD) vet ./...

# Run golint
.PHONY: lint
lint:
	golangci-lint run

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Tidy up dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Install binary to GOPATH/bin
.PHONY: install
install:
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)

# Cross compilation
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v $(MAIN_PATH)

.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) -v $(MAIN_PATH)

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)_darwin -v $(MAIN_PATH)

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin

# Run the application
.PHONY: run
run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v $(MAIN_PATH) && ./$(BINARY_NAME)

# Development workflow: format, vet, test, build
.PHONY: dev
dev: fmt vet test build

# CI workflow: deps, fmt, vet, test, build
.PHONY: ci
ci: deps fmt vet test build

# Release workflow: clean, deps, fmt, vet, test, build-all
.PHONY: release
release: clean deps fmt vet test build-all-named

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-race    - Run race tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run golangci-lint"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy up dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  build-linux  - Cross compile for Linux"
	@echo "  build-windows- Cross compile for Windows"
	@echo "  build-darwin - Cross compile for macOS"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-all-named - Build versioned artifacts to dist/"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Development workflow (fmt, vet, test, build)"
	@echo "  ci           - CI workflow (deps, fmt, vet, test, build)"
	@echo "  release      - Release workflow (clean, deps, fmt, vet, test, build-all-named)"
	@echo "  help         - Show this help message"

# Ensure dist directory exists
.PHONY: distdir
distdir:
	mkdir -p $(DIST)

# Build cross-platform artifacts with versioned filenames
.PHONY: build-all-named
build-all-named: distdir
	@echo "Building versioned artifacts with VERSION=$(VERSION) COMMIT=$(GIT_COMMIT) DATE=$(BUILD_DATE)"
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST)/$(BINARY_NAME)-$(VERSION)-linux-amd64 -v $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST)/$(BINARY_NAME)-$(VERSION)-linux-arm64 -v $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 -v $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 -v $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST)/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe -v $(MAIN_PATH)

