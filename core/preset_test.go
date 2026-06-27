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

func TestDefaultPresetNames(t *testing.T) {
	names := DefaultPresetNames()
	if len(names) == 0 {
		t.Fatal("expected preset names")
	}
	if names[0] != "Fast 1080p" {
		t.Fatalf("expected 'Fast 1080p' first, got %q", names[0])
	}
}
