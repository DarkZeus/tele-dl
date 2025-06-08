# tele-dl

A high-performance Go tool for downloading media from Telegraph (telegra.ph) pages.

## Features

- **Blazing Fast**: Concurrent downloads with configurable worker pools
- **Memory Efficient**: Streaming downloads without loading entire files into memory
- **Progress Tracking**: Real-time download progress and statistics
- **Retry Logic**: Automatic retry for failed downloads
- **JSON Output**: Machine-readable output format option

## Installation

```bash
# Clone the repository
git clone <your-repository-url>
cd tele-dl

# Build the binary
go build -o tele-dl main.go
```

## Usage

```bash
# Download all media from a Telegraph page
./tele-dl -l "https://telegra.ph/example-page-12-34"

# Use custom output directory and worker count
./tele-dl -l "https://telegra.ph/example-page-12-34" -o downloads -w 100

# Quiet mode with JSON output
./tele-dl -l "https://telegra.ph/example-page-12-34" -q --json

# Full options
./tele-dl --help
```

## Options

- `-l, --link`: Telegraph page URL (required)
- `-o, --output`: Output directory (default: current directory)
- `-w, --workers`: Number of concurrent downloads (default: 50)
- `-t, --timeout`: HTTP request timeout (default: 30s)
- `-p, --progress`: Show progress bar (default: true)
- `-q, --quiet`: Quiet mode (minimal output)
- `--retries`: Number of retry attempts (default: 3)
- `--json`: Output results in JSON format

## Requirements

- Go 1.21 or later
- Internet connection
