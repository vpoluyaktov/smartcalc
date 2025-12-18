#!/bin/bash

# Build script for SuperCalc (Wails)
# This script builds the application and creates a macOS .app bundle

set -e

echo "🧮 Building SuperCalc..."

# Get GOPATH and add bin to PATH
export GOPATH=$(go env GOPATH)
export PATH=$PATH:$GOPATH/bin

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

echo "Building version: $VERSION (commit: $GIT_COMMIT, built: $BUILD_DATE)"
echo "Architecture: $GOARCH"

# Build frontend
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Detect OS for build tags
OS=$(uname -s)
if [ "$OS" = "Darwin" ]; then
    WAILS_TAGS="desktop,production"
    # macOS needs UniformTypeIdentifiers framework
    export CGO_LDFLAGS="-framework UniformTypeIdentifiers"
elif [ "$OS" = "Linux" ]; then
    WAILS_TAGS="desktop,production,webkit2_41"
else
    WAILS_TAGS="desktop,production"
fi

# Set output filename based on OS
if [ "$OS" = "Darwin" ]; then
    OUTFILE="SuperCalc-darwin-${VERSION}-${GOARCH}"
elif [ "$OS" = "Linux" ]; then
    OUTFILE="SuperCalc-linux-${VERSION}-${GOARCH}"
else
    OUTFILE="SuperCalc-${VERSION}-${GOARCH}"
fi

# Build the binary with Wails
echo "Building binary with Wails..."
CGO_ENABLED=1 GOARCH=$GOARCH go build \
    -tags "$WAILS_TAGS" \
    -ldflags="-s -w -X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitCommit=$GIT_COMMIT" \
    -o "$OUTFILE" .

# Create macOS .app bundle (macOS only)
if [ "$OS" = "Darwin" ]; then
    echo "Creating macOS .app bundle..."
    
    APP_NAME="SuperCalc.app"
    rm -rf "$APP_NAME"
    
    # Create app bundle structure
    mkdir -p "$APP_NAME/Contents/MacOS"
    mkdir -p "$APP_NAME/Contents/Resources"
    
    # Copy binary
    cp "$OUTFILE" "$APP_NAME/Contents/MacOS/SuperCalc"
    
    # Copy icon if exists
    if [ -f "assets/icon.icns" ]; then
        cp "assets/icon.icns" "$APP_NAME/Contents/Resources/icon.icns"
    elif [ -f "build/appicon.icns" ]; then
        cp "build/appicon.icns" "$APP_NAME/Contents/Resources/icon.icns"
    fi
    
    # Create Info.plist
    cat > "$APP_NAME/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>SuperCalc</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.supercalc.app</string>
    <key>CFBundleName</key>
    <string>SuperCalc</string>
    <key>CFBundleDisplayName</key>
    <string>SuperCalc</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
</dict>
</plist>
EOF
    
    # Clear extended attributes and sign
    xattr -cr "$APP_NAME" 2>/dev/null || true
    codesign --force --deep --sign - "$APP_NAME" 2>/dev/null || true
    
    # Zip the .app bundle
    echo "Zipping macOS .app bundle..."
    ZIP_NAME="SuperCalc-darwin-${VERSION}-${GOARCH}.zip"
    zip -r "$ZIP_NAME" "$APP_NAME"
    
    echo "✅ Build complete!"
    echo "📦 macOS .app bundle created: $ZIP_NAME"
    echo "🚀 You can now drag SuperCalc.app to your Applications folder"
else
    echo "✅ Build complete!"
    echo "📦 Binary created: $OUTFILE"
fi 