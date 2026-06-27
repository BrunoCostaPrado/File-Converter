# File Converter

Video/audio converter built with Go and Fyne, powered by ffmpeg. CLI + GUI. Like Handbrake, lighter.

## Features

- **CLI + GUI** — headless batch processing or visual interface
- **6 built-in presets** — Fast 1080p, Small 720p, H265 Compact, Lossless, Audio Only, Copy Stream
- **GPU acceleration** — NVENC, QSV, AMD AMF, VideoToolbox (toggle per preset)
- **Concurrent encodes** — N parallel ffmpeg processes (configurable)
- **Crash recovery** — queue persists to disk, resumes on restart
- **Preset presets** — load/save custom presets as JSON
- **Cross-platform** — Windows, macOS, Linux

## Install

### Prerequisites

- **ffmpeg** — [Download](https://ffmpeg.org/download.html) and add to PATH, or place binary in `ffmpeg/<os>-<arch>/` next to the executable

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
file_converter input.mp4 --preset "Fast 1080p"

# Convert multiple files
file_converter *.mp4 --preset "H265 Compact" --output ./converted

# Specify output directory
file_converter video.mp4 --preset "Small 720p" --output ~/Videos

# Use GPU encoding
# Select a GPU-compatible preset in GUI or add HWAccel to custom preset

# Queue mode (headless batch from JSON)
echo '[{"InputPath":"a.mp4","PresetName":"Fast 1080p"}]' | file_converter --queue

# Custom ffmpeg path
file_converter input.mp4 --ffmpeg-path /usr/local/bin/ffmpeg
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--preset` | `Fast 1080p` | Preset name |
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
2. **Select Preset** — pick from dropdown (Fast 1080p, Small 720p, etc.)
3. **Start** — click "Start" to begin encoding
4. **Monitor** — progress bars per job in the queue pane

Settings (File → Settings): ffmpeg path, output directory, default preset, theme, concurrent jobs.

## Presets

| Name | Container | Video | Audio | Quality | Notes |
|---|---|---|---|---|---|
| Fast 1080p | mp4 | H.264 | AAC | CRF 23 | Good balance size/quality |
| Small 720p | mp4 | H.264 | AAC | CRF 28 | Smaller file, 720p |
| H265 Compact | mkv | H.265 | AAC | CRF 28 | ~50% smaller than H.264 |
| Lossless | mkv | H.264 | copy | CRF 0 | Visually lossless |
| Audio Only | m4a | copy | AAC | 192k | Strip video, transcode audio |
| Copy Stream | mkv | copy | copy | — | Remux, no re-encode |

**GPU override:** All video presets support hardware encoding. Set `HWAccel` to `nvenc`, `qsv`, `amd`, or `videotoolbox` in custom presets.

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
  "default_preset": "Fast 1080p",
  "concurrent_jobs": 2
}
```

Queue file (`queue.json`) lives in the same directory for crash recovery.

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
