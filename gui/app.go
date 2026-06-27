package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyne.App
	window     fyne.Window
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
	label := widget.NewLabel("File Converter v0.1")
	content := container.NewBorder(
		widget.NewLabel("File Converter"),
		widget.NewLabel("Ready"),
		nil, nil,
		container.NewCenter(label),
	)
	a.window.SetContent(content)
	a.window.ShowAndRun()
}
