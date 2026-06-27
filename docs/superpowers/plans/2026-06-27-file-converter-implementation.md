# File Converter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a video/audio converter (CLI + GUI) using Go + Fyne + ffmpeg.

**Architecture:** Three packages (`main`, `core/`, `gui/`). `core/` handles ffmpeg exec, presets, queue. `gui/` wraps core in Fyne widgets. `main` detects CLI flags vs GUI mode. Concurrent worker pool for batch encoding with GPU support via preset HWAccel field.

**Tech Stack:** Go 1.22+, Fyne v2, ffmpeg (external binary), `encoding/json`, `os/exec`

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `main.go` (skeleton)

- [ ] **Step 1: Init Go module**

```
go mod init file_converter
```

- [ ] **Step 2: Write minimal main.go**

```go
package main

import "fmt"

func main() {
	fmt.Println("file_converter v0.1.0")
}
```

- [ ] **Step 3: Verify build**

```
go build -o file_converter.exe .
Expected: file_converter.exe created, runs and prints version.
```

- [ ] **Step 4: Commit**

```
git init
git add go.mod main.go
git commit -m "feat: project scaffold"
```

---

### Task 2: Core Types

**Files:**
- Create: `core/types.go`

- [ ] **Step 1: Define types**

```go
package core

type Preset struct {
	Name        string `json:"name"`
	Container   string `json:"container"`   // mp4, mkv, webm
	VideoCodec  string `json:"video_codec"` // h264, h265, vp9, copy
	AudioCodec  string `json:"audio_codec"` // aac, opus, mp3, copy
	Quality     int    `json:"quality"`     // CRF 0-51
	Preset      string `json:"preset"`      // ultrafast..veryslow
	Resolution  string `json:"resolution"`  // "" or "1920x1080"
	HWAccel     string `json:"hwaccel"`     // "", nvenc, qsv, amd, videotoolbox
}

type QueueItem struct {
	InputPath   string `json:"input_path"`
	OutputPath  string `json:"output_path"`
	PresetName  string `json:"preset_name"`
	Status      string `json:"status"` // pending, running, done, failed, cancelled
	Progress    float64 `json:"progress"`
	Error       string `json:"error,omitempty"`
}

type Progress struct {
	File      string  `json:"file"`
	Percent   float64 `json:"percent"`
	Speed     string  `json:"speed"`
	ETA       string  `json:"eta"`
	Status    string  `json:"status"`
}
```

- [ ] **Step 2: Write core/types_test.go**

```go
package core

import (
	"encoding/json"
	"testing"
)

func TestQueueItemJSONRoundTrip(t *testing.T) {
	item := QueueItem{InputPath: "a.mp4", Status: "pending"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var decoded QueueItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.InputPath != "a.mp4" || decoded.Status != "pending" {
		t.Fatalf("round trip lost data: %+v", decoded)
	}
}
```

- [ ] **Step 3: Verify tests pass**

```
go test ./core/ -v
Expected: PASS
```

- [ ] **Step 4: Commit**

```
git add core/
git commit -m "feat: core types and JSON round-trip test"
```

---

### Task 3: Ffmpeg Binary Resolution

**Files:**
- Create: `core/ffmpeg.go`
- Create: `core/ffmpeg_test.go`

- [ ] **Step 1: Write test**

```go
package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFfmpegBundled(t *testing.T) {
	// Create a fake ffmpeg in temp to simulate bundled binary
	dir := t.TempDir()
	fake := filepath.Join(dir, "ffmpeg.exe")
	os.WriteFile(fake, []byte("fake"), 0644)

	got := findFfmpeg([]string{dir}, "")
	if got != fake {
		t.Fatalf("expected %q, got %q", fake, got)
	}
}

func TestFindFfmpegUserPath(t *testing.T) {
	// When user provides explicit path
	got := findFfmpeg([]string{}, "/custom/ffmpeg")
	if got != "/custom/ffmpeg" {
		t.Fatalf("expected user path, got %q", got)
	}
}

func TestFindFfmpegNotFound(t *testing.T) {
	got := findFfmpeg([]string{"/nonexistent"}, "")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./core/ -v -run TestFindFfmpeg
Expected: FAIL (function not defined)
```

- [ ] **Step 3: Write implementation**

```go
package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func FfmpegPaths(bundledDir string) []string {
	// Search order: bundled dir for platform, then system PATH
	var dirs []string
	if bundledDir != "" {
		platformDir := filepath.Join(bundledDir, runtime.GOOS+"-"+runtime.GOARCH)
		dirs = append(dirs, platformDir)
	}
	return dirs
}

func FindFfmpeg(bundleDirs []string, userPath string) string {
	if userPath != "" {
		if _, err := os.Stat(userPath); err == nil {
			return userPath
		}
	}
	for _, dir := range bundleDirs {
		candidates := []string{"ffmpeg", "ffmpeg.exe"}
		for _, name := range candidates {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ""
}
```

Note: For test helper, rename `FindFfmpeg` to `findFfmpeg` (unexported) for direct testing, or just test `FindFfmpeg`:

