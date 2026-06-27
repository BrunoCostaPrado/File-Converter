package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type PresetPanel struct {
	selectWidget *widget.Select
	current      *core.Preset
	presets      []core.Preset
	summary      *widget.Label
}

func NewPresetPanel() *PresetPanel {
	p := &PresetPanel{}
	p.presets = core.DefaultPresets()
	names := core.DefaultPresetNames()
	p.current = &p.presets[0]
	p.summary = widget.NewLabel(p.presetSummary(p.current))
	p.selectWidget = widget.NewSelect(names, func(selected string) {
		for i, pr := range p.presets {
			if pr.Name == selected {
				p.current = &p.presets[i]
				p.summary.SetText(p.presetSummary(p.current))
				break
			}
		}
	})
	p.selectWidget.SetSelected(p.current.Name)
	return p
}

func (p *PresetPanel) CurrentPreset() *core.Preset {
	return p.current
}

func (p *PresetPanel) Container() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Preset"),
		p.selectWidget,
		p.summary,
	)
}

func (p *PresetPanel) presetSummary(pr *core.Preset) string {
	h := ""
	if pr.HWAccel != "" {
		h = fmt.Sprintf(" (GPU: %s)", pr.HWAccel)
	}
	return fmt.Sprintf("Format: %s\nVideo: %s%s\nAudio: %s\nQuality: %d",
		pr.Container, pr.VideoCodec, h, pr.AudioCodec, pr.Quality)
}
