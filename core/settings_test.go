package core

import (
	"testing"
)

func TestSettingsSaveLoad(t *testing.T) {
	path := t.TempDir() + "/settings.json"
	s := &Settings{Theme: "dark", ConcurrentJobs: 4}
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "dark" || loaded.ConcurrentJobs != 4 {
		t.Fatalf("got %+v", loaded)
	}
}

func TestLoadSettingsMissing(t *testing.T) {
	s, err := LoadSettings("/nonexistent/settings.json")
	if err != nil {
		t.Fatal(err)
	}
	if s.Theme != "system" {
		t.Fatalf("expected defaults, got %+v", s)
	}
}
