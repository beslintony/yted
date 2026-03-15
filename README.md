# YTed

A modern, user-friendly YouTube downloader and editor built with Go, Wails, and React.

![YTed Logo](./frontend/public/logo.svg)

## Features

- **Modern UI**: Built with Mantine UI components, featuring a clean dark theme
- **Download Queue**: Manage multiple downloads with pause/resume/retry support
- **Video Library**: Organized view of downloaded videos with search and filter
- **Configurable**: Extensive user settings including download presets
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Tech Stack

### Frontend
- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Mantine v7** - Component library
- **Zustand** - State management
- **Tabler Icons** - Icon library

### Backend
- **Go 1.21+** - Backend language
- **Wails v2** - Desktop framework
- **go-ytdlp** - yt-dlp bindings
- **modernc.org/sqlite** - Pure Go SQLite

## Installation

### Prerequisites
- Go 1.21 or later
- Node.js 18 or later
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build from Source

```bash
# Clone the repository
git clone https://github.com/beslintony/yted.git
cd yted

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run in development mode
wails dev

# Build for production
wails build
```

### Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/beslintony/yted/releases) page.

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
│   │   └── tests/         # Unit tests
│   ├── public/            # Static assets
│   └── package.json
├── internal/              # Go backend
│   ├── app/              # Wails app handlers
│   ├── config/           # Configuration management
│   ├── db/               # SQLite database
│   └── ytdl/             # yt-dlp client
├── build/                # Build assets
└── main.go              # Entry point
```

### Frontend Development

```bash
cd frontend
npm run dev       # Start Vite dev server
npm run test      # Run tests
npm run build     # Build for production
```

### Backend Development

```bash
# Generate Wails bindings
wails generate module

# Build Go binary
wails build
```

### Running Tests

```bash
# Frontend tests
cd frontend
npm test

# Backend tests (if any)
go test ./...
```

## Configuration

YTed stores configuration in `~/.yted/config/settings.json`. The following settings are available:

### Downloads
- `download_path` - Where downloaded videos are saved
- `max_concurrent_downloads` - Number of simultaneous downloads (1-10)
- `default_quality` - Default quality preference (best, 1080p, 720p, 480p, audio)
- `filename_template` - Template for output filenames

### UI
- `theme` - Color theme (dark, light, auto)
- `accent_color` - Primary accent color
- `sidebar_collapsed` - Whether sidebar is collapsed

### Network
- `speed_limit_kbps` - Download speed limit in KB/s
- `proxy_url` - HTTP/SOCKS proxy URL

## Download Presets

YTed comes with default download presets:
- **Best Quality** - Best available quality
- **1080p** - 1080p video + audio
- **720p** - 720p video + audio  
- **Audio Only** - Audio only (MP3)

Users can create custom presets with specific formats and qualities.

## Database Schema

YTed uses SQLite to store:

### Videos Table
- Video metadata (title, channel, duration, etc.)
- File information (path, size, format)
- Watch progress and count

### Downloads Table
- Download queue with status tracking
- Progress information
- Error messages

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

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - The amazing YouTube downloader
- [go-ytdlp](https://github.com/lrstanley/go-ytdlp) - Go bindings for yt-dlp
- [Wails](https://wails.io/) - Build desktop apps with Go and web technologies
- [Mantine](https://mantine.dev/) - React components library

## Support

For issues, questions, or contributions, please use [GitHub Issues](https://github.com/beslintony/yted/issues).

---

**YTed** - Download and enjoy YouTube videos offline, your way.
