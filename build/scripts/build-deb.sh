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

# Copy app binary and bundled FFmpeg
cp build/bin/yted "$BUILD_DIR/opt/yted/yted"
chmod +x "$BUILD_DIR/opt/yted/yted"

if [ -f build/bin/ffmpeg ]; then
    cp build/bin/ffmpeg "$BUILD_DIR/opt/yted/ffmpeg"
    chmod +x "$BUILD_DIR/opt/yted/ffmpeg"
fi
if [ -f build/bin/ffprobe ]; then
    cp build/bin/ffprobe "$BUILD_DIR/opt/yted/ffprobe"
    chmod +x "$BUILD_DIR/opt/yted/ffprobe"
fi

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