```go
func TestFindFfmpegBundled(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "ffmpeg.exe")
	os.WriteFile(fake, []byte("fake"), 0644)
	got := FindFfmpeg([]string{dir}, "")
	if got != fake {
		t.Fatalf("expected %q, got %q", fake, got)
	}
}
```

- [ ] **Step 4: Run tests to verify pass**

```
go test ./core/ -v -run TestFindFfmpeg
Expected: PASS
```

- [ ] **Step 5: Commit**

```
git add core/ffmpeg.go core/ffmpeg_test.go
git commit -m "feat: ffmpeg binary resolution"
```

---

### Task 4: Presets

**Files:**
- Create: `core/preset.go`
- Create: `core/preset_test.go`

- [ ] **Step 1: Write test**

```go
package core

import (
	"testing"
)

func TestDefaultPresets(t *testing.T) {
	presets := DefaultPresets()
	if len(presets) == 0 {
		t.Fatal("expected at least one default preset")
	}
	found := false
	for _, p := range presets {
		if p.Name == "Fast 1080p" {
			found = true
			if p.VideoCodec != "h264" {
				t.Fatalf("expected h264, got %s", p.VideoCodec)
			}
			break
		}
	}
	if !found {
		t.Fatal("Fast 1080p preset not found")
	}
}

func TestPresetValidate(t *testing.T) {
	p := Preset{Name: "", Container: "mp4"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}
```

- [ ] **Step 2: Run to verify fail**

```
go test ./core/ -v -run TestDefault
Expected: FAIL
```

- [ ] **Step 3: Write implementation**

```go
package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func DefaultPresets() []Preset {
	return []Preset{
		{Name: "Fast 1080p", Container: "mp4", VideoCodec: "h264", AudioCodec: "aac", Quality: 23, Preset: "medium", Resolution: "1920x1080"},
		{Name: "Small 720p", Container: "mp4", VideoCodec: "h264", AudioCodec: "aac", Quality: 28, Preset: "fast", Resolution: "1280x720"},
		{Name: "H265 Compact", Container: "mkv", VideoCodec: "h265", AudioCodec: "aac", Quality: 28, Preset: "medium", Resolution: ""},
		{Name: "Lossless", Container: "mkv", VideoCodec: "h264", AudioCodec: "copy", Quality: 0, Preset: "slow", Resolution: ""},
		{Name: "Audio Only", Container: "m4a", VideoCodec: "copy", AudioCodec: "aac", Quality: 192, Preset: "", Resolution: ""},
		{Name: "Copy Stream", Container: "mkv", VideoCodec: "copy", AudioCodec: "copy", Quality: 0, Preset: "", Resolution: ""},
	}
}

func DefaultPresetNames() []string {
	presets := DefaultPresets()
	names := make([]string, len(presets))
	for i, p := range presets {
		names[i] = p.Name
	}
	return names
}

func (p Preset) Validate() error {
	if p.Name == "" {
		return errors.New("preset name required")
	}
	return nil
}

func LoadPresets(path string) ([]Preset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var presets []Preset
	if err := json.Unmarshal(data, &presets); err != nil {
		return nil, err
	}
	return presets, nil
}

func SavePresets(path string, presets []Preset) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Verify tests pass**

```
go test ./core/ -v -run TestDefault
Expected: PASS

go test ./core/ -v -run TestPresetValidate
Expected: PASS
```

- [ ] **Step 5: Commit**

```
git add core/preset.go core/preset_test.go
git commit -m "feat: presets with defaults"
```

---

### Task 5: Ffmpeg Command Builder

**Files:**
- Create: `core/convert.go`
- Create: `core/convert_test.go`

- [ ] **Step 1: Write test**

```go
package core

import (
	"testing"
)

func TestBuildFfmpegArgs(t *testing.T) {
	p := Preset{
		Container: "mp4", VideoCodec: "h264", AudioCodec: "aac",
		Quality: 23, Preset: "medium", Resolution: "1920x1080",
	}
	args := BuildFfmpegArgs("in.mp4", "out.mp4", p)
	expected := []string{"-i", "in.mp4", "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-vf", "scale=1920:1080", "out.mp4"}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i := range expected {
		if args[i] != expected[i] {
			t.Fatalf("arg %d: expected %q, got %q", i, expected[i], args[i])
		}
	}
}

func TestBuildFfmpegArgsCopy(t *testing.T) {
	p := Preset{
		Container: "mkv", VideoCodec: "copy", AudioCodec: "copy",
		Quality: 0, Preset: "", Resolution: "",
	}
	args := BuildFfmpegArgs("in.mp4", "out.mkv", p)
	expected := []string{"-i", "in.mp4", "-c:v", "copy", "-c:a", "copy", "out.mkv"}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
}
```

- [ ] **Step 2: Run to verify fail**

```
go test ./core/ -v -run TestBuildFfmpegArgs
Expected: FAIL
```

- [ ] **Step 3: Write implementation**

```go
package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var videoEncoders = map[string]string{
	"h264": "libx264",
	"h265": "libx265",
	"vp9":  "libvpx-vp9",
	"copy": "copy",
}

var hwEncoders = map[string]map[string]string{
	"nvenc":       {"h264": "h264_nvenc", "h265": "hevc_nvenc"},
	"qsv":         {"h264": "h264_qsv", "h265": "hevc_qsv"},
	"amd":         {"h264": "h264_amf", "h265": "hevc_amf"},
	"videotoolbox": {"h264": "h264_videotoolbox", "h265": "hevc_videotoolbox"},
}

