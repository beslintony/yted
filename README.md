# <img src="./build/appicon.png" alt="YTed Logo" width="32"> YTed

A modern, user-friendly YouTube downloader and library manager built with Go, Wails, and React.

[![Version](https://img.shields.io/badge/version-1.4.1-blue.svg)](https://github.com/beslintony/yted/releases)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- **Modern UI**: Built with Mantine UI v8, featuring a clean dark theme
- **Download Queue**: Manage multiple downloads with pause/resume/retry support
- **Video Library**: Organized view of downloaded videos with search, filter, and sorting
- **Watch Progress**: Automatically track and resume playback position
- **Configurable**: Extensive user settings including download presets and speed limits
- **Cross-Platform**: Native builds for Windows and Linux with official installers
- **FFmpeg Integration**: Works with your existing FFmpeg installation or helps you set it up
- **Custom App Icon**: Branded YTed icon across all platforms

## Screenshots

*Screenshots coming soon*

## Tech Stack

### Frontend
- **React 18.3** - UI library
- **TypeScript 5.9** - Type safety
- **Vite 8** - Build tool
- **Mantine v8** - Component library
- **Zustand 5** - State management
- **Tabler Icons** - Icon library
- **ESLint 9** - Linting (flat config)
- **Vitest** - Testing framework

### Backend
- **Go 1.25+** - Backend language
- **Wails v2** - Desktop framework
- **go-ytdlp** - yt-dlp bindings
- **modernc.org/sqlite** - Pure Go SQLite

## Installation

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/beslintony/yted/releases) page.

#### Windows
- **Installer** (recommended): Download `YTed-amd64-installer.exe` and run it. Creates Start Menu and Desktop shortcuts.
- **Portable**: Download `YTed.exe` and place it anywhere. Requires FFmpeg to be installed separately.

#### Linux
- **Ubuntu/Debian** (recommended): Download the `.deb` package and install it:
  ```bash
  sudo dpkg -i yted_x.x.x_amd64.deb
  sudo apt-get install -f  # fix dependencies if needed
  ```
- **Portable**: Download `YTed-linux-amd64.tar.gz`, extract it, and run `./YTed`:
  ```bash
  tar xzf YTed-linux-amd64.tar.gz
  ./YTed
  ```

### Build from Source

#### Prerequisites
- Go 1.25 or later
- Node.js 20 or later
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Linux only: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`
- Windows installer only: [NSIS](https://nsis.sourceforge.io/Download)

#### Steps

```bash
# Clone the repository
git clone https://github.com/beslintony/yted.git
cd yted

# Install dependencies and build
make deps
make build

# Or build with version info
make build-versioned VERSION=1.4.1
```



### Install from Source (Linux)

```bash
# Install for current user (~/.local/)
make install

# Or install system-wide (/usr/local/)
make install-system

# Uninstall
make uninstall
make uninstall-system
```

### Build Installers from Source

```bash
# Linux .deb package
make build-installer-linux

# Windows NSIS installer (run on Windows or with NSIS installed)
make build-installer-windows
```

## Development

### Project Structure

```
yted/
├── frontend/              # React + Vite frontend
│   ├── src/
│   │   ├── components/    # UI components
│   │   ├── pages/         # Page components
│   │   ├── stores/        # Zustand stores
│   │   ├── types/         # TypeScript types
│   │   └── tests/         # Unit tests (75 tests)
│   ├── public/            # Static assets
│   └── package.json
├── internal/              # Go backend
│   ├── app/              # Wails app handlers
│   ├── config/           # Configuration management
│   ├── db/               # SQLite database
│   ├── log/              # Logging system
│   ├── version/          # Version management
│   └── ytdl/             # yt-dlp client
├── build/                # Build assets, icons, scripts & installers
│   ├── bin/              # Build output
│   ├── linux/            # Linux install scripts and desktop files
│   ├── scripts/          # Helper scripts (build .deb, etc)
│   └── windows/          # Windows icons and NSIS installer files
└── main.go              # Entry point
```

### Makefile Commands

```bash
make dev                   # Run in development mode
make build                 # Build for production
make build-versioned       # Build with version injection
make build-installer-linux    # Build Linux .deb package
make build-installer-windows  # Build Windows NSIS installer
make test                  # Run all tests
make lint                  # Run linters
make fmt                   # Format code
make install               # Install to ~/.local/
make install-system        # Install system-wide
make clean                 # Clean build artifacts
make help                  # Show all commands
```

### Frontend Development

```bash
cd frontend
npm run dev       # Start Vite dev server
npm run test      # Run tests (Vitest)
npm run build     # Build for production
npm run lint      # Run ESLint
npm run format    # Format with Prettier
```

### Backend Development

```bash
# Generate Wails bindings
wails generate module

# Run Go tests
go test -v ./...

# Build Go binary
wails build
```

## Configuration

YTed stores configuration in `~/.yted/config/settings.json`. The following settings are available:

### Downloads
- `download_path` - Where downloaded videos are saved (default: ~/Downloads/YTed)
- `max_concurrent_downloads` - Number of simultaneous downloads (1-10)
- `default_quality` - Default quality preference (best, 2160p, 1440p, 1080p, 720p, 480p, 360p, audio)
- `filename_template` - Template for output filenames
- `speed_limit_kbps` - Download speed limit in KB/s

### UI
- `theme` - Color theme (dark, light, auto)
- `accent_color` - Primary accent color (default: YouTube red #ff0000)
- `sidebar_collapsed` - Whether sidebar is collapsed

### Network
- `proxy_url` - HTTP/SOCKS proxy URL

### Player
- `default_volume` - Default player volume (0-100)
- `remember_position` - Remember playback position

## Download Presets

YTed comes with default download presets:
- **4K (2160p)** - Ultra HD quality
- **1440p (2K)** - Quad HD quality
- **1080p HD** - Full HD
- **720p HD** - Standard HD
- **480p** - DVD quality
- **Best Available** - Highest quality available
- **Audio Only (MP3)** - MP3 audio extraction
- **Audio Only (M4A)** - M4A audio extraction

Users can create custom presets with specific formats and qualities.

## Database Schema

YTed uses SQLite to store:

### Videos Table
- Video metadata (title, channel, duration, description, thumbnail)
- File information (path, size, format, quality)
- Watch progress and count
- Download timestamp

### Downloads Table
- Download queue with status tracking (pending, downloading, paused, completed, error)
- Progress information (percentage, speed, ETA)
- Error messages and retry count

## Architecture Highlights

### Performance Optimizations
- **Video Info Caching**: 5-minute TTL cache for YouTube API calls
- **Progress Debouncing**: Throttled progress events (max 2/sec) to prevent UI flooding
- **SQLite**: Fast local database for metadata and download queue

### Cross-Platform Support
- **Linux**: Native GTK3/WebKit2GTK with `.desktop` integration and `.deb` package
- **Windows**: Native build with custom icon and NSIS installer

### Sandbox Compatibility
- File/folder opening with fallback to browser for snap/AppImage
- Native commands preferred, BrowserOpenURL as fallback

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -am 'feat: add new feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Submit a pull request

### Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test updates
- `chore:` - Build/config changes

### Code Quality

```bash
# Before committing, run:
make fmt    # Format code
make lint   # Check linting
make test   # Run tests
```

## License

MIT License - see [LICENSE](LICENSE) for details.

### Third-Party Licenses

This software uses [FFmpeg](https://ffmpeg.org/) as an external dependency. 
FFmpeg is a trademark of Fabrice Bellard, originator of the FFmpeg project.

See [LICENSE-THIRD-PARTY](LICENSE-THIRD-PARTY) for FFmpeg license information.

Users are responsible for complying with FFmpeg's license terms when installing
and using FFmpeg with this software.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - The amazing YouTube downloader
- [go-ytdlp](https://github.com/lrstanley/go-ytdlp) - Go bindings for yt-dlp
- [FFmpeg](https://ffmpeg.org/) - Complete multimedia solution
- [Wails](https://wails.io/) - Build desktop apps with Go and web technologies
- [Mantine](https://mantine.dev/) - React components library
- [Vite](https://vitejs.dev/) - Next generation frontend tooling

## Support

For issues, questions, or contributions, please use [GitHub Issues](https://github.com/beslintony/yted/issues).

---

**YTed** - Download and enjoy YouTube videos offline, your way.
