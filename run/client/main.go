package main

import (
	"github.com/flopp/go-findfont"
	"github.com/hacash/channelpay/client"
	"os"
	"strings"
)

func init() {

	// 中文字体支持
	fontPaths := findfont.List()
	for _, path := range fontPaths {

		if strings.Contains(path, "uming.ttc") ||
			strings.Contains(path, "ukai.ttc") ||
			strings.Contains(path, "simkai.ttf") ||
			strings.Contains(path, "simhei.ttf") ||
			strings.Contains(path, "simsun.ttf") ||
			strings.Contains(path, "STHeiti") {
			//fmt.Println(path)
			os.Setenv("FYNE_FONT", path)
			break
		}
	}

}

func main() {

	mainRelease()
	//mainDev1()

}

func mainRelease() {
	// 开发者模式
	client.DevDebug = false

	// 启动
	client.MainNewAppRun()

	//回退字体设置
	os.Unsetenv("FYNE_FONT")
}
