#!/bin/bash

# Build script for SuperCalc
# This script builds the application and creates a macOS .app bundle

set -e

echo "🧮 Building SuperCalc..."

# Get GOPATH and add bin to PATH
export GOPATH=$(go env GOPATH)
export PATH=$PATH:$GOPATH/bin

# Check if Fyne CLI is installed
if ! command -v fyne &> /dev/null; then
    echo "Installing Fyne CLI..."
    go install fyne.io/tools/cmd/fyne@latest
    export PATH=$PATH:$GOPATH/bin
fi

# Verify Fyne CLI is available
if ! command -v fyne &> /dev/null; then
    echo "❌ Error: Fyne CLI not found. Trying to install again..."
    go install fyne.io/tools/cmd/fyne@latest
    export PATH=$PATH:$GOPATH/bin
fi

# Get version information from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Extract x.y.z for app version (strip leading 'v' and any suffix after '-')
APP_VERSION=$(echo "$VERSION" | sed -E 's/^v?([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
if [ -z "$APP_VERSION" ] || [ "$APP_VERSION" = "$VERSION" ]; then
    APP_VERSION="1.0.0"
fi

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    GOARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    GOARCH="arm64"
else
    GOARCH="amd64"
fi

OUTFILE=SuperCalc-darwin-${VERSION}-${GOARCH}

echo "Building version: $VERSION (commit: $GIT_COMMIT, built: $BUILD_DATE)"
echo "Architecture: $GOARCH"

# Build the binary
CGO_ENABLED=1 GOARCH=$GOARCH go build \
    -ldflags="-s -w -X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitCommit=$GIT_COMMIT" \
    -o $OUTFILE .

# Create macOS .app bundle using fyne
echo "Creating macOS .app bundle..."
fyne package \
    --os darwin \
    --name "SuperCalc" \
    --app-id com.supercalc.app \
    --app-version "$APP_VERSION" \
    --app-build 1 \
    --release \
    --executable ./$OUTFILE

# Zip the .app bundle
echo "Zipping macOS .app bundle..."
ZIP_NAME="SuperCalc-darwin-${VERSION}-${GOARCH}.zip"
zip -r "$ZIP_NAME" "SuperCalc.app"

echo "✅ Build complete!"
echo "📦 macOS .app bundle created: $ZIP_NAME"
echo "🚀 You can now drag SuperCalc.app to your Applications folder" 