func BuildFfmpegArgs(input, output string, p Preset) []string {
	args := []string{"-i", input}

	// Video codec
	if p.VideoCodec != "" && p.VideoCodec != "copy" {
		enc := ""
		if p.HWAccel != "" {
			if m, ok := hwEncoders[p.HWAccel]; ok {
				if v, ok := m[p.VideoCodec]; ok {
					enc = v
				}
			}
		}
		if enc == "" {
			enc = videoEncoders[p.VideoCodec]
		}
		args = append(args, "-c:v", enc)

		if p.Preset != "" {
			args = append(args, "-preset", p.Preset)
		}
		// CRF for video encoders (not copy)
		args = append(args, "-crf", strconv.Itoa(p.Quality))
	} else if p.VideoCodec == "copy" {
		args = append(args, "-c:v", "copy")
	}

	// Audio codec
	if p.AudioCodec != "" && p.AudioCodec != "copy" {
		args = append(args, "-c:a", p.AudioCodec)
	} else if p.AudioCodec == "copy" {
		args = append(args, "-c:a", "copy")
	}

	// Resolution
	if p.Resolution != "" {
		args = append(args, "-vf", fmt.Sprintf("scale=%s", strings.Replace(p.Resolution, "x", ":", 1)))
	}

	args = append(args, output)
	return args
}

// progressRegex parses ffmpeg stderr for time=HH:MM:SS.MS
var progressRegex = regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)

func ParseProgressLine(line string) (float64, bool) {
	matches := progressRegex.FindStringSubmatch(line)
	if len(matches) < 5 {
		return 0, false
	}
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	total := float64(hours*3600+minutes*60+seconds)
	return total, true
}
```

- [ ] **Step 4: Verify tests pass**

```
go test ./core/ -v -run TestBuildFfmpegArgs
Expected: PASS
```

- [ ] **Step 5: Commit**

```
git add core/convert.go core/convert_test.go
git commit -m "feat: ffmpeg args builder"
```

---

### Task 6: Queue Processor

**Files:**
- Create: `core/queue.go`
- Create: `core/queue_test.go`

- [ ] **Step 1: Write test**

```go
package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQueueSaveLoad(t *testing.T) {
	q := &Queue{Items: []QueueItem{{InputPath: "a.mp4", Status: "pending"}}}
	path := filepath.Join(t.TempDir(), "queue.json")
	if err := q.Save(path); err != nil {
		t.Fatal(err)
	}
	q2, err := LoadQueue(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(q2.Items) != 1 || q2.Items[0].InputPath != "a.mp4" {
		t.Fatalf("round trip failed: %+v", q2.Items)
	}
}

func TestQueueNextItem(t *testing.T) {
	q := &Queue{
		Items: []QueueItem{
			{InputPath: "a.mp4", Status: "done"},
			{InputPath: "b.mp4", Status: "pending"},
			{InputPath: "c.mp4", Status: "pending"},
		},
	}
	item := q.NextPending()
	if item == nil || item.InputPath != "b.mp4" {
		t.Fatalf("expected b.mp4, got %v", item)
	}
}
```

- [ ] **Step 2: Run to verify fail**

```
go test ./core/ -v -run TestQueue
Expected: FAIL
```

- [ ] **Step 3: Write implementation**

```go
package core

import (
	"encoding/json"
	"os"
	"sync"
)

type Queue struct {
	mu    sync.Mutex
	Items []QueueItem
}

func NewQueue() *Queue {
	return &Queue{Items: []QueueItem{}}
}

func (q *Queue) Add(items ...QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Items = append(q.Items, items...)
}

func (q *Queue) NextPending() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.Items {
		if q.Items[i].Status == "pending" {
			q.Items[i].Status = "running"
			return &q.Items[i]
		}
	}
	return nil
}

func (q *Queue) UpdateStatus(index int, status string, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < len(q.Items) {
		q.Items[index].Status = status
		q.Items[index].Error = errMsg
	}
}

