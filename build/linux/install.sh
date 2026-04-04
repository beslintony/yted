#!/bin/bash
# YTed Linux Installation Script

set -e

APP_NAME="YTed"
EXECUTABLE="yted"
ICON_NAME="yted"
DESKTOP_FILE="yted.desktop"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Ensure binary exists
if [ ! -f "$PROJECT_ROOT/build/bin/yted" ]; then
    echo "Error: yted binary not found at $PROJECT_ROOT/build/bin/yted"
    echo "Please build first with: make build"
    exit 1
fi

# Check for FFmpeg
if ! command -v ffmpeg &> /dev/null; then
    echo ""
    echo "WARNING: FFmpeg not found on your system."
    echo "YTed requires FFmpeg to download videos."
    echo ""
    echo "To install FFmpeg:"
    echo "  Ubuntu/Debian: sudo apt install ffmpeg"
    echo "  Fedora:        sudo dnf install ffmpeg"
    echo "  Arch:          sudo pacman -S ffmpeg"
    echo ""
    read -p "Continue with installation anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Determine installation type
if [ "$1" == "--system" ]; then
    INSTALL_DIR="/usr/local/bin"
    ICON_DIR="/usr/share/pixmaps"
    DESKTOP_DIR="/usr/share/applications"
    echo "Installing YTed system-wide (requires sudo)..."
    SUDO="sudo"
else
    INSTALL_DIR="$HOME/.local/bin"
    ICON_DIR="$HOME/.local/share/icons"
    DESKTOP_DIR="$HOME/.local/share/applications"
    echo "Installing YTed for current user..."
    SUDO=""
    
    # Create directories if they don't exist
    mkdir -p "$INSTALL_DIR" "$ICON_DIR" "$DESKTOP_DIR"
fi

# Install binary
echo "Installing binary to $INSTALL_DIR..."
$SUDO cp "$PROJECT_ROOT/build/bin/yted" "$INSTALL_DIR/$EXECUTABLE"
$SUDO chmod +x "$INSTALL_DIR/$EXECUTABLE"

# Install icon
echo "Installing icon to $ICON_DIR..."
$SUDO cp "$PROJECT_ROOT/build/appicon.png" "$ICON_DIR/$ICON_NAME.png"

# Install desktop file
echo "Installing desktop entry to $DESKTOP_DIR..."
if [ "$SUDO" == "sudo" ]; then
    $SUDO cp "$SCRIPT_DIR/$DESKTOP_FILE" "$DESKTOP_DIR/"
else
    # Update desktop file for user install (replace Exec and Icon paths)
    sed -e "s|Exec=yted|Exec=$INSTALL_DIR/yted|" \
        -e "s|Icon=yted|Icon=$ICON_DIR/yted.png|" "$SCRIPT_DIR/$DESKTOP_FILE" > "$DESKTOP_DIR/$DESKTOP_FILE"
fi

# Install license files
echo "Installing license files..."
if [ "$SUDO" == "sudo" ]; then
    $SUDO mkdir -p "/usr/share/doc/yted"
    $SUDO cp "$PROJECT_ROOT/LICENSE" "/usr/share/doc/yted/"
    $SUDO cp "$PROJECT_ROOT/LICENSE-THIRD-PARTY" "/usr/share/doc/yted/"
else
    mkdir -p "$HOME/.local/share/doc/yted"
    cp "$PROJECT_ROOT/LICENSE" "$HOME/.local/share/doc/yted/"
    cp "$PROJECT_ROOT/LICENSE-THIRD-PARTY" "$HOME/.local/share/doc/yted/"
fi

# Update desktop database
if command -v update-desktop-database &> /dev/null; then
    echo "Updating desktop database..."
    $SUDO update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
fi

echo ""
echo "YTed has been installed successfully!"
echo ""
if [ "$SUDO" == "sudo" ]; then
    echo "You can now launch YTed from your applications menu or by running 'yted'"
else
    echo "Make sure ~/.local/bin is in your PATH"
    echo "You can now launch YTed from your applications menu or by running '$INSTALL_DIR/yted'"
fi

# Final FFmpeg reminder
if ! command -v ffmpeg &> /dev/null; then
    echo ""
    echo "NOTE: FFmpeg is required but not installed. Please install it:"
    echo "  Ubuntu/Debian: sudo apt install ffmpeg"
    echo "  Fedora:        sudo dnf install ffmpeg"
    echo "  Arch:          sudo pacman -S ffmpeg"
fi
