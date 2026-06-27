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