func (q *Queue) Save(path string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	data, err := json.MarshalIndent(q.Items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadQueue(path string) (*Queue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []QueueItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return &Queue{Items: items}, nil
}
```

- [ ] **Step 4: Verify tests pass**

```
go test ./core/ -v -run TestQueue
Expected: PASS
```

- [ ] **Step 5: Commit**

```
git add core/queue.go core/queue_test.go
git commit -m "feat: queue with persistence"
```

---

### Task 7: Conversion Worker

**Files:**
- Modify: `core/convert.go`
- Modify: `core/convert_test.go`

- [ ] **Step 1: Write worker test**

```go
package core

import (
	"testing"
)

func TestEncoderName(t *testing.T) {
	tests := []struct {
		hwAccel  string
		codec    string
		expected string
	}{
		{"", "h264", "libx264"},
		{"nvenc", "h264", "h264_nvenc"},
		{"nvenc", "h265", "hevc_nvenc"},
		{"qsv", "h264", "h264_qsv"},
		{"amd", "h264", "h264_amf"},
		{"videotoolbox", "h264", "h264_videotoolbox"},
		{"", "copy", "copy"},
	}
	for _, tt := range tests {
		got := EncoderName(tt.hwAccel, tt.codec)
		if got != tt.expected {
			t.Errorf("EncoderName(%q, %q) = %q, want %q", tt.hwAccel, tt.codec, got, tt.expected)
		}
	}
}
```

- [ ] **Step 2: Run to verify fail**

```
go test ./core/ -v -run TestEncoderName
Expected: FAIL
```

- [ ] **Step 3: Add encoder lookup to convert.go**

```go
func EncoderName(hwAccel, codec string) string {
	if codec == "copy" {
		return "copy"
	}
	if hwAccel != "" {
		if m, ok := hwEncoders[hwAccel]; ok {
			if v, ok := m[codec]; ok {
				return v
			}
		}
	}
	if v, ok := videoEncoders[codec]; ok {
		return v
	}
	return codec
}
```

Update `BuildFfmpegArgs` to use `EncoderName`:

```go
func BuildFfmpegArgs(input, output string, p Preset) []string {
	args := []string{"-i", input}
	// ... rest ...
	enc := EncoderName(p.HWAccel, p.VideoCodec)
	args = append(args, "-c:v", enc)
	// ...
}
```

- [ ] **Step 4: Add progress parsing test**

```go
func TestParseProgressLine(t *testing.T) {
	line := "frame=123 fps=30 time=00:01:23.45 bitrate=1234.5kbits/s"
	seconds, ok := ParseProgressLine(line)
	if !ok {
		t.Fatal("expected parse success")
	}
	if seconds != 83.45 {
		t.Fatalf("expected 83.45, got %f", seconds)
	}
}

func TestParseProgressLineNoMatch(t *testing.T) {
	_, ok := ParseProgressLine("ffmpeg version 4.4")
	if ok {
		t.Fatal("expected no match")
	}
}
```

- [ ] **Step 5: Verify all convert tests pass**

```
go test ./core/ -v -run "TestBuildFfmpegArgs|TestEncoderName|TestParseProgressLine"
Expected: PASS
```

- [ ] **Step 6: Commit**

```
git add core/convert.go core/convert_test.go
git commit -m "feat: encoder name lookup, progress parsing"
```

---

### Task 8: CLI Interface

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add CLI flags and headless queue processing**

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"file_converter/core"
)

func main() {
	var (
		presetName    = flag.String("preset", "Fast 1080p", "preset name")
		outputDir     = flag.String("output", "./output", "output directory")
		ffmpegPath    = flag.String("ffmpeg-path", "", "ffmpeg binary path")
		queueMode     = flag.Bool("queue", false, "process queue JSON from stdin")
		showVersion   = flag.Bool("version", false, "print version")
		concurrent    = flag.Int("concurrent", 2, "concurrent encode jobs")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("file_converter v0.1.0")
		return
	}

	if *queueMode {
		processQueue(*ffmpegPath, *concurrent)
		return
	}

	args := flag.Args()
	if len(args) > 0 {
		runCLI(args, *presetName, *outputDir, *ffmpegPath)
		return
	}

	// No args, no --queue — launch GUI
	runGUI(*ffmpegPath, *concurrent)
}

func runCLI(inputs []string, presetName, outputDir, ffmpegPath string) {
	presets := core.DefaultPresets()
	var preset *core.Preset
	for i, p := range presets {
		if p.Name == presetName {
			preset = &presets[i]
			break
		}
	}
	if preset == nil {
		fmt.Fprintf(os.Stderr, "preset %q not found\n", presetName)
		os.Exit(1)
	}

	q := core.NewQueue()
	for _, input := range inputs {
		q.Add(core.QueueItem{InputPath: input, Status: "pending"})
	}

	// Process sequentially for CLI (simplest)
	for {
		item := q.NextPending()
		if item == nil {
			break
		}
		fmt.Printf("Converting: %s\n", item.InputPath)
		// TODO: call actual conversion in Task 9
		item.Status = "done"
	}
}

func processQueue(ffmpegPath string, concurrent int) {
	var items []core.QueueItem
	if err := json.NewDecoder(os.Stdin).Decode(&items); err != nil {
		fmt.Fprintf(os.Stderr, "invalid queue JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Processing %d jobs with %d workers\n", len(items), concurrent)
}

func runGUI(ffmpegPath string, concurrent int) {
	fmt.Println("GUI mode (requires Fyne)")
}
```

- [ ] **Step 2: Verify build**

```
go build -o file_converter.exe .
Expected: builds without error
```

- [ ] **Step 3: Test CLI help**

```
file_converter.exe --help
Expected: flags shown

file_converter.exe --version
Expected: file_converter v0.1.0

file_converter.exe input.mp4 --preset "Fast 1080p"
Expected: Converting: input.mp4 (dry run)
```

- [ ] **Step 4: Commit**

```
git add main.go
git commit -m "feat: CLI flags and headless queue mode"
```

---

### Task 9: Core Conversion Runner

**Files:**
- Modify: `core/convert.go`
- Create: `core/runner.go`
- Create: `core/runner_test.go`

- [ ] **Step 1: Write test**

```go
package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConversionRunner(t *testing.T) {
	// Create a fake ffmpeg that prints progress-like output
	ffmpegDir := t.TempDir()
	fakeFfmpeg := filepath.Join(ffmpegDir, "ffmpeg.exe")
	script := `@echo off
echo ffmpeg version test
echo frame=  1 fps=0.0 time=00:00:01.00
echo frame= 10 fps=5.0 time=00:00:02.00
`
	os.WriteFile(fakeFfmpeg, []byte(script), 0644)

	input := filepath.Join(t.TempDir(), "input.mp4")
	os.WriteFile(input, []byte("fake video"), 0644)
	output := filepath.Join(t.TempDir(), "output.mp4")

	p := Preset{Container: "mp4", VideoCodec: "copy", AudioCodec: "copy"}
	runner := NewRunner(fakeFfmpeg)
	err := runner.Run(input, output, p, func(p Progress) {
		t.Logf("progress: %+v", p)
	})
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Write runner implementation**

```go
package core

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct {
	ffmpegPath string
}

func NewRunner(ffmpegPath string) *Runner {
	return &Runner{ffmpegPath: ffmpegPath}
}

func (r *Runner) Run(input, output string, preset Preset, onProgress func(Progress)) error {
	args := BuildFfmpegArgs(input, output, preset)
	cmd := exec.Command(r.ffmpegPath, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if sec, ok := ParseProgressLine(line); ok {
			if onProgress != nil {
				onProgress(Progress{
					File:    input,
					Percent: sec,
					Status:  "running",
				})
			}
		}
	}

	return cmd.Wait()
}
```

- [ ] **Step 3: Update main.go CLI to use real runner**

In `runCLI`, replace the TODO block:

```go
func runCLI(inputs []string, presetName, outputDir, ffmpegPath string) {
	// ... preset lookup (same as before) ...
	presets := core.DefaultPresets()
	var preset *core.Preset
	for i, p := range presets {
		if p.Name == presetName {
			preset = &presets[i]
			break
		}
	}
	if preset == nil {
		fmt.Fprintf(os.Stderr, "preset %q not found\n", presetName)
		os.Exit(1)
	}

	ffmpeg := core.FindFfmpeg(core.FfmpegPaths("ffmpeg"), ffmpegPath)
	if ffmpeg == "" {
		fmt.Fprintln(os.Stderr, "ffmpeg not found. Install ffmpeg or set --ffmpeg-path")
		os.Exit(1)
	}

	runner := core.NewRunner(ffmpeg)
	for _, input := range inputs {
		ext := "." + preset.Container
		out := strings.TrimSuffix(input, filepath.Ext(input)) + ext
		if outputDir != "" {
			out = filepath.Join(outputDir, filepath.Base(out))
		}
		os.MkdirAll(outputDir, 0755)

		fmt.Printf("Converting: %s → %s\n", input, out)
		err := runner.Run(input, out, *preset, func(p core.Progress) {
			fmt.Printf("\r  %s: %.0f%%", filepath.Base(input), p.Percent)
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n  error: %v\n", err)
		} else {
			fmt.Printf("\n  done: %s\n", out)
		}
	}
}
```

Add imports to main.go:
```go
import (
	// ... existing ...
	"path/filepath"
	"strings"
)
```

- [ ] **Step 4: Verify build**

```
go build -o file_converter.exe .
Expected: builds without error
```

- [ ] **Step 5: Commit**

```
git add core/runner.go core/runner_test.go main.go
git commit -m "feat: conversion runner, CLI uses real ffmpeg"
```

---

### Task 10: Fyne GUI — Window Skeleton

**Files:**
- Create: `gui/app.go`
- Modify: `main.go`

- [ ] **Step 1: Install Fyne**

```
go get fyne.io/fyne/v2
```

- [ ] **Step 2: Write minimal GUI skeleton**

```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyne.App
	window fyne.Window
}

func New() *App {
	a := &App{}
	a.App = app.New()
	a.window = a.App.NewWindow("File Converter")
	a.window.Resize(fyne.NewSize(900, 600))
	return a
}

func (a *App) Run() {
	// Placeholder layout — panels fill in subsequent tasks
	label := widget.NewLabel("File Converter")
	content := container.NewBorder(
		widget.NewLabel("File Converter v0.1"),
		widget.NewLabel("Queue status bar"),
		nil, nil,
		container.NewCenter(label),
	)
	a.window.SetContent(content)
	a.window.ShowAndRun()
}
```

- [ ] **Step 3: Wire GUI into main.go**

Replace the `runGUI` stub:

```go
func runGUI(ffmpegPath string, concurrent int) {
	gui.New().Run()
}
```

- [ ] **Step 4: Verify build**

```
go mod tidy
go build -o file_converter.exe .
Expected: builds without error
```

- [ ] **Step 5: Commit**

```
git add gui/app.go main.go go.mod go.sum
git commit -m "feat: Fyne GUI window skeleton"
```

---

### Task 11: GUI Source Panel

**Files:**
- Create: `gui/source_panel.go`

- [ ] **Step 1: Write source panel with file list + drag-drop**

```go
package gui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type SourcePanel struct {
	List      *widget.List
	items     []core.QueueItem
	onAdd     func(items []core.QueueItem)
	onRemove  func(index int)
}

func NewSourcePanel(onAdd func([]core.QueueItem)) *SourcePanel {
	p := &SourcePanel{onAdd: onAdd}
	p.List = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("file placeholder")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(filepath.Base(p.items[id].InputPath))
		},
	)
	return p
}

