package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type QueuePanel struct {
	items   []core.QueueItem
	list    *widget.List
	onStart func()
	onStop  func()
}

func NewQueuePanel(onStart, onStop func()) *QueuePanel {
	p := &QueuePanel{onStart: onStart, onStop: onStop}
	p.list = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("file"),
				widget.NewLabel("status"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(p.items) {
				return
			}
			item := p.items[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(filepath.Base(item.InputPath))
			status := item.Status
			if item.Status == "running" && item.Progress > 0 {
				status = fmt.Sprintf("%.0f%%", item.Progress)
			}
			box.Objects[1].(*widget.Label).SetText(status)
		},
	)
	return p
}

func (p *QueuePanel) SetItems(items []core.QueueItem) {
	p.items = items
	p.list.Refresh()
}

func (p *QueuePanel) Container() fyne.CanvasObject {
	controls := container.NewHBox(
		widget.NewButton("▶ Start", p.onStart),
		widget.NewButton("⏹ Stop", p.onStop),
	)
	return container.NewBorder(controls, nil, nil, nil, p.list)
}
