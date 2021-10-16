package client

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
)

// 打开窗口
func NewVScrollWindowAndShow(app fyne.App, windowSize *fyne.Size, content fyne.CanvasObject, title string) fyne.Window {
	w := app.NewWindow(title)
	return NewVScrollAndShowWindow(w, windowSize, content)
}

// 打开窗口
func NewVScrollAndShowWindow(window fyne.Window, windowSize *fyne.Size, content fyne.CanvasObject) fyne.Window {

	//w := app.NewWindow(title)
	window.Resize(*windowSize)

	// 页面翻动
	scroll := container.NewVScroll(content)
	scroll.Resize(*windowSize)

	window.SetContent(scroll)

	window.Show()

	// return ok
	return window
}
