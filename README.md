# <img src="./build/appicon.png" alt="YTed Logo" width="32"> YTed

A modern, user-friendly YouTube downloader and library manager built with Go, Wails, and React.

[![Version](https://img.shields.io/badge/version-1.4.1-blue.svg)](https://github.com/beslintony/yted/releases)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- **Download Queue** - Manage multiple downloads with pause/resume/retry
- **Video Library** - Browse, search, and organize downloaded videos
- **Watch Progress** - Automatically track and resume playback position
- **Customizable** - Download presets, speed limits, themes, and more
- **Cross-Platform** - Native builds for Windows and Linux

## Installation

Download the latest release from the [Releases](https://github.com/beslintony/yted/releases) page.

### Windows

**Installer** (recommended):
```bash
# Run the installer
YTed-amd64-installer.exe
```

**Portable**:
```bash
# Download YTed.exe and run directly
# Requires FFmpeg installed separately
```

### Linux

**Debian/Ubuntu** (recommended):
```bash
sudo dpkg -i yted_1.4.1_amd64.deb
sudo apt-get install -f  # fix dependencies if needed
```

**Portable**:
```bash
tar xzf YTed-linux-amd64.tar.gz
./YTed
```

> **Note:** FFmpeg is required for video/audio merging. Install it via your package manager or download from [ffmpeg.org](https://ffmpeg.org/download.html).

## Development

### Prerequisites

- Go 1.25+
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Linux: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`

### Quick Start

```bash
# Clone and setup
git clone https://github.com/beslintony/yted.git
cd yted
make deps

# Development
make dev          # Run in dev mode
make test         # Run all tests
make lint         # Run linters

# Build
make build        # Build binary
make build-versioned VERSION=1.4.1

# Installers
make build-installer-linux    # Build .deb package
make build-installer-windows  # Build Windows installer
```

### Project Structure

```
yted/
├── frontend/        # React + Vite + TypeScript
├── internal/        # Go backend
│   ├── app/        # App logic & downloads
│   ├── config/     # Settings management
│   ├── db/         # SQLite database
│   ├── log/        # Structured logging
│   └── ytdl/       # yt-dlp client
├── build/          # Build assets & installers
└── main.go         # Entry point
```

## Configuration

Config is stored in `~/.yted/config/settings.json`:

| Setting | Description | Default |
|---------|-------------|---------|
| `download_path` | Download directory | `~/Downloads/YTed` |
| `max_concurrent_downloads` | Parallel downloads | `3` |
| `default_quality` | Default quality | `best` |
| `speed_limit_kbps` | Speed limit (0 = unlimited) | `0` |
| `theme` | UI theme | `dark` |
| `proxy_url` | HTTP/SOCKS proxy | - |

## Contributing

1. Fork the repository
2. Create a branch: `git checkout -b feat/my-feature`
3. Commit with [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` - New features
   - `fix:` - Bug fixes
   - `docs:` - Documentation
   - `chore:` - Build/config changes
4. Push and submit a PR

Before committing:
```bash
make fmt && make lint && make test
```

## Tech Stack

**Frontend:** React 18, TypeScript, Vite, Mantine UI, Zustand  
**Backend:** Go 1.25, Wails v2, go-ytdlp, SQLite  
**Tools:** ESLint, Vitest, golangci-lint

## License

[MIT](LICENSE) - YTed Contributors

**Third-Party:** This software uses [FFmpeg](https://ffmpeg.org/) as an external dependency. Users are responsible for complying with FFmpeg's license terms. See [LICENSE-THIRD-PARTY](LICENSE-THIRD-PARTY) for details.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube downloader
- [go-ytdlp](https://github.com/lrstanley/go-ytdlp) - Go bindings
- [Wails](https://wails.io/) - Desktop framework
- [Mantine](https://mantine.dev/) - React components

---

**YTed** - Download and enjoy YouTube videos offline, your way.
