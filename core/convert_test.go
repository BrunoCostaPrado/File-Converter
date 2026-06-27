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

func TestBuildFfmpegArgsHWAccel(t *testing.T) {
	p := Preset{
		Container: "mp4", VideoCodec: "h264", AudioCodec: "aac",
		Quality: 23, Preset: "medium", Resolution: "1920x1080",
		HWAccel: "nvenc",
	}
	args := BuildFfmpegArgs("in.mp4", "out.mp4", p)
	// HWAccel should skip -preset and -crf
	expected := []string{"-i", "in.mp4", "-c:v", "h264_nvenc", "-c:a", "aac", "-vf", "scale=1920:1080", "out.mp4"}
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
