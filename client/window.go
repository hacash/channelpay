package client

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
)

// 打开窗口
func NewVScrollWindowAndShow(app fyne.App, windowSize *fyne.Size, content fyne.CanvasObject, title string) fyne.Window {

	w := app.NewWindow(title)
	w.Resize(*windowSize)

	box := container.NewVBox()

	box.Add(content)

	// 页面翻动
	scroll := container.NewVScroll(box)
	scroll.Resize(*windowSize)

	w.SetContent(scroll)

	w.Show()

	// return ok
	return w
}
