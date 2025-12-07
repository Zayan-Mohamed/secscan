# SecScan Makefile

# Variables
BINARY_NAME=secscan
VERSION=2.0.0
BUILD_DIR=build
INSTALL_PATH=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

.PHONY: all build clean test install uninstall run help install-local

## help: Display this help message
help:
	@echo "SecScan v$(VERSION) - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "⚡ Quick Start:"
	@echo "  make install        - Build and install globally (recommended, requires sudo)"
	@echo "  make install-local  - Build and install to ~/.local/bin (no sudo)"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/##//g'

## all: Build for all platforms
all: clean build-linux build-darwin build-windows

## build: Build for current platform
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux (amd64)
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "Linux build complete"

## build-darwin: Build for macOS (amd64 and arm64)
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@echo "macOS builds complete"

## build-windows: Build for Windows (amd64)
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "Windows build complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.json
	@echo "Clean complete"

## install: Install to system (requires sudo) - RECOMMENDED
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed successfully to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "Verify installation:"
	@echo "  $(BINARY_NAME) -version"
	@echo "  which $(BINARY_NAME)"

## install-local: Install to ~/.local/bin (no sudo required)
install-local: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "Installed successfully to ~/.local/bin/$(BINARY_NAME)"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo "~/.local/bin is already in your PATH"; \
		echo "Verify installation:"; \
		echo "  $(BINARY_NAME) -version"; \
	else \
		echo "~/.local/bin is NOT in your PATH"; \
		echo "Add this line to your ~/.bashrc or ~/.zshrc:"; \
		echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
		echo "Then run: source ~/.bashrc (or ~/.zshrc)"; \
	fi

## uninstall: Remove from system (requires sudo)
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled successfully"

## run: Build and run with default options
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## run-verbose: Build and run with verbose output
run-verbose: build
	@$(BUILD_DIR)/$(BINARY_NAME) -verbose

## run-fast: Build and run without git history
run-fast: build
	@$(BUILD_DIR)/$(BINARY_NAME) -history=false

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w main.go
	@echo "Format complete"

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
	fi

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

## release: Create release builds for all platforms
release: clean all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@cd $(BUILD_DIR) && zip -q releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release archives created in $(BUILD_DIR)/releases/"

## version: Display version information
version:
	@echo "SecScan v$(VERSION)"

.DEFAULT_GOAL := help
