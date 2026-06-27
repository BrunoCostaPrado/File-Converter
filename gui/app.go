package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	pool       *core.WorkerPool
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
	ffmpeg := core.FindFfmpeg(core.FfmpegPaths("ffmpeg"), a.ffmpegPath)
	if ffmpeg == "" {
		dialog.ShowError(fmt.Errorf("ffmpeg not found. Install ffmpeg or set path in settings"), a.window)
		return
	}

	items := a.queue.GetItems()
	if len(items) == 0 {
		return
	}
	q := core.NewQueue()
	for i := range items {
		items[i].PresetName = a.preset.CurrentPreset().Name
		items[i].Status = "pending"
	}
	q.Add(items...)

	runner := core.NewRunner(ffmpeg)
	a.pool = core.NewWorkerPool(runner, q, a.concurrent)
	a.pool.OnProgress = func(p core.Progress) {
		a.queue.UpdateProgress(p.File, p.Percent, p.Status)
	}
	go a.pool.Start()
}

func (a *App) stopQueue() {
	if a.pool != nil {
		a.pool.Stop()
		a.pool = nil
	}
}
