#!/bin/bash
# Bundle FFmpeg binaries for the current platform

set -e

PLATFORM=$(uname -s)
mkdir -p build/bin

if [ "$PLATFORM" = "Linux" ]; then
    if [ -f build/bin/ffmpeg ]; then
        echo "FFmpeg already bundled"
        exit 0
    fi
    echo "Downloading FFmpeg for Linux..."
    curl -sL -o /tmp/ffmpeg-linux.tar.xz https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz
    tar -xf /tmp/ffmpeg-linux.tar.xz -C /tmp
    cp /tmp/ffmpeg-master-latest-linux64-gpl/bin/ffmpeg build/bin/ffmpeg
    cp /tmp/ffmpeg-master-latest-linux64-gpl/bin/ffprobe build/bin/ffprobe
    chmod +x build/bin/ffmpeg build/bin/ffprobe
    rm -rf /tmp/ffmpeg-linux.tar.xz /tmp/ffmpeg-master-latest-linux64-gpl
    echo "FFmpeg bundled for Linux"
else
    echo "Unsupported platform: $PLATFORM. Please bundle FFmpeg manually."
    exit 1
fi
