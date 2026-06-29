# File Converter

Video/audio converter built with Rust, powered by ffmpeg. CLI + GUI. Like Handbrake, lighter.

## Features

- **CLI + GUI** — headless batch (`file-converter-cli`) or visual interface (`file-converter-gui`)
- **10 built-in presets** — HandBrake-style quality tiers + GPU presets
- **GPU acceleration** — NVENC/AMD/QSV/VideoToolbox (dedicated presets + global GPU override)
- **Concurrent encodes** — N parallel ffmpeg processes (configurable)
- **Crash recovery** — queue persists to disk, resumes on restart
- **Cross-platform** — Windows, macOS, Linux (pure Rust, no CGO)

## Install

### Prerequisites

- **ffmpeg** — [Download](https://ffmpeg.org/download.html) and add to PATH

### Build from source

```bash
git clone https://github.com/your-username/file_converter.git
cd file_converter

# CLI (default binary)
cargo build

# GUI (egui/eframe)
cargo build -p file-converter-gui
```

## Usage

### CLI

```bash
# Convert single file
cargo run -- input.mp4 --preset "Fast 1080p30"

# Convert multiple files
cargo run -- *.mp4 --preset "H.265 1080p" --output ./converted

# Use GPU encoding
cargo run -- video.mp4 --preset "NVENC 1080p"

# Global GPU override for CPU presets
cargo run -- input.mp4 --hwaccel nvenc

# Queue mode (headless batch from JSON)
echo '[{"input_path":"a.mp4","preset_name":"Fast 1080p30"}]' | cargo run -- --queue

# Custom ffmpeg path
cargo run -- input.mp4 --ffmpeg-path /usr/local/bin/ffmpeg
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--preset` | `Fast 1080p30` | Preset name |
| `--output` | `./output` | Output directory |
| `--ffmpeg-path` | `""` | ffmpeg binary path (auto-detect if empty) |
| `--hwaccel` | `""` | Global GPU backend override |
| `--concurrent` | `2` | Concurrent encode jobs |
| `--queue` | `false` | Process queue JSON from stdin |
| `--keep-format` | `false` | Keep original container extension |
| `--bitrate` | `""` | Video bitrate (e.g. `2M`, overrides quality) |

### GUI

```bash
# Run the GUI binary
cargo run -p file-converter-gui
```

1. **Add Files** — click "Add Files" to select media files
2. **Select Preset** — pick from dropdown (Fast 1080p30, H.265 1080p, NVENC 1080p, etc.)
3. **Start** — click "Start" to begin encoding
4. **Monitor** — progress bars per job in queue pane

Settings: ffmpeg path, GPU backend, output directory, concurrent jobs.

## Presets

| Name | Container | Video | Audio | Quality | Notes |
|---|---|---|---|---|---|
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

**Global GPU override:** Set `--hwaccel` flag (CLI) or pick HW Accel (Settings) to override CPU presets at runtime.

## Configuration

Settings stored in platform config directory:

- **Windows:** `%APPDATA%\file_converter\settings.json`
- **macOS:** `~/Library/Application Support/file_converter/settings.json`
- **Linux:** `~/.config/file_converter/settings.json`

```json
{
  "ffmpeg_path": "",
  "hwaccel": "",
  "output_dir": "~/Videos",
  "default_preset": "Fast 1080p30",
  "concurrent_jobs": 2
}
```

Queue file (`queue.json`) lives in same directory for crash recovery.

## Project Structure

```
file_converter/
├── Cargo.toml              # Workspace root (core, cli, gui)
├── core/                   # Library crate
│   └── src/
│       ├── types.rs        # Preset, QueueItem, Progress structs
│       ├── ffmpeg.rs       # Binary resolution (PATH, user path)
│       ├── preset.rs       # 10 HandBrake-style presets
│       ├── convert.rs      # Ffmpeg arg builder + progress parsing
│       ├── queue.rs        # Job queue with JSON persistence
│       ├── runner.rs       # Ffmpeg execution with progress
│       ├── worker.rs       # Concurrent worker pool (mpsc)
│       └── config.rs       # User settings save/load
├── cli/                    # CLI binary crate (clap)
│   └── src/main.rs
└── gui/                    # GUI binary crate (egui/eframe)
    └── src/
        ├── main.rs
        ├── app.rs          # Eframe app, panel wiring
        ├── source_panel.rs # File picker + file list
        ├── preset_panel.rs # Preset selector
        ├── queue_panel.rs  # Job list + progress bars
        └── settings.rs     # Settings form dialog
```

## Development

```bash
# Run tests
cargo test

# Build all
cargo build --workspace

# CLI
cargo run -- --help

# GUI
cargo run -p file-converter-gui
```

## License

MIT
