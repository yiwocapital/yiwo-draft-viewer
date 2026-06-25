package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/app"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// TagVersion, CommitID, and ReleaseVersion are set at build time via -ldflags.
// Defaults are "dev" / "unknown" / "" for go run / quick local builds.
// ReleaseVersion is non-empty only when HEAD is exactly on a tag.
var (
	TagVersion     = "dev"
	CommitID       = "unknown"
	ReleaseVersion = ""
)

// appWrapper exposes a small surface of cross-cutting operations that need a
// runtime context (Quit, EventsEmit) but don't belong on the persisted Service
// API. The frontend calls SaveAndClose from the Cmd+Q → "保存" handler in the
// dirty-quit flow: it saves the buffer, then quits if Save succeeded.
type appWrapper struct {
	ctx context.Context
	svc *app.Service
}

func (a *appWrapper) startup(ctx context.Context) {
	a.ctx = ctx
}

// SaveAndClose writes content via Service.Save. On success, calls runtime.Quit
// so the app exits after the dirty-quit dialog's "保存" branch. On failure,
// frontend already shows an error toast; app stays open for retry.
func (a *appWrapper) SaveAndClose(content string) {
	if a.svc == nil {
		return
	}
	res := a.svc.Save(content)
	if res.Ok && a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// Quit requests app shutdown. Called by the frontend after the user picks
// "discard" in the dirty-quit dialog (or anytime the frontend wants to
// trigger a clean exit without going through the OnBeforeClose dialog flow).
func (a *appWrapper) Quit() {
	if a.ctx == nil {
		return
	}
	runtime.Quit(a.ctx)
}

func main() {
	svc := app.NewService()
	wrapper := &appWrapper{svc: svc}

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
		// Only emit; don't run svc.CloseFile yet. Frontend listener checks
		// dirty state and decides whether to call api.closeFile() to clear
		// backend state.
		runtime.EventsEmit(svc.Ctx(), "close-file")
	})

	viewMenu := AppMenu.AddSubmenu("View")
	viewMenu.AddText("Refresh", keys.CmdOrCtrl("r"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(svc.Ctx(), "reload")
	})

	AppMenu.Append(menu.AppMenu())
	AppMenu.Append(menu.EditMenu())
	AppMenu.Append(menu.WindowMenu())

	err := wails.Run(&options.App{
		Title:  windowTitle(),
		Width:  1280,
		Height: 800,
		Menu:   AppMenu,
		Mac: &mac.Options{
			// macOS Big Sur+ renders a second copy of the window title in the
			// toolbar area under the title bar when TitlebarAppearsTransparent
			// is false (the default) and no NSToolbar is configured — visible
			// as a white bar with the title text repeated just below the system
			// title bar. Setting TitlebarAppearsTransparent=true keeps the
			// system title bar text but drops the redundant in-webview copy.
			// Verified on macOS 14.6.1 with Wails v2.12.0.
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			svc.Startup(ctx)
			wrapper.startup(ctx)
			// Restore window size + position if saved values are non-zero.
			// Skip position restoration if zero (first launch), let macOS center it.
			cfg, _ := config.Load(svc.ConfigDir())
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
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// Capture the final window state right before exit, so the very
			// last position/size is persisted even if no resize or mouseup
			// fired between the user's last action and the close.
			svc.WindowChanged()

			// Not dirty → allow close (no dialog needed).
			if !svc.IsDirty() {
				return false
			}

			// Dirty → emit event to frontend; frontend will run a confirm()
			// and either call appWrapper.Quit() (discard), SaveAndClose
			// (save), or do nothing (cancel). Keep open until frontend
			// resolves. The previous MessageDialog-based flow silently
			// failed in some Wails v2 + macOS combinations, so we moved
			// the prompt to the frontend.
			runtime.EventsEmit(ctx, "request-dirty-quit")
			return true
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		Bind: []interface{}{
			svc,
			wrapper,
		},
	})
	if err != nil {
		panic(err)
	}
}

func windowTitle() string {
	// If HEAD is exactly on a tag, show only the tag (clean release look).
	if ReleaseVersion != "" {
		return fmt.Sprintf("Yiwo Draft Viewer %s", ReleaseVersion)
	}
	// Otherwise (unreleased / dev build), show only the commit id.
	if CommitID == "" || CommitID == "unknown" {
		return "Yiwo Draft Viewer"
	}
	return fmt.Sprintf("Yiwo Draft Viewer (%s)", CommitID)
}