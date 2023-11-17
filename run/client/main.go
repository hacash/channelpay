package main

import (
	"github.com/flopp/go-findfont"
	"github.com/hacash/channelpay/client"
	"github.com/hacash/channelpay/client/accui"
	"os"
	"strings"
)

func init() {

	// Chinese font support
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

	//mainRelease()
	mainDev1()

}

func mainRelease() {
	// Developer mode
	client.DevDebug = false

	// start-up
	client.MainNewAppRun()

	//回退字体设置
	os.Unsetenv("FYNE_FONT")
}


func mainDev1() {

	accui.AssembleUIfiles("./client/accui")

	// Developer mode
	client.DevDebug = true

	// start-up
	client.MainNewAppRun()

	//回退字体设置
	os.Unsetenv("FYNE_FONT")
}
