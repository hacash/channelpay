package client

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
)

const (
	DevDebug bool = true // 全局开发者测试模式
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
