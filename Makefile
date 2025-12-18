# SmartCalc - Makefile

# Variables
BINARY_NAME=smartcalc
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

# Detect OS for Wails build tags and CGO flags
# For dev: use -tags dev
# For production: use -tags "desktop,production" (or webkit2_41,production on Ubuntu 24.04+)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    # Ubuntu 24.04+ needs webkit2_41 instead of desktop
    WAILS_TAGS_DEV=-tags "dev,webkit2_41"
    WAILS_TAGS_PROD=-tags "desktop,production,webkit2_41"
    CGO_LDFLAGS_EXTRA=
else ifeq ($(UNAME_S),Darwin)
    WAILS_TAGS_DEV=-tags dev
    WAILS_TAGS_PROD=-tags "desktop,production"
    # macOS needs UniformTypeIdentifiers framework for newer SDKs
    CGO_LDFLAGS_EXTRA=-framework UniformTypeIdentifiers
else
    WAILS_TAGS_DEV=-tags dev
    WAILS_TAGS_PROD=-tags "desktop,production"
    CGO_LDFLAGS_EXTRA=
endif

.PHONY: all build clean test deps run help app-bundle release check frontend

# Default target
all: deps frontend build

# Build frontend
frontend:
	@echo "Building frontend..."
	@cd frontend && npm install && npm run build
	@echo "✓ Frontend build complete"

# Build the main application (production)
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 CGO_LDFLAGS="$(CGO_LDFLAGS_EXTRA)" $(GOBUILD) $(WAILS_TAGS_PROD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run in development mode
dev: frontend
	@echo "Running $(BINARY_NAME) in dev mode..."
	CGO_ENABLED=1 CGO_LDFLAGS="$(CGO_LDFLAGS_EXTRA)" $(GOCMD) run $(WAILS_TAGS_DEV) .

# Build macOS .app bundle
app-bundle: build
	@echo "Building macOS .app bundle..."
	@./scripts/build-app.sh
	@echo "✓ macOS .app bundle created"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download
	@echo "✓ Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f SmartCalc-*
	rm -rf "SmartCalc.app"
	rm -f *.zip
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "✓ Tests complete"

# Run the application in development mode
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install the application system-wide (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✓ Installation complete"

# Uninstall the application
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Uninstall complete"

# Create release build with optimizations
release: clean deps test app-bundle
	@echo "Creating release build..."
	@mkdir -p $(BUILD_DIR)/release
	@if [ -d "SmartCalc.app" ]; then cp -r "SmartCalc.app" $(BUILD_DIR)/release/; fi
	@echo "✓ Release build complete: $(BUILD_DIR)/release/"

# Check for common issues
check:
	@echo "Running checks..."
	$(GOCMD) fmt ./...
	$(GOCMD) vet ./...
	$(GOTEST) -race -short ./...
	@echo "✓ Checks complete"

# Display help
help:
	@echo "SmartCalc - Build Commands"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  all        - Install dependencies and build application"
	@echo "  build      - Build the main application"
	@echo "  app-bundle - Build macOS .app bundle"
	@echo "  deps       - Install Go dependencies"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run all tests"
	@echo "  run        - Build and run the application"
	@echo "  install    - Install application system-wide (requires sudo)"
	@echo "  uninstall  - Remove application from system"
	@echo "  release    - Create optimized release build"
	@echo "  check      - Run code formatting, vetting, and tests"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make deps   # Install dependencies"
	@echo "  make run    # Build and run the application"
	@echo ""
	@echo "Version: $(VERSION)" 