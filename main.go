package main

import (
	"embed"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/app"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := app.NewService()
	err := wails.Run(&options.App{
		Title:  "YiwoDraftViewer",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			svc,
		},
	})
	if err != nil {
		panic(err)
	}
}