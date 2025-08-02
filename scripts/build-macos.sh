#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# Get the directory where the script is located
SCRIPT_DIR=$(dirname "$0")
# Project root is one level up from the script directory
BASE_DIR=$(realpath "$SCRIPT_DIR/..")
CLIENT_DIR="$BASE_DIR/client"

APP_NAME="Smart Finder"
BINARY_NAME="smart-finder-client"
APP_BUNDLE_NAME="$APP_NAME.app"
DIST_DIR="$BASE_DIR/dist"
APP_BUNDLE_PATH="$DIST_DIR/$APP_BUNDLE_NAME"
SRC_ICON_PATH="$CLIENT_DIR/internal/icon/icon.png"

# --- 1. Pre-flight Checks ---
echo "--- 1. Running pre-flight checks..."
if ! command -v sips &> /dev/null; then
    echo "!!! Error: 'sips' command not found. This script requires macOS."
    exit 1
fi
if ! command -v iconutil &> /dev/null; then
    echo "!!! Error: 'iconutil' command not found. This script requires macOS."
    exit 1
fi
if [ ! -f "$SRC_ICON_PATH" ]; then
    echo "!!! Error: Source icon not found at $SRC_ICON_PATH"
    exit 1
fi
echo "    > Checks passed."

# --- 2. Setup Directories ---
echo "--- 2. Cleaning up and creating directories..."
rm -rf "$APP_BUNDLE_PATH"
mkdir -p "$APP_BUNDLE_PATH/Contents/MacOS"
mkdir -p "$APP_BUNDLE_PATH/Contents/Resources"
# Use a temporary directory for all intermediate files
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
ICONSET_DIR="$TMP_DIR/icons.iconset"
mkdir -p "$ICONSET_DIR"

# --- 3. Create .icns file ---
echo "--- 3. Creating .icns file from $SRC_ICON_PATH..."

# LAUNDER THE SOURCE ICON to ensure it's a standard PNG format
echo "    > Standardizing source icon..."
CLEAN_ICON_PATH="$TMP_DIR/clean_icon.png"
sips -s format png "$SRC_ICON_PATH" --out "$CLEAN_ICON_PATH"

echo "    > Generating different icon sizes from standardized icon..."
# Generate all required sizes from the CLEAN icon
sips -z 16 16     "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_16x16.png"
sips -z 32 32     "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_16x16@2x.png"
sips -z 32 32     "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_32x32.png"
sips -z 64 64     "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_32x32@2x.png"
sips -z 128 128   "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_128x128.png"
sips -z 256 256   "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_128x128@2x.png"
sips -z 256 256   "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_256x256.png"
sips -z 512 512   "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_256x256@2x.png"
sips -z 512 512   "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_512x512.png"
sips -z 1024 1024 "$CLEAN_ICON_PATH" --out "$ICONSET_DIR/icon_512x512@2x.png"

# Convert the iconset to a single .icns file
echo "    > Converting to .icns format..."
iconutil -c icns "$ICONSET_DIR" -o "$APP_BUNDLE_PATH/Contents/Resources/icon.icns"
echo "    > .icns file created successfully."

# --- 4. Build Go Binary ---
echo "--- 4. Building Go binary..."
cd "$CLIENT_DIR"
go build -ldflags="-s -w" -o "$APP_BUNDLE_PATH/Contents/MacOS/$BINARY_NAME" .
cd "$BASE_DIR"
echo "    > Go binary built successfully."

# --- 5. Create Info.plist ---
echo "--- 5. Creating Info.plist..."
PLIST_PATH="$APP_BUNDLE_PATH/Contents/Info.plist"
cat > "$PLIST_PATH" << EOL
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>${BINARY_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>CFBundleIdentifier</key>
    <string>com.yourcompany.smartfinder</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
EOL
echo "    > Info.plist created successfully."

# --- 6. Ad-hoc Sign ---
echo "--- 6. Ad-hoc signing the application..."
codesign --force --deep --sign - "$APP_BUNDLE_PATH"
echo "    > Application signed."

echo ""
echo "✅ Success! Application bundle created at:"
echo "$APP_BUNDLE_PATH"
echo ""