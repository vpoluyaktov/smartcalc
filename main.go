package main

import (
	"embed"
	"smartcalc/internal/data"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "dev"

func main() {
	app := NewApp()

	appMenu := createAppMenu(app)

	err := wails.Run(&options.App{
		Title:  "SmartCalc",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Menu:             appMenu,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func createAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// App menu (macOS) - this becomes the "SmartCalc" menu on macOS
	appSubmenu := appMenu.AddSubmenu("SmartCalc")
	appSubmenu.AddText("About SmartCalc", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:about")
	})
	appSubmenu.AddSeparator()
	appSubmenu.AddText("Quit SmartCalc", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	// File menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New", keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:new")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Open...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:open")
	})
	fileMenu.AddText("Save", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:save")
	})
	fileMenu.AddText("Save As...", keys.CmdOrCtrl("S"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:saveAs")
	})
	fileMenu.AddSeparator()

	// Recent files submenu
	recentMenu := fileMenu.AddSubmenu("Recent")
	recentFiles := app.GetRecentFiles()
	if len(recentFiles) == 0 {
		recentMenu.AddText("(No recent files)", nil, nil)
	} else {
		for _, path := range recentFiles {
			p := path // capture for closure
			recentMenu.AddText(path, nil, func(_ *menu.CallbackData) {
				runtime.EventsEmit(app.ctx, "menu:openRecent", p)
			})
		}
	}

	// Edit menu
	editMenu := appMenu.AddSubmenu("Edit")
	editMenu.AddText("Cut", keys.CmdOrCtrl("x"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:cut")
	})
	editMenu.AddText("Copy", keys.CmdOrCtrl("c"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:copy")
	})
	editMenu.AddText("Paste", keys.CmdOrCtrl("v"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:paste")
	})

	// Snippets menu - populated from data package
	snippetsMenu := appMenu.AddSubmenu("Snippets")
	for _, category := range data.GetSnippetCategories() {
		categoryMenu := snippetsMenu.AddSubmenu(category.Name)
		for _, s := range category.Snippets {
			snippet := s.Content
			categoryMenu.AddText(s.Name, nil, func(_ *menu.CallbackData) {
				runtime.EventsEmit(app.ctx, "menu:snippet", snippet)
			})
		}
	}

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Manual", keys.Key("F1"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:manual")
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("About SmartCalc", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:about")
	})

	return appMenu
}
