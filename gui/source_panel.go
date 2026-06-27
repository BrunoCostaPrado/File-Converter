package gui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"file_converter/core"
)

type SourcePanel struct {
	window fyne.Window
	List   *widget.List
	items  []core.QueueItem
	onAdd  func(items []core.QueueItem)
}

func NewSourcePanel(win fyne.Window, onAdd func([]core.QueueItem)) *SourcePanel {
	p := &SourcePanel{window: win, onAdd: onAdd}
	p.List = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("file placeholder")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(p.items) {
				obj.(*widget.Label).SetText(filepath.Base(p.items[id].InputPath))
			}
		},
	)
	return p
}

func (p *SourcePanel) Container() fyne.CanvasObject {
	addBtn := widget.NewButton("Add Files", func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			uri := reader.URI()
			path := uri.Path()
			if len(path) > 2 && path[0] == '/' && path[2] == ':' {
				path = path[1:]
			}
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				p.items = append(p.items, core.QueueItem{InputPath: path, Status: "pending"})
				p.List.Refresh()
				p.onAdd(p.items)
			}
			reader.Close()
		}, p.window)
		dlg.SetFilter(storage.NewExtensionFileFilter([]string{
			".mp4", ".mkv", ".avi", ".mov", ".webm",
			".flv", ".wmv", ".m4v", ".mp3", ".flac",
			".wav", ".ogg", ".aac", ".wma",
		}))
		dlg.Show()
	})
	return container.NewBorder(addBtn, nil, nil, nil, p.List)
}
