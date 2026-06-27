package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"file_converter/core"
)

type App struct {
	fyne.App
	window     fyne.Window
	source     *SourcePanel
	preset     *PresetPanel
	queue      *QueuePanel
	ffmpegPath string
	concurrent int
}

func New(ffmpegPath string, concurrent int) *App {
	a := &App{
		ffmpegPath: ffmpegPath,
		concurrent: concurrent,
	}
	a.App = app.New()
	a.window = a.App.NewWindow("File Converter")
	a.window.Resize(fyne.NewSize(900, 600))
	return a
}

func (a *App) Run() {
	a.source = NewSourcePanel(a.window, func(items []core.QueueItem) {
		queueItems := make([]core.QueueItem, len(items))
		copy(queueItems, items)
		for i := range queueItems {
			queueItems[i].PresetName = a.preset.CurrentPreset().Name
		}
		a.queue.SetItems(queueItems)
	})
	a.preset = NewPresetPanel()
	a.queue = NewQueuePanel(a.startQueue, a.stopQueue)
	a.queue.SetItems([]core.QueueItem{})

	left := a.source.Container()
	right := container.NewVSplit(
		a.preset.Container(),
		a.queue.Container(),
	)
	split := container.NewHSplit(left, right)
	split.SetOffset(0.4)

	a.window.SetContent(split)
	a.window.ShowAndRun()
}

func (a *App) startQueue() {
	// Will be wired in Task 14
}

func (a *App) stopQueue() {
	// Will be wired in Task 14
}
