package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

func ShowSettingsDialog(win fyne.Window, settings *core.Settings, onSave func(*core.Settings)) {
	ffmpegEntry := widget.NewEntry()
	ffmpegEntry.SetText(settings.FfmpegPath)
	ffmpegEntry.PlaceHolder = "Auto-detect (system PATH or bundled)"

	outputEntry := widget.NewEntry()
	outputEntry.SetText(settings.OutputDir)

	presetSelect := widget.NewSelect(core.DefaultPresetNames(), nil)
	presetSelect.SetSelected(settings.DefaultPreset)

	concurrentEntry := widget.NewEntry()
	concurrentEntry.SetText(fmt.Sprintf("%d", settings.ConcurrentJobs))

	themeSelect := widget.NewSelect([]string{"system", "light", "dark"}, nil)
	themeSelect.SetSelected(settings.Theme)

	items := []*widget.FormItem{
		{Text: "Ffmpeg Path", Widget: ffmpegEntry},
		{Text: "Output Dir", Widget: outputEntry},
		{Text: "Default Preset", Widget: presetSelect},
		{Text: "Concurrent Jobs", Widget: concurrentEntry},
		{Text: "Theme", Widget: themeSelect},
	}

	dialog.ShowForm("Settings", "Save", "Cancel", items, func(b bool) {
		if !b {
			return
		}
		settings.FfmpegPath = ffmpegEntry.Text
		settings.OutputDir = outputEntry.Text
		settings.DefaultPreset = presetSelect.Selected
		settings.Theme = themeSelect.Selected
		fmt.Sscanf(concurrentEntry.Text, "%d", &settings.ConcurrentJobs)
		if settings.ConcurrentJobs < 1 {
			settings.ConcurrentJobs = 1
		}
		onSave(settings)
	}, win)
}
