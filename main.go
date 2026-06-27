package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"file_converter/core"
)

func main() {
	var (
		presetName  = flag.String("preset", "Fast 1080p", "preset name")
		outputDir   = flag.String("output", "./output", "output directory")
		ffmpegPath  = flag.String("ffmpeg-path", "", "ffmpeg binary path")
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
		runCLI(args, *presetName, *outputDir, *ffmpegPath)
		return
	}

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
		out := strings.TrimSuffix(input, filepath.Ext(input)) + ext
		if outputDir != "" {
			out = filepath.Join(outputDir, filepath.Base(out))
		}
		fmt.Printf("Converting: %s → %s\n", input, out)

		err := runner.Run(input, out, *preset, func(p core.Progress) {
			fmt.Printf("\r  %s: %.0f%%", filepath.Base(input), p.Percent)
		})
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
	fmt.Println("GUI mode (requires Fyne — Task 10)")
}
