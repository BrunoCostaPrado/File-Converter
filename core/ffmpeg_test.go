package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFindFfmpegBundled(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "ffmpeg.exe")
	os.WriteFile(fake, []byte("fake"), 0644)
	got := FindFfmpeg([]string{dir}, "")
	if got != fake {
		t.Fatalf("expected %q, got %q", fake, got)
	}
}

func TestFindFfmpegUserPath(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "ffmpeg.exe")
	os.WriteFile(fake, []byte("fake"), 0644)
	got := FindFfmpeg([]string{}, fake)
	if got != fake {
		t.Fatalf("expected %q, got %q", fake, got)
	}
}

func TestFindFfmpegNotFound(t *testing.T) {
	got := FindFfmpeg([]string{"/nonexistent"}, "")
	if got != "" {
		p, err := exec.LookPath("ffmpeg")
		if err != nil || got != p {
			t.Fatalf("expected empty or PATH fallback, got %q", got)
		}
	}
}
