# <img src="./build/appicon.png" alt="YTed Logo" width="32"> YTed

A modern, user-friendly YouTube downloader and library manager built with Go, Wails, and React.

[![Version](https://img.shields.io/badge/version-1.5.0-blue.svg)](https://github.com/beslintony/yted/releases)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- **Download Queue** - Manage multiple downloads with pause/resume/retry
- **Playlist Downloads** - Queue entire YouTube playlists in one click
- **Video Library** - Browse, search, and organize downloaded videos
- **Video Editor** - Trim, convert, watermark, and adjust downloaded videos
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
> **Linux runtime libs:** the binary (tarball/manual builds) needs GTK + WebKit at *runtime*, not just at build time. On Debian/Ubuntu: `sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0` (the `.deb` installer pulls these in automatically). If the app won't start, run `ldd ./build/bin/yted | grep "not found"` — any line listed there is a missing library. A graphical session (`$DISPLAY` on X11, or Wayland) is also required; headless/SSH sessions need a display (e.g. `xvfb-run`) to launch the window.
> For full YouTube extraction quality, a JS runtime is recommended — YTed auto-detects `deno`, `node`, or `bun` on your PATH (e.g. `sudo apt install nodejs` or install [deno](https://deno.land)). Downloads still work without one, but some formats may be unavailable.

## Development

### Prerequisites

- Go 1.25+
- Node.js 24+
- pnpm 11+ (via `corepack enable pnpm`)
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
│   ├── editor/     # Video editing (crop, convert, watermark, effects)
│   ├── log/        # Structured logging
│   ├── version/    # Build version info
│   └── ytdl/       # yt-dlp client
├── build/          # Build assets & installers
└── main.go         # Entry point
```

### Troubleshooting

**`T3-Code.AppImage: symbol lookup error: ... libpthread.so.0: undefined symbol: __libc_pthread_init`** when running `./build/bin/yted`:
The AppImage (and snap `core20`) exports a stale `LD_LIBRARY_PATH` (e.g. `/snap/core20/...` or `/tmp/.mount_T3-.../usr/lib` with glibc 2.31) that poisons every child process. Host binaries built on glibc 2.43 then load the old `libpthread`/`libc` and crash. This is not a YTed bug — it's an AppImage/snap env leak.
**Fixed in this repo:** `make build` now wraps the binary — `build/bin/yted` is a static Go wrapper that clears `LD_LIBRARY_PATH`/`LD_PRELOAD` before execing `build/bin/yted.bin` (the real binary). So `./build/bin/yted` and `make run` work even from a poisoned shell. If you invoke `make` itself with a poisoned env and `make` crashes (`make: ... libc.so.6: version GLIBC_2.33 not found`), run `unset LD_LIBRARY_PATH` (shell builtin) first, or `env -u LD_LIBRARY_PATH make run`. For a permanent fix, launch the AppImage clean: `env -u LD_LIBRARY_PATH /home/bax/.local/bin/T3-Code.AppImage` — then child shells start clean.

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
