package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	ffmpegPath string
}

func NewRunner(ffmpegPath string) *Runner {
	return &Runner{ffmpegPath: ffmpegPath}
}

func (r *Runner) Run(input, output string, preset Preset, onProgress func(Progress)) error {
	// Sanitize input: "test" as sole filename meant file named "-", not stdin
	inputPath := input
	if inputPath == "-" {
		inputPath = filepath.Join(".", "-")
	}

	// Check input exists before invoking ffmpeg
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	args := BuildFfmpegArgs(inputPath, output, preset)
	cmd := exec.Command(r.ffmpegPath, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	// Collect stderr for error reporting
	var stderrBuf []string
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		stderrBuf = append(stderrBuf, line)
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

	err = cmd.Wait()
	if err != nil {
		if len(stderrBuf) > 5 {
			stderrBuf = stderrBuf[len(stderrBuf)-5:]
		}
		return fmt.Errorf("ffmpeg failed:\n%s", strings.Join(stderrBuf, "\n"))
	}
	return nil
}
