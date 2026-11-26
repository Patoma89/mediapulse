package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "MediaPulse",
		Width:     1240,
		Height:    800,
		MinWidth:  1000,
		MinHeight: 660,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "mediapulse-desktop-b3e7c941-4c28-4f5a-9b19-1d82f63e8042",
			OnSecondInstanceLaunch: app.onSecondInstance,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 13, B: 22, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			Theme:                windows.Dark,
			WebviewIsTransparent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
