package client

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

var (
	DevDebug bool = false // Global developer test mode
)

func MainNewAppRun() fyne.App {

	a := app.New()

	// Bright theme
	a.Settings().SetTheme(theme.LightTheme())

	// Show login window
	CreateShowRunLoginWindow(a)

	// start-up
	a.Run()

	return a

}
