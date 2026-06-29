# File Converter

Video/audio converter built with Go + Fyne, powered by ffmpeg. CLI + GUI. Like Handbrake, lighter.

## Features

- **CLI + GUI** — headless batch or visual interface
- **10 built-in presets** — HandBrake-style quality tiers + GPU presets
- **GPU acceleration** — Dedicated NVENC/AMD presets (global GPU override for CPU presets)
- **Concurrent encodes** — N parallel ffmpeg processes (configurable)
- **Crash recovery** — queue persists to disk, resumes on restart
- **Preset presets** — load/save custom presets as JSON
- **Cross-platform** — Windows, macOS, Linux

## Install

### Prerequisites

- **ffmpeg** — [Download](https://ffmpeg.org/download.html) and add to PATH, or place binary in `ffmpeg/<os>-<arch>/` next to executable

### Binaries

Download from [Releases]() or build from source (see below).

### Build from source

```bash
git clone https://github.com/your-username/file_converter.git
cd file_converter
go build -o file_converter .
```

Requires Go 1.22+, gcc (CGO — Fyne dependency).

## Usage

### CLI

```bash
# Convert single file
file_converter input.mp4 --preset "Fast 1080p30"

# Convert multiple files
file_converter *.mp4 --preset "H.265 1080p" --output ./converted

# Specify output directory
file_converter video.mp4 --preset "Very Fast 720p" --output ~/Videos

# Use GPU encoding
# Select a GPU-compatible preset in GUI or add HWAccel to custom preset

# Queue mode (headless batch from JSON)
echo '[{"InputPath":"a.mp4","PresetName":"Fast 1080p30"}]' | file_converter --queue

# Custom ffmpeg path
file_converter input.mp4 --ffmpeg-path /usr/local/bin/ffmpeg
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--preset` | `Fast 1080p30` | Preset name |
| `--output` | `./output` | Output directory |
| `--ffmpeg-path` | `""` | ffmpeg binary path (auto-detect if empty) |
| `--concurrent` | `2` | Concurrent encode jobs |
| `--queue` | `false` | Process queue JSON from stdin |
| `--version` | `false` | Print version |

### GUI

Run without arguments:

```bash
file_converter
```

1. **Add Files** — click "Add Files" to select media files
2. **Select Preset** — pick from dropdown (Fast 1080p30, H.265 1080p, NVENC 1080p, etc.)
3. **Start** — click "Start" to begin encoding
4. **Monitor** — progress bars per job in queue pane

Settings (File → Settings): ffmpeg path, output directory, default preset, theme, concurrent jobs.

## Presets

| Name | Container | Video | Audio | Quality | Notes |
|---|---|---|---|---|---|---|
| Fast 1080p30 | mp4 | H.264 | AAC | RF 22 | HandBrake Fast — good balance |
| H.265 1080p | mkv | H.265 | AAC | RF 24 | Modern codec, ~50% smaller |
| Super HQ 1080p | mp4 | H.264 | AAC | RF 18 | Near-lossless source |
| Very Fast 720p | mp4 | H.264 | AAC | RF 22 | Quick encode, 720p target |
| Lossless | mkv | H.264 | copy | RF 0 | Visually lossless |
| Audio Only | m4a | copy | AAC | 192k | Strip video |
| Copy Stream | mkv | copy | copy | — | Remux only |
| NVENC 1080p | mp4 | H.264 | AAC | CQ 23 | GPU-accelerated, NVENC |
| NVENC H.265 1080p | mkv | H.265 | AAC | CQ 24 | GPU H.265, NVENC |
| AMD 1080p | mp4 | H.264 | AAC | 23 | GPU-accelerated, AMF |

**Global GPU override:** Set `--hwaccel` flag (CLI) or pick GPU Backend (Settings) to override CPU presets at runtime.

## Configuration

Settings stored in platform config directory:

- **Windows:** `%APPDATA%\file_converter\settings.json`
- **macOS:** `~/Library/Application Support/file_converter/settings.json`
- **Linux:** `~/.config/file_converter/settings.json`

```json
{
  "ffmpeg_path": "",
  "theme": "system",
  "output_dir": "~/Videos",
  "default_preset": "Fast 1080p30",
  "concurrent_jobs": 2
}
```

Queue file (`queue.json`) lives in same directory for crash recovery.

## Project Structure

```
file_converter/
├── main.go              # Entry — CLI vs GUI detect
├── core/
│   ├── types.go         # Preset, QueueItem, Progress structs
│   ├── ffmpeg.go        # Binary resolution (bundled, PATH, user)
│   ├── preset.go        # Default presets + load/save
│   ├── convert.go       # Ffmpeg arg builder + progress parsing
│   ├── queue.go         # Job queue with persistence
│   ├── runner.go        # Ffmpeg execution with progress
│   ├── worker.go        # Concurrent worker pool
│   └── settings.go      # User settings save/load
├── gui/
│   ├── app.go           # Fyne window, menu, panel wiring
│   ├── source_panel.go  # File picker + file list
│   ├── preset_panel.go  # Preset selector + summary
│   ├── queue_panel.go   # Job list + progress bars
│   └── settings.go      # Settings form dialog
└── ffmpeg/              # Bundled binaries (platform dirs)
```

## Development

```bash
# Run tests
go test ./core/ -v

# Build
go build -o file_converter .

# Run
./file_converter
```

## License

MIT
