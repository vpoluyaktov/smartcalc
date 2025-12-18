package main

import (
	"embed"

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

	// Snippets menu
	snippetsMenu := appMenu.AddSubmenu("Snippets")
	snippets := []struct {
		Name    string
		Content string
	}{
		{"Basic Math", "10 + 20 * 3 =\n"},
		{"Percentage", "$100 - 20% =\n"},
		{"Currency Calculation", "$1,500.00 + $250.50 =\n"},
		{"Line Reference", "100 =\n\\1 * 2 =\n"},
		{"Scientific", "sin(45) + cos(30) =\n"},
		{"Complex Expression", "$1,000 x 12 - 15% + $500 =\n"},
		{"Comparison", "25 > 2.5 =\n100 >= 100 =\n5 != 3 =\n"},
		{"Base Conversion", "255 in hex =\n0xFF in dec =\n25 in bin =\n0b11001 in oct =\n"},
		{"Current Time", "now ="},
		{"Time in City", "now in Seattle =\nnow in New York =\nnow in Kiev =\n"},
		{"Date Arithmetic", "today() =\n\\1 + 30 days =\n\\1 - 1 week =\n"},
		{"Date Difference", "19/01/22 - now =\n"},
		{"Duration Conversion", "861.5 hours in days =\n"},
		{"Time Zone Conversion", "6:00 am Seattle in Kiev =\n"},
		{"Date Range", "Dec 6 till March 11 =\n"},
		{"Subnet Info", "10.100.0.0/24 =\n"},
		{"Split to Subnet", "10.100.0.0/16 / 4 subnets =\n"},
		{"Split by Hosts", "10.100.0.0/28 / 16 hosts =\n"},
		{"Subnet Mask", "mask for /24 =\nwildcard for /24 =\n"},
		{"IP in Range", "is 10.100.0.50 in 10.100.0.0/24 =\n"},
	}
	for _, s := range snippets {
		snippet := s.Content
		snippetsMenu.AddText(s.Name, nil, func(_ *menu.CallbackData) {
			runtime.EventsEmit(app.ctx, "menu:snippet", snippet)
		})
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