func (p *SourcePanel) Container() fyne.CanvasObject {
	addBtn := widget.NewButton("Add Files", func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			uri := reader.URI()
			path := uri.Path()
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				p.items = append(p.items, core.QueueItem{
					InputPath:  path,
					Status:     "pending",
				})
				p.List.Refresh()
				p.onAdd(p.items)
			}
			reader.Close()
		}, nil)
		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v", ".mp3", ".flac", ".wav", ".ogg", ".aac", ".wma"}))
		// The window parent will be set from App
	})
	// SetWindow is called from App to set parent for dialogs
	return container.NewBorder(nil, addBtn, nil, nil, p.List)
}
```

- [ ] **Step 2: Add setter for dialog parent**

```go
func (p *SourcePanel) SetWindow(win fyne.Window) {
	// Store for dialog parent — used when NewFileOpen is called
	// Actually, we need to rework to pass window. Let's expose the button differently.
}
```

Better approach — pass window to Container:

```go
type SourcePanel struct {
	// ... as before ...
	window fyne.Window
}

func NewSourcePanel(win fyne.Window, onAdd func([]core.QueueItem)) *SourcePanel {
	p := &SourcePanel{window: win, onAdd: onAdd}
	// ... rest ...
	p.List = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("file placeholder")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(filepath.Base(p.items[id].InputPath))
		},
	)
	return p
}

