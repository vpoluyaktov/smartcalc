#!/bin/bash

# Build script for SmartCalc (Wails)
# This script builds the application and creates a macOS .app bundle

set -e

echo "🧮 Building SmartCalc..."

# Get GOPATH and add bin to PATH
export GOPATH=$(go env GOPATH)
export PATH=$PATH:$GOPATH/bin

# Get version information
BASE_VERSION=$(cat VERSION 2>/dev/null || echo "0.1.0")
COMMIT_COUNT=$(git rev-list --all --count HEAD 2>/dev/null || echo "0")
# Extract MAJOR.MINOR from base version and use commit count as PATCH
MAJOR_MINOR=$(echo $BASE_VERSION | cut -d. -f1,2)
VERSION="$MAJOR_MINOR.$COMMIT_COUNT"
APP_VERSION="$VERSION"
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

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
    OUTFILE="SmartCalc-darwin-${VERSION}-${GOARCH}"
elif [ "$OS" = "Linux" ]; then
    OUTFILE="SmartCalc-linux-${VERSION}-${GOARCH}"
else
    OUTFILE="SmartCalc-${VERSION}-${GOARCH}"
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
    
    APP_NAME="SmartCalc.app"
    rm -rf "$APP_NAME"
    
    # Create app bundle structure
    mkdir -p "$APP_NAME/Contents/MacOS"
    mkdir -p "$APP_NAME/Contents/Resources"
    
    # Copy binary
    cp "$OUTFILE" "$APP_NAME/Contents/MacOS/SmartCalc"
    
    # Generate icns from PNG if needed (macOS only)
    if [ -f "assets/icon.png" ] && [ ! -f "build/appicon.icns" ]; then
        echo "Generating app icon..."
        mkdir -p build
        ICONSET="build/icon.iconset"
        mkdir -p "$ICONSET"
        
        # Generate all required sizes using sips (macOS built-in)
        sips -z 16 16     "assets/icon.png" --out "$ICONSET/icon_16x16.png" 2>/dev/null
        sips -z 32 32     "assets/icon.png" --out "$ICONSET/icon_16x16@2x.png" 2>/dev/null
        sips -z 32 32     "assets/icon.png" --out "$ICONSET/icon_32x32.png" 2>/dev/null
        sips -z 64 64     "assets/icon.png" --out "$ICONSET/icon_32x32@2x.png" 2>/dev/null
        sips -z 128 128   "assets/icon.png" --out "$ICONSET/icon_128x128.png" 2>/dev/null
        sips -z 256 256   "assets/icon.png" --out "$ICONSET/icon_128x128@2x.png" 2>/dev/null
        sips -z 256 256   "assets/icon.png" --out "$ICONSET/icon_256x256.png" 2>/dev/null
        sips -z 512 512   "assets/icon.png" --out "$ICONSET/icon_256x256@2x.png" 2>/dev/null
        sips -z 512 512   "assets/icon.png" --out "$ICONSET/icon_512x512.png" 2>/dev/null
        cp "assets/icon.png" "$ICONSET/icon_512x512@2x.png"
        
        # Convert to icns
        iconutil -c icns "$ICONSET" -o "build/appicon.icns" 2>/dev/null
        rm -rf "$ICONSET"
    fi
    
    # Copy icon if exists
    if [ -f "build/appicon.icns" ]; then
        cp "build/appicon.icns" "$APP_NAME/Contents/Resources/icon.icns"
    elif [ -f "assets/icon.icns" ]; then
        cp "assets/icon.icns" "$APP_NAME/Contents/Resources/icon.icns"
    fi
    
    # Create Info.plist
    cat > "$APP_NAME/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>SmartCalc</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.smartcalc.app</string>
    <key>CFBundleName</key>
    <string>SmartCalc</string>
    <key>CFBundleDisplayName</key>
    <string>SmartCalc</string>
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
    ZIP_NAME="SmartCalc-darwin-${VERSION}-${GOARCH}.zip"
    zip -r "$ZIP_NAME" "$APP_NAME"
    
    echo "✅ Build complete!"
    echo "📦 macOS .app bundle created: $ZIP_NAME"
    echo "🚀 You can now drag SmartCalc.app to your Applications folder"
else
    echo "✅ Build complete!"
    echo "📦 Binary created: $OUTFILE"
fi 