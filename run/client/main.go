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

		if strings.Contains(path, "simkai") ||
			strings.Contains(path, "simhei") ||
			strings.Contains(path, "simsun") {
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
	client.DevDebug = true

	// 启动
	client.MainNewAppRun()

	//回退字体设置
	os.Unsetenv("FYNE_FONT")
}
