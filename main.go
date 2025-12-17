package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"supercalc/internal/calc"
	"supercalc/internal/eval"
	"supercalc/internal/menu"
	"supercalc/internal/storage"
	"supercalc/internal/ui"
)

// Version info set via ldflags
var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	a := app.NewWithID("com.supercalc.app")
	a.Settings().SetTheme(ui.NewLargerTextTheme())
	w := a.NewWindow("SuperCalc - Untitled")
	w.Resize(fyne.NewSize(1000, 700))

	lineNums := widget.NewLabel("1")
	lineNums.TextStyle = fyne.TextStyle{Monospace: true}
	lineNums.Wrapping = fyne.TextWrapOff
	lineNums.Alignment = fyne.TextAlignTrailing

	entry := widget.NewMultiLineEntry()
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	entry.Wrapping = fyne.TextWrapOff
	entry.SetPlaceHolder("Type expressions like: ($95.88 x (167 + 175) - 20% =\nReference prior results as \\\\1, \\\\2, ...")

	lineNumBox := container.New(&ui.FixedWidthLayout{Width: 50}, container.NewStack(lineNums))
	editorArea := container.NewBorder(nil, nil, lineNumBox, nil, entry)

	// Status bar at bottom
	statusLabel := widget.NewLabel(fmt.Sprintf("SuperCalc %s (built %s, commit %s)", version, buildDate, gitCommit))
	statusLabel.Alignment = fyne.TextAlignCenter
	statusBar := container.NewCenter(statusLabel)

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

	// Storage manager for file operations
	storageMgr := storage.NewManager(a, w,
		func(content string) { setContent(content) },
		func() string { return entry.Text },
		func() { setContent("") },
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
		Copy:  func() { entry.TypedShortcut(&fyne.ShortcutCopy{Clipboard: w.Clipboard()}) },
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

	w.ShowAndRun()
}
