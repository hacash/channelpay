package main

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/flopp/go-findfont"
	"os"
	"strings"
)

func init() {

	// 中文字体支持
	fontPaths := findfont.List()
	for _, path := range fontPaths {

		if strings.Contains(path, "simkai") ||
			strings.Contains(path, "simhei") ||
			strings.Contains(path, "simsun") {
			fmt.Println(path)
			os.Setenv("FYNE_FONT", path)
			break
		}
	}

}

func main() {

	a := app.New()
	w := a.NewWindow("Hacash Channel Chain Payment User Client")
	windowSize := fyne.Size{
		Width:  900,
		Height: 600,
	}
	w.Resize(windowSize)

	objs := container.NewVBox()

	objs.Add(widget.NewLabel("\n\n"))

	// 页面翻动
	scroll := container.NewVScroll(objs)

	w.SetContent(scroll)

	w.Show()

	a.Run()

	//回退字体设置
	os.Unsetenv("FYNE_FONT")

}
