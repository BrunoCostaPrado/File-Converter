package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func FfmpegPaths(bundledDir string) []string {
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

func ProbeVideoBitrate(ffmpegPath, input string) string {
	// Derive ffprobe path from ffmpeg path
	dir := filepath.Dir(ffmpegPath)
	base := strings.TrimSuffix(filepath.Base(ffmpegPath), filepath.Ext(ffmpegPath))
	probe := filepath.Join(dir, base+"probe")
	if _, err := os.Stat(probe); err != nil {
		probe = filepath.Join(dir, base+"probe.exe")
		if _, err := os.Stat(probe); err != nil {
			// Fallback to PATH
			if p, err := exec.LookPath("ffprobe"); err != nil {
				return ""
			} else {
				probe = p
			}
		}
	}

	out, err := exec.Command(probe, "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=bit_rate",
		"-of", "default=noprint_wrappers=1:nokey=1", input).Output()
	if err != nil {
		return ""
	}

	bitrateStr := strings.TrimSpace(string(out))
	if bitrateStr == "" || bitrateStr == "N/A" {
		return ""
	}

	bits, err := strconv.Atoi(bitrateStr)
	if err != nil || bits <= 0 {
		return ""
	}

	// Convert to k suffix for precision (avoid integer-truncated M)
	if bits >= 1_000 {
		return strconv.Itoa(bits/1_000) + "k"
	}
	return strconv.Itoa(bits)
}
