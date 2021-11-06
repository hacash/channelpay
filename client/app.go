package client

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

var (
	DevDebug bool = false // 全局开发者测试模式
)

func MainNewAppRun() fyne.App {

	a := app.New()

	// 明亮主题
	a.Settings().SetTheme(theme.LightTheme())

	// 显示登录窗口
	CreateShowRunLoginWindow(a)

	// 启动
	a.Run()

	return a

}
