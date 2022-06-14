package client

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// open windows
func NewVScrollWindowAndShow(app fyne.App, windowSize *fyne.Size, content fyne.CanvasObject, title string) fyne.Window {
	w := app.NewWindow(title)
	return NewVScrollAndShowWindow(w, windowSize, content)
}

// open windows
func NewVScrollAndShowWindow(window fyne.Window, windowSize *fyne.Size, content fyne.CanvasObject) fyne.Window {

	//w := app.NewWindow(title)
	window.Resize(*windowSize)

	// Page flipping
	scroll := container.NewVScroll(content)
	scroll.Resize(*windowSize)

	window.SetContent(scroll)

	window.Show()

	// return ok
	return window
}
