package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"file_converter/core"
	"file_converter/gui"
)

func main() {
	var (
		presetName  = flag.String("preset", "Fast 1080p", "preset name")
		outputDir   = flag.String("output", "./output", "output directory")
		ffmpegPath  = flag.String("ffmpeg-path", "", "ffmpeg binary path")
		hwaccel     = flag.String("hwaccel", "", "GPU backend: nvenc, qsv, amd, videotoolbox")
		bitrate     = flag.String("bitrate", "", "video bitrate (e.g. 2M, 1000k)")
		keepFormat  = flag.Bool("keep-format", false, "keep original file extension")
		queueMode   = flag.Bool("queue", false, "process queue JSON from stdin")
		showVersion = flag.Bool("version", false, "print version")
		concurrent  = flag.Int("concurrent", 2, "concurrent encode jobs")
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
		runCLI(args, *presetName, *outputDir, *ffmpegPath, *hwaccel, *bitrate, *keepFormat)
		return
	}

	runGUI(*ffmpegPath, *concurrent)
}

func runCLI(inputs []string, presetName, outputDir, ffmpegPath, hwaccel, bitrate string, keepFormat bool) {
	presets := core.DefaultPresets()
	var preset *core.Preset
	for i, p := range presets {
		if p.Name == presetName {
			preset = &presets[i]
			break
		}
	}
	if preset == nil {
		fmt.Fprintf(os.Stderr, "preset %q not found. Available: %v\n", presetName, strings.Join(core.DefaultPresetNames(), ", "))
		os.Exit(1)
	}
	if hwaccel != "" {
		preset.HWAccel = hwaccel
	}

	ffmpeg := core.FindFfmpeg(core.FfmpegPaths("ffmpeg"), ffmpegPath)
	if ffmpeg == "" {
		fmt.Fprintln(os.Stderr, "ffmpeg not found. Install ffmpeg or set --ffmpeg-path")
		os.Exit(1)
	}

	if outputDir != "" {
		os.MkdirAll(outputDir, 0755)
	}

	runner := core.NewRunner(ffmpeg)
	for _, input := range inputs {
		ext := "." + preset.Container
		if keepFormat {
			ext = filepath.Ext(input)
		}
		out := strings.TrimSuffix(input, filepath.Ext(input)) + ext
		if outputDir != "" {
			out = filepath.Join(outputDir, filepath.Base(out))
		}
		usedBitrate := bitrate
		if usedBitrate == "" && preset.HWAccel != "" {
			if b := core.ProbeVideoBitrate(ffmpeg, input); b != "" {
				usedBitrate = b
			}
		}
		fmt.Printf("Converting: %s → %s (bitrate: %s)\n", input, out, usedBitrate)

		err := runner.Run(input, out, *preset, func(p core.Progress) {
			fmt.Printf("\r  %s: %.0f%%", filepath.Base(input), p.Percent)
		}, usedBitrate)
		if err != nil {
			fmt.Printf("\n  error: %v\n", err)
		} else {
			fmt.Printf("\n  done: %s\n", out)
		}
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
	settingsPath := core.SettingsPath()
	settings, _ := core.LoadSettings(settingsPath)
	if ffmpegPath != "" {
		settings.FfmpegPath = ffmpegPath
	}
	if concurrent > 0 {
		settings.ConcurrentJobs = concurrent
	}
	gui.New(settings).Run()
}
