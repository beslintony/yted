#!/bin/bash
# Build a .deb package for YTed

set -e

VERSION="${1:-$(grep '"productVersion"' wails.json | head -1 | sed -E 's/.*"productVersion": "([^"]+)".*/\1/')}"
ARCH="amd64"
PKG_NAME="yted_${VERSION}_${ARCH}"
BUILD_DIR="build/linux/deb-pkg"

# Clean previous build
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/DEBIAN"
mkdir -p "$BUILD_DIR/opt/yted"
mkdir -p "$BUILD_DIR/usr/bin"
mkdir -p "$BUILD_DIR/usr/share/applications"
mkdir -p "$BUILD_DIR/usr/share/pixmaps"
mkdir -p "$BUILD_DIR/usr/share/doc/yted"

# Copy app binary
APP_BINARY="build/bin/yted"
if [ ! -f "$APP_BINARY" ] && [ -f "build/bin/YTed" ]; then
    APP_BINARY="build/bin/YTed"
fi

if [ ! -f "$APP_BINARY" ]; then
    echo "Error: could not find built app binary at build/bin/yted or build/bin/YTed"
    exit 1
fi

cp "$APP_BINARY" "$BUILD_DIR/opt/yted/yted"
chmod +x "$BUILD_DIR/opt/yted/yted"

# Copy license files
cp LICENSE "$BUILD_DIR/usr/share/doc/yted/LICENSE"
cp LICENSE-THIRD-PARTY "$BUILD_DIR/usr/share/doc/yted/LICENSE-THIRD-PARTY"
gzip -9 -n -c > "$BUILD_DIR/usr/share/doc/yted/changelog.gz" <<EOF
yted ($VERSION) unstable; urgency=medium

  * Release version $VERSION

 -- beslintony <beslintony@gmail.com>  $(date -R)
EOF

# Symlink main binary
ln -s /opt/yted/yted "$BUILD_DIR/usr/bin/yted"

# Desktop entry and icon
cp build/linux/yted.desktop "$BUILD_DIR/usr/share/applications/yted.desktop"
cp build/appicon.png "$BUILD_DIR/usr/share/pixmaps/yted.png"

# Control file
cat > "$BUILD_DIR/DEBIAN/control" <<EOF
Package: yted
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Depends: libgtk-3-0, libwebkit2gtk-4.1-0
Maintainer: beslintony <beslintony@gmail.com>
Description: YTed - YouTube Downloader and Library Manager
 A modern, user-friendly YouTube downloader and library manager
 built with Go, Wails, and React.
 This package requires FFmpeg to be installed separately.
EOF

# Post-install script
cat > "$BUILD_DIR/DEBIAN/postinst" <<'EOF'
#!/bin/bash
set -e
if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database /usr/share/applications
fi
EOF
chmod 755 "$BUILD_DIR/DEBIAN/postinst"

# Build package
dpkg-deb --build "$BUILD_DIR" "build/bin/${PKG_NAME}.deb"

echo "Created build/bin/${PKG_NAME}.deb"
