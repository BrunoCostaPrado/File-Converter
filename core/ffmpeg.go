package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
