package main

import (
	"context"
	"embed"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/app"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := app.NewService()

	AppMenu := menu.NewMenu()

	fileMenu := AppMenu.AddSubmenu("File")
	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
		path, _ := runtime.OpenFileDialog(svc.Ctx(), runtime.OpenDialogOptions{
			Title: "Open Markdown File",
			Filters: []runtime.FileFilter{
				{DisplayName: "Markdown", Pattern: "*.md"},
			},
		})
		if path != "" {
			runtime.EventsEmit(svc.Ctx(), "open-file", path)
		}
	})
	fileMenu.AddText("Close", keys.CmdOrCtrl("w"), func(cd *menu.CallbackData) {
		svc.CloseFile()
		runtime.EventsEmit(svc.Ctx(), "close-file")
	})

	viewMenu := AppMenu.AddSubmenu("View")
	viewMenu.AddText("Refresh", keys.CmdOrCtrl("r"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(svc.Ctx(), "reload")
	})

	AppMenu.Append(menu.EditMenu())
	AppMenu.Append(menu.WindowMenu())

	err := wails.Run(&options.App{
		Title:  "YiwoDraftViewer",
		Width:  1280,
		Height: 800,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			svc.Startup(ctx)
			// Restore window size + position if saved values are non-zero.
			// Skip position restoration if zero (first launch), let macOS center it.
			cfg, _ := config.Load(".")
			if cfg.Window.Width > 0 && cfg.Window.Height > 0 {
				runtime.WindowSetSize(ctx, cfg.Window.Width, cfg.Window.Height)
			}
			if cfg.Window.X != 0 || cfg.Window.Y != 0 {
				runtime.WindowSetPosition(ctx, cfg.Window.X, cfg.Window.Y)
			}
			runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
				if len(paths) > 0 {
					runtime.EventsEmit(ctx, "file-dropped", paths[0])
				}
			})
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		Bind: []interface{}{
			svc,
		},
	})
	if err != nil {
		panic(err)
	}
}