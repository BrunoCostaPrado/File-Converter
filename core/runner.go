package core

import (
	"bufio"
	"fmt"
	"os/exec"
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