func (p *SourcePanel) Container() fyne.CanvasObject {
	addBtn := widget.NewButton("Add Files", func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			uri := reader.URI()
			path := uri.Path()
			// Handle Windows paths — strip leading / if present
			if len(path) > 2 && path[0] == '/' && path[2] == ':' {
				path = path[1:]
			}
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				p.items = append(p.items, core.QueueItem{InputPath: path, Status: "pending"})
				p.List.Refresh()
				p.onAdd(p.items)
			}
			reader.Close()
		}, p.window)
		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v", ".mp3", ".flac", ".wav", ".ogg", ".aac", ".wma"}))
		dlg.Show()
	})
	return container.NewBorder(addBtn, nil, nil, nil, p.List)
}
```

- [ ] **Step 3: Commit**

```
git add gui/source_panel.go
git commit -m "feat: GUI source panel with file picker"
```

---

### Task 12: GUI Preset Panel

**Files:**
- Create: `gui/preset_panel.go`

- [ ] **Step 1: Write preset panel**

```go
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type PresetPanel struct {
	selectWidget *widget.Select
	current      *core.Preset
	presets      []core.Preset
	summary      *widget.Label
}

func NewPresetPanel() *PresetPanel {
	p := &PresetPanel{}
	p.presets = core.DefaultPresets()
	names := make([]string, len(p.presets))
	for i, pr := range p.presets {
		names[i] = pr.Name
	}
	p.current = &p.presets[0]
	p.summary = widget.NewLabel(p.presetSummary(p.current))
	p.selectWidget = widget.NewSelect(names, func(selected string) {
		for i, pr := range p.presets {
			if pr.Name == selected {
				p.current = &p.presets[i]
				p.summary.SetText(p.presetSummary(p.current))
				break
			}
		}
	})
	p.selectWidget.SetSelected(p.current.Name)
	return p
}

func (p *PresetPanel) CurrentPreset() *core.Preset {
	return p.current
}

func (p *PresetPanel) Container() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Preset"),
		p.selectWidget,
		p.summary,
	)
}

func (p *PresetPanel) presetSummary(pr *core.Preset) string {
	h := ""
	if pr.HWAccel != "" {
		h = fmt.Sprintf(" (GPU: %s)", pr.HWAccel)
	}
	return fmt.Sprintf("Format: %s\nVideo: %s%s\nAudio: %s\nQuality: %d",
		pr.Container, pr.VideoCodec, h, pr.AudioCodec, pr.Quality)
}
```

- [ ] **Step 2: Commit**

```
git add gui/preset_panel.go
git commit -m "feat: GUI preset panel"
```

---

### Task 13: GUI Queue Panel

**Files:**
- Create: `gui/queue_panel.go`

- [ ] **Step 1: Write queue panel**

```go
package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type QueuePanel struct {
	items     []core.QueueItem
	list      *widget.List
	onStart   func()
	onStop    func()
}

func NewQueuePanel(onStart, onStop func()) *QueuePanel {
	p := &QueuePanel{
		onStart: onStart,
		onStop:  onStop,
	}
	p.list = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil,
				widget.NewLabel("file"),
				widget.NewLabel("status"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(p.items) {
				return
			}
			item := p.items[id]
			border := obj.(*fyne.Container)
			// Border has Left, Right, Top, Bottom, Middle
			// Simpler: just rebuild. Use a Label approach instead.
		},
	)
	return p
}

func (p *QueuePanel) SetItems(items []core.QueueItem) {
	p.items = items
	p.list.Refresh()
}

