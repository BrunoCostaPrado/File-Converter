package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Settings struct {
	FfmpegPath     string `json:"ffmpeg_path"`
	Theme          string `json:"theme"`
	OutputDir      string `json:"output_dir"`
	DefaultPreset  string `json:"default_preset"`
	ConcurrentJobs int    `json:"concurrent_jobs"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Theme:          "system",
		OutputDir:      "~/Videos",
		DefaultPreset:  "Fast 1080p",
		ConcurrentJobs: 2,
	}
}

func SettingsPath() string {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "file_converter", "settings.json")
}

func (s *Settings) Save(path string) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings(), nil
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), nil
	}
	if s.ConcurrentJobs < 1 {
		s.ConcurrentJobs = 1
	}
	return &s, nil
}
