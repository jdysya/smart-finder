#!/bin/bash
set -e

if [[ "$RUNNER_OS" == "Windows" ]]; then
  export CGO_ENABLED=1
  go install github.com/tc-hib/go-winres@latest
  go-winres make --icon client/internal/icon/icon.png -o client/winres.syso
  cd client
  go build -ldflags="-s -w -H=windowsgui" -tags="osusergo,netgo" -o "smart-finder-client-${PLATFORM}-${ARCH}${EXT}" .
  cd ..
  ZIP_NAME="smart-finder-client-${PLATFORM}-${ARCH}.zip"
  powershell -Command "Compress-Archive -Path \"client/smart-finder-client-${PLATFORM}-${ARCH}${EXT}\" -DestinationPath \"client/${ZIP_NAME}\" -Force"
  if [[ -n "$GITHUB_OUTPUT" ]]; then
    echo "artifact_name=client-${PLATFORM}-${ARCH}-zip" >> $GITHUB_OUTPUT
    echo "artifact_path=client/${ZIP_NAME}" >> $GITHUB_OUTPUT
  else
    echo "Created artifact: client/${ZIP_NAME}"
  fi

elif [[ "$RUNNER_OS" == "macOS" ]]; then
  export CGO_ENABLED=1
  APP_NAME="Smart Finder"
  BINARY_NAME="smart-finder-client"
  APP_BUNDLE_NAME="$APP_NAME.app"
  DIST_DIR="dist"
  mkdir -p $DIST_DIR
  APP_BUNDLE_PATH="$DIST_DIR/$APP_BUNDLE_NAME"
  SRC_ICON_PATH="client/internal/icon/icon.png"
  mkdir -p "$APP_BUNDLE_PATH/Contents/MacOS"
  mkdir -p "$APP_BUNDLE_PATH/Contents/Resources"
  TMP_DIR=$(mktemp -d)
  ICONSET_DIR="$TMP_DIR/icons.iconset"
  mkdir -p "$ICONSET_DIR"
  CLEAN_ICON_PATH="$TMP_DIR/clean_icon.png"
  sips -s format png "$SRC_ICON_PATH" --out "$CLEAN_ICON_PATH"
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
  iconutil -c icns "$ICONSET_DIR" -o "$APP_BUNDLE_PATH/Contents/Resources/icon.icns"
  cd client
  go build -ldflags="-s -w" -o "../$APP_BUNDLE_PATH/Contents/MacOS/$BINARY_NAME" .
  cd ..
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

  codesign --force --deep --sign - "$APP_BUNDLE_PATH"
  DMG_NAME="smart-finder-client-${PLATFORM}-${ARCH}.dmg"
  hdiutil create -volname "$APP_NAME" -srcfolder "$APP_BUNDLE_PATH" -ov -format UDZO "$DIST_DIR/$DMG_NAME"
  if [[ -n "$GITHUB_OUTPUT" ]]; then
    echo "artifact_name=client-${PLATFORM}-${ARCH}-dmg" >> $GITHUB_OUTPUT
    echo "artifact_path=$DIST_DIR/$DMG_NAME" >> $GITHUB_OUTPUT
  else
    echo "Created artifact: $DIST_DIR/$DMG_NAME"
  fi

fi