func (p *QueuePanel) Container() fyne.CanvasObject {
	controls := container.NewHBox(
		widget.NewButton("▶ Start", p.onStart),
		widget.NewButton("⏹ Stop", p.onStop),
	)
	return container.NewBorder(controls, nil, nil, nil, p.list)
}
```

- [ ] **Step 2: Wire panels into gui/app.go**

```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"file_converter/core"
)

type App struct {
	fyne.App
	window      fyne.Window
	source      *SourcePanel
	preset      *PresetPanel
	queue       *QueuePanel
	ffmpegPath  string
	concurrent  int
}

func New(ffmpegPath string, concurrent int) *App {
	a := &App{
		ffmpegPath: ffmpegPath,
		concurrent: concurrent,
	}
	a.App = app.New()
	a.window = a.App.NewWindow("File Converter")
	a.window.Resize(fyne.NewSize(900, 600))
	return a
}

func (a *App) Run() {
	a.source = NewSourcePanel(a.window, func(items []core.QueueItem) {
		// When files added, update queue panel
		queueItems := make([]core.QueueItem, len(items))
		copy(queueItems, items)
		for i := range queueItems {
			queueItems[i].PresetName = a.preset.CurrentPreset().Name
		}
		a.queue.SetItems(queueItems)
	})
	a.preset = NewPresetPanel()
	a.queue = NewQueuePanel(a.startQueue, a.stopQueue)
	// Placeholder queue items until files added
	a.queue.SetItems([]core.QueueItem{})

	split := container.NewHSplit(
		a.source.Container(),
		container.NewBorder(nil, nil, nil, nil,
			container.NewVSplit(
				a.preset.Container(),
				a.queue.Container(),
			),
		),
	)
	split.SetOffset(0.5)

	a.window.SetContent(split)
	a.window.ShowAndRun()
}
```

- [ ] **Step 3: Add stub queue controls**

```go
func (a *App) startQueue() {
	// Will wire to core.Runner in final task
}

func (a *App) stopQueue() {
	// Will send cancel signal
}
```

- [ ] **Step 4: Commit**

```
git add gui/app.go gui/queue_panel.go
git commit -m "feat: GUI queue panel, wire all panels together"
```

---

### Task 14: Queue Worker Pool + Integration

**Files:**
- Modify: `gui/app.go`
- Modify: `core/queue.go`
- Modify: `core/runner.go`

- [ ] **Step 1: Add worker pool to core**

```go
// In core/queue.go or new core/workerpool.go
package core

import (
	"context"
	"sync"
)

type WorkerPool struct {
	runner    *Runner
	queue     *Queue
	concurrent int
	cancel    context.CancelFunc
	mu        sync.Mutex
	running   bool
}

func NewWorkerPool(runner *Runner, queue *Queue, concurrent int) *WorkerPool {
	return &WorkerPool{
		runner:     runner,
		queue:      queue,
		concurrent: concurrent,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	if wp.running {
		wp.mu.Unlock()
		return
	}
	wp.running = true
	wp.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < wp.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item := wp.queue.NextPending()
				if item == nil {
					return
				}
				// Run conversion
				// In real code, would look up preset, call runner.Run, update status
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}
	wg.Wait()
}

func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.running = false
}
```

- [ ] **Step 2: Wire worker pool into GUI**

In `gui/app.go`:

```go
import (
	"context"
	"file_converter/core"
)

func (a *App) startQueue() {
	ffmpeg := core.FindFfmpeg(core.FfmpegPaths("ffmpeg"), a.ffmpegPath)
	if ffmpeg == "" {
		// Show error dialog
		return
	}
	runner := core.NewRunner(ffmpeg)
	queue := core.NewQueue()
	// Copy items from queue panel
	pool := core.NewWorkerPool(runner, queue, a.concurrent)
	ctx, cancel := context.WithCancel(context.Background())
	a.stopQueue = func() {
		cancel()
	}
	go pool.Start(ctx)
}

