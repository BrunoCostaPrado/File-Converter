# File Converter — Design Spec

Media converter (video/audio) using ffmpeg backend. CLI + GUI. Like Handbrake.

## Stack

- **Language:** Go
- **GUI:** Fyne (native cross-platform widgets)
- **Backend:** `os/exec` calling ffmpeg
- **Platform:** Windows / macOS / Linux
- **Ffmpeg:** Optional bundled binary per platform, fallback to system PATH

## Project Structure

```
file_converter/
├── main.go              # entry — CLI vs GUI detect
├── core/
│   ├── convert.go       # ffmpeg exec + progress parsing
│   ├── preset.go        # preset struct + built-in list
│   ├── ffmpeg.go        # binary resolution (bundled vs PATH)
│   └── queue.go         # job queue (JSON file)
├── gui/
│   ├── app.go           # fyne window + tabs
│   ├── source_panel.go  # file picker + list
│   ├── preset_panel.go  # codec/quality dropdown
│   ├── queue_panel.go   # job list + progress bars
│   └── settings.go      # prefs dialog
├── ffmpeg/              # bundled binaries per platform
├── go.mod
└── go.sum
```

3 packages: `main`, `core/`, `gui/`. CLI is thin wrapper via flags, shares `core/`.

## Ffmpeg Binary Resolution

1. Check `settings.ffmpegPath` (user-configured)
2. Check `file_converter/ffmpeg/<os>-<arch>/` (bundled)
3. Fallback `exec.LookPath("ffmpeg")` (system PATH)
4. Error dialog if none found

## Presets

```go
type Preset struct {
    Name        string
    Container   string   // mp4, mkv, webm
    VideoCodec  string   // h264, h265, vp9, copy
    AudioCodec  string   // aac, opus, mp3, copy
    Quality     int      // CRF 0-51 (video) or bitrate
    Preset      string   // ultrafast..veryslow (x264/x265)
    Resolution  string   // "" (source), "1920x1080", etc.
    HWAccel     string   // "", "nvenc", "qsv", "amd", "videotoolbox"
}
```

Built-in presets: "Fast 1080p", "Small 720p", "H265 Compact", "Lossless", "Audio Only", "Copy Stream". User saves custom to JSON.

HWAccel field maps to encoder names: `h264_nvenc`, `hevc_amdf`, `h264_qsv`, `h264_videotoolbox`, etc. Auto-detect available by parsing `ffmpeg -encoders`. Show only available GPU options in GUI dropdown.

## Conversion Engine

`core/convert.go`:
- Build ffmpeg args from Preset
- `exec.Command` with stdout/stderr pipes
- Parse stderr for `time=` to compute progress (ffmpeg's only progress channel)
- Emit progress via callback — GUI updates progress bar
- Threads: ffmpeg `-threads auto` per encode process

Concurrent jobs: N parallel ffmpeg processes (default `runtime.NumCPU()/2`, configurable 1-16). Each job tracks own progress. Cap at setting max.

## Queue

`core/queue.go`:
- JSON file stores `[]QueueItem`
- Fields: InputPath, OutputPath, PresetName, Status(pending/running/done/failed/cancelled), Progress, Error
- Queue processor goroutine spawns N workers from pool
- Each worker picks next pending item, runs ffmpeg, updates status

Crash recovery:
- On startup, re-mark `running` items as `pending`
- Skip `done` items
- Failed item stores stderr in Error field, queue continues

CLI `--queue` flag: reads JSON from stdin, processes headless, prints progress to stderr, writes results to stdout.

## CLI Interface

```
file_converter [input...] [flags]

Flags:
  --preset string       preset name (default "Fast 1080p")
  --output string       output dir (default "./output")
  --ffmpeg-path string  ffmpeg binary path
  --queue               process queue JSON from stdin
  --version             print version
```

Multiple inputs: `file_converter *.mp4 --preset "Small 720p"` — queues all sequentially.

## GUI Layout (Fyne)

Three-panel layout:
1. **Source panel** — file drag-drop area, file list with metadata (resolution, size, duration)
2. **Preset panel** — preset dropdown + summary, HWAccel toggle, output folder picker, "Add to Queue" button
3. **Queue panel** — job table with per-item progress bar, Start/Stop buttons, status badges

Queue panel shows per-encode progress with ETA. Each concurrent job gets own progress row.

ffmpeg not found → error dialog with file picker to locate binary. ffmpeg crash → stderr captured, job marked failed with error message. Disk full → pause queue, alert. Cancel → SIGTERM ffmpeg, mark cancelled.

## Settings

JSON file at platform config dir:
- Windows: `%APPDATA%\file_converter\settings.json`
- macOS: `~/Library/Application Support/file_converter/settings.json`
- Linux: `~/.config/file_converter/settings.json`

```json
{
  "ffmpeg_path": "",
  "theme": "system",
  "output_dir": "~/Videos",
  "default_preset": "Fast 1080p",
  "concurrent_jobs": 2
}
```

GUI settings panel: ffmpeg path (file picker), theme (light/dark/system), default preset, output dir, concurrent job count.

## Testing

- `core/` unit tests with `go test`. Mock ffmpeg binary that outputs fake progress.
- CLI integration: shell scripts testing `--queue` flag.
- GUI: manual (fyne test framework optional, not required for v1).

## Out of Scope (v1)

- Thumbnail previews (add when requested)
- GPU auto-detection UI polish (simple dropdown is fine)
- Watch folder / auto-convert
- Audio waveform / trimming UI
- Hardware encoding benchmarking
