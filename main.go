package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"supercalc/internal/calc"
	"supercalc/internal/eval"
	"supercalc/internal/menu"
	"supercalc/internal/storage"
	"supercalc/internal/ui"
)

// Version info set via ldflags
var version = "dev"

func main() {
	a := app.NewWithID("com.supercalc.app")
	a.Settings().SetTheme(ui.NewLargerTextTheme())
	w := a.NewWindow("SuperCalc - Untitled")
	w.Resize(fyne.NewSize(1000, 700))

	lineNums := widget.NewLabel("1")
	lineNums.TextStyle = fyne.TextStyle{Monospace: true}
	lineNums.Wrapping = fyne.TextWrapOff
	lineNums.Alignment = fyne.TextAlignTrailing

	entry := ui.NewCustomMultiLineEntry()
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	entry.SetPlaceHolder("Type expressions like: $95.88 x (167 + 175) - 20% =\nDate/Time: now in Seattle =, today() + 30 days =\nNetwork: split 10.0.0.0/16 to 4 subnets =\nReference prior results as \\\\1, \\\\2, ...")

	// Place line numbers and entry side by side in a single row, then scroll together
	lineNumBox := container.New(&ui.FixedWidthLayout{Width: 50}, lineNums)
	editorRow := container.NewBorder(nil, nil, lineNumBox, nil, entry)
	editorArea := container.NewScroll(editorRow)

	// Status bar at bottom - version on the right
	statusLabel := widget.NewLabel(fmt.Sprintf("Version %s", version))
	statusLabel.Alignment = fyne.TextAlignTrailing
	statusBar := container.NewBorder(nil, nil, nil, statusLabel, nil)

	content := container.NewBorder(nil, statusBar, nil, nil, editorArea)

	var (
		mu            sync.Mutex
		debounce      *time.Timer
		lastText      string
		prevText      string
		prevLineCount = 1
		updating      bool
	)

	recalc := func(text string) {
		mu.Lock()
		if updating || text == lastText {
			mu.Unlock()
			return
		}
		lastText = text
		mu.Unlock()

		lines := strings.Split(text, "\n")
		results := calc.EvalLines(lines)

		outLines := make([]string, len(results))
		for i, r := range results {
			outLines[i] = r.Output
		}

		newText := strings.Join(outLines, "\n")
		if newText != text {
			mu.Lock()
			updating = true
			mu.Unlock()
			entry.SetText(newText)
			mu.Lock()
			updating = false
			mu.Unlock()
		}
		lineNums.SetText(calc.BuildLineNumbers(len(lines)))
	}

	updateLineNums := func(text string) {
		n := strings.Count(text, "\n") + 1
		lineNums.SetText(calc.BuildLineNumbers(n))
	}

	setContent := func(text string) {
		mu.Lock()
		updating = true
		prevLineCount = strings.Count(text, "\n") + 1
		lastText = ""
		mu.Unlock()
		entry.SetText(text)
		mu.Lock()
		updating = false
		mu.Unlock()
		updateLineNums(text)
		recalc(text)
	}

	entry.OnChanged = func(s string) {
		mu.Lock()
		if updating {
			mu.Unlock()
			return
		}

		currentLineCount := strings.Count(s, "\n") + 1
		delta := currentLineCount - prevLineCount
		oldText := prevText

		// Check if line count changed and we have previous text to compare
		if delta != 0 && oldText != "" {
			adjusted := eval.AdjustReferences(oldText, s)
			if adjusted != s {
				updating = true
				prevLineCount = currentLineCount
				prevText = adjusted
				mu.Unlock()
				entry.SetText(adjusted)
				mu.Lock()
				updating = false
				mu.Unlock()
				updateLineNums(adjusted)
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(150*time.Millisecond, func() { recalc(adjusted) })
				return
			}
		}
		prevLineCount = currentLineCount
		prevText = s
		mu.Unlock()

		updateLineNums(s)
		mu.Lock()
		if debounce != nil {
			debounce.Stop()
		}
		debounce = time.AfterFunc(150*time.Millisecond, func() { recalc(s) })
		mu.Unlock()
	}

	// Custom copy function that replaces references with values
	customCopy := func() {
		text := entry.Text
		if entry.SelectedText() != "" {
			text = entry.SelectedText()
		}
		resolved := calc.ReplaceRefsWithValues(text)
		w.Clipboard().SetContent(resolved)
	}

	// Set the custom copy handler on the entry widget
	entry.OnCopy = customCopy

	// Storage manager for file operations
	var storageMgr *storage.Manager
	storageMgr = storage.NewManager(a, w,
		func(content string) {
			setContent(content)
			storageMgr.MarkSaved()
		},
		func() string { return entry.Text },
		func() {
			setContent("")
			storageMgr.MarkSaved()
		},
	)

	// Create main menu
	mainMenu := menu.CreateMainMenu(w, menu.Callbacks{
		New:        storageMgr.New,
		Open:       storageMgr.Open,
		Save:       storageMgr.Save,
		SaveAs:     storageMgr.SaveAs,
		OpenRecent: func(path string) { storageMgr.OpenFile(path) },
		GetRecent:  storageMgr.GetRecentFiles,

		Cut:   func() { entry.TypedShortcut(&fyne.ShortcutCut{Clipboard: w.Clipboard()}) },
		Copy:  customCopy,
		Paste: func() { entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: w.Clipboard()}) },

		InsertSnippet: func(snippet string) {
			current := entry.Text
			if current != "" && !strings.HasSuffix(current, "\n") {
				current += "\n"
			}
			setContent(current + snippet)
		},

		ShowManual: func() { menu.ShowManualDialog(w) },
		ShowAbout:  func() { menu.ShowAboutDialog(w) },
	})
	w.SetMainMenu(mainMenu)

	updateLineNums("")
	w.SetContent(content)

	// Load last opened file on startup
	storageMgr.LoadLastFile()

	// Autosave timer - save every 30 seconds if file is set
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			storageMgr.AutoSave()
		}
	}()

	// Handle window close - autosave if file exists, warn only if no file specified
	w.SetCloseIntercept(func() {
		if storageMgr.CurrentFile() != "" {
			// File exists - autosave and close
			storageMgr.AutoSave()
			w.Close()
		} else if storageMgr.HasUnsavedChanges() {
			// No file specified but has unsaved changes - warn user
			dialog.ShowConfirm("Unsaved Changes",
				"You have unsaved changes that will be lost. Do you want to save before closing?",
				func(save bool) {
					if save {
						storageMgr.SaveAs()
					} else {
						w.Close()
					}
				}, w)
		} else {
			w.Close()
		}
	})

	w.ShowAndRun()
}