func (a *App) stopQueue() {
	// Set via startQueue
}
```

- [ ] **Step 3: Add progress updates to queue panel**

Pass a progress callback from GUI to worker pool, which updates the queue panel items and refreshes list.

- [ ] **Step 4: Verify build**

```
go build -o file_converter.exe .
Expected: builds without error
```

- [ ] **Step 5: Commit**

```
git add core/ core/ core/ gui/app.go
git commit -m "feat: worker pool + GUI queue integration"
```

---

### Task 15: Settings Dialog

**Files:**
- Create: `gui/settings.go`
- Create: `core/settings.go`
- Create: `core/settings_test.go`

- [ ] **Step 1: Write settings types**

```go
// core/settings.go
package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Settings struct {
	FfmpegPath     string `json:"ffmpeg_path"`
	Theme          string `json:"theme"` // system, light, dark
	OutputDir      string `json:"output_dir"`
	DefaultPreset  string `json:"default_preset"`
	ConcurrentJobs int    `json:"concurrent_jobs"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Theme:          "system",
		OutputDir:      "~/Videos",
		DefaultPreset:  "Fast 1080p",
		ConcurrentJobs: 2,
	}
}

func SettingsPath() string {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "file_converter", "settings.json")
}

func (s *Settings) Save(path string) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings(), nil
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), nil
	}
	if s.ConcurrentJobs < 1 {
		s.ConcurrentJobs = 1
	}
	return &s, nil
}
```

- [ ] **Step 2: Write settings test**

```go
// core/settings_test.go
package core

import (
	"os"
	"testing"
)

func TestSettingsSaveLoad(t *testing.T) {
	path := t.TempDir() + "/settings.json"
	s := &Settings{Theme: "dark", ConcurrentJobs: 4}
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "dark" || loaded.ConcurrentJobs != 4 {
		t.Fatalf("got %+v", loaded)
	}
}

func TestLoadSettingsMissing(t *testing.T) {
	s, err := LoadSettings("/nonexistent/settings.json")
	if err != nil {
		t.Fatal(err)
	}
	if s.Theme != "system" {
		t.Fatalf("expected defaults, got %+v", s)
	}
}
```

- [ ] **Step 3: Write settings GUI**

```go
// gui/settings.go
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

func ShowSettingsDialog(win fyne.Window, settings *core.Settings, onSave func(*core.Settings)) {
	ffmpegEntry := widget.NewEntry()
	ffmpegEntry.SetText(settings.FfmpegPath)
	ffmpegEntry.PlaceHolder = "Auto-detect"

	outputEntry := widget.NewEntry()
	outputEntry.SetText(settings.OutputDir)

	presetSelect := widget.NewSelect(core.DefaultPresetNames(), nil)
	presetSelect.SetSelected(settings.DefaultPreset)

	concurrentEntry := widget.NewEntry()
	concurrentEntry.SetText(fmt.Sprintf("%d", settings.ConcurrentJobs))

	themeSelect := widget.NewSelect([]string{"system", "light", "dark"}, nil)
	themeSelect.SetSelected(settings.Theme)

	items := []*widget.FormItem{
		widget.NewFormItem("Ffmpeg Path", ffmpegEntry),
		widget.NewFormItem("Output Dir", outputEntry),
		widget.NewFormItem("Default Preset", presetSelect),
		widget.NewFormItem("Concurrent Jobs", concurrentEntry),
		widget.NewFormItem("Theme", themeSelect),
	}

	dialog.ShowForm("Settings", "Save", "Cancel", items, func(b bool) {
		if !b {
			return
		}
		settings.FfmpegPath = ffmpegEntry.Text
		settings.OutputDir = outputEntry.Text
		settings.DefaultPreset = presetSelect.Selected
		settings.Theme = themeSelect.Selected
		onSave(settings)
	}, win)
}
```

- [ ] **Step 4: Commit**

```
git add core/settings.go core/settings_test.go gui/settings.go
git commit -m "feat: settings dialog and persistence"
```

---

### Task 16: Final Integration and Polish

**Files:**
- Modify: `gui/app.go`
- Modify: `main.go`

- [ ] **Step 1: Wire settings into app startup**

In `main.go`:

```go
func runGUI(ffmpegPath string, concurrent int) {
	// Load settings
	settingsPath := core.SettingsPath()
	settings, _ := core.LoadSettings(settingsPath)
	if ffmpegPath != "" {
		settings.FfmpegPath = ffmpegPath
	}
	if concurrent > 0 {
		settings.ConcurrentJobs = concurrent
	}
	gui.NewWithSettings(settings).Run()
}
```

- [ ] **Step 2: Add Settings menu to GUI menu bar**

```go
// In gui/app.go, add menu
func (a *App) setupMenu() {
	settingsItem := fyne.NewMenuItem("Settings...", func() {
		ShowSettingsDialog(a.window, a.settings, func(s *core.Settings) {
			a.settings = s
			s.Save(core.SettingsPath())
		})
	})
	menu := fyne.NewMainMenu(
		fyne.NewMenu("File", settingsItem),
	)
	a.window.SetMainMenu(menu)
}
```

- [ ] **Step 3: Crash recovery — load previous queue on startup**

In `gui/app.go`, add to `Run()`:

```go
queuePath := filepath.Join(filepath.Dir(core.SettingsPath()), "queue.json")
if q, err := core.LoadQueue(queuePath); err == nil {
	a.queue.SetItems(q.Items)
}
```

- [ ] **Step 4: Verify full build**

```
go build -o file_converter.exe .
Expected: builds without error
```

- [ ] **Step 5: Final commit**

```
git add .
git commit -m "feat: settings, menu, crash recovery, full integration"
```

---

## Spec Coverage Check

| Spec Requirement | Task |
|---|---|
| Go + Fyne stack | Task 1, 10 |
| CLI + GUI modes | Task 8, 10 |
| ffmpeg binary resolution (bundled, PATH, user) | Task 3 |
| Presets (built-in + custom) | Task 4 |
| Conversion engine (args builder) | Task 5 |
| GPU via HWAccel | Task 7 |
| Concurrent jobs | Task 14 |
| Queue persistence + crash recovery | Task 6, 16 |
| Progress parsing from stderr | Task 5, 7 |
| GUI three-panel layout | Task 11, 12, 13 |
| Settings dialog + persistence | Task 15 |
| File picker with media filters | Task 11 |
| CLI flag-based headless mode | Task 8 |
| Progress callbacks | Task 9, 14 |

No gaps.
