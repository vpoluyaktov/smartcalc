package menu

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Callbacks holds all menu action callbacks.
type Callbacks struct {
	New        func()
	Open       func()
	Save       func()
	SaveAs     func()
	OpenRecent func(path string)
	GetRecent  func() []string

	Cut   func()
	Copy  func()
	Paste func()

	InsertSnippet func(snippet string)

	ShowManual func()
	ShowAbout  func()
}

// Snippets contains example expressions.
var Snippets = []struct {
	Name    string
	Content string
}{
	{"Basic Math", "10 + 20 * 3 ="},
	{"Percentage", "$100 - 20% ="},
	{"Currency Calculation", "$1,500.00 + $250.50 ="},
	{"Line Reference", "100 =\n\\1 * 2 ="},
	{"Scientific", "sin(45) + cos(30) ="},
	{"Complex Expression", "$1,000 x 12 - 15% + $500 ="},
}

// CreateMainMenu creates the application main menu.
func CreateMainMenu(window fyne.Window, cb Callbacks) *fyne.MainMenu {
	// File menu
	newItem := fyne.NewMenuItem("New", cb.New)

	openItem := fyne.NewMenuItem("Open...", cb.Open)
	saveItem := fyne.NewMenuItem("Save", cb.Save)
	saveAsItem := fyne.NewMenuItem("Save As...", cb.SaveAs)

	recentMenu := fyne.NewMenuItem("Recent", nil)
	recentMenu.ChildMenu = buildRecentMenu(cb.GetRecent, cb.OpenRecent)

	fileMenu := fyne.NewMenu("File",
		newItem,
		fyne.NewMenuItemSeparator(),
		openItem,
		saveItem,
		saveAsItem,
		fyne.NewMenuItemSeparator(),
		recentMenu,
	)

	// Edit menu
	cutItem := fyne.NewMenuItem("Cut", cb.Cut)
	copyItem := fyne.NewMenuItem("Copy", cb.Copy)
	pasteItem := fyne.NewMenuItem("Paste", cb.Paste)

	editMenu := fyne.NewMenu("Edit",
		cutItem,
		copyItem,
		pasteItem,
	)

	// Snippets menu
	snippetItems := make([]*fyne.MenuItem, len(Snippets))
	for i, s := range Snippets {
		snippet := s.Content
		snippetItems[i] = fyne.NewMenuItem(s.Name, func() {
			cb.InsertSnippet(snippet)
		})
	}
	snippetsMenu := fyne.NewMenu("Snippets", snippetItems...)

	// Help menu
	manualItem := fyne.NewMenuItem("Manual", cb.ShowManual)
	aboutItem := fyne.NewMenuItem("About", cb.ShowAbout)

	helpMenu := fyne.NewMenu("Help",
		manualItem,
		fyne.NewMenuItemSeparator(),
		aboutItem,
	)

	return fyne.NewMainMenu(fileMenu, editMenu, snippetsMenu, helpMenu)
}

func buildRecentMenu(getRecent func() []string, openRecent func(string)) *fyne.Menu {
	recent := getRecent()
	if len(recent) == 0 {
		return fyne.NewMenu("",
			fyne.NewMenuItem("(No recent files)", nil),
		)
	}

	items := make([]*fyne.MenuItem, len(recent))
	for i, path := range recent {
		p := path
		items[i] = fyne.NewMenuItem(p, func() {
			openRecent(p)
		})
	}
	return fyne.NewMenu("", items...)
}

// ShowManualDialog displays the manual/help dialog.
func ShowManualDialog(window fyne.Window) {
	content := widget.NewRichTextFromMarkdown(`# SuperCalc Manual

## Basic Usage
Type mathematical expressions followed by = to calculate results.

## Supported Operations
- **Addition**: +
- **Subtraction**: -
- **Multiplication**: * or x
- **Division**: /
- **Power**: ^
- **Parentheses**: ( )

## Percentages
- 100 + 20% = 120 (adds 20% of 100)
- 100 - 20% = 80 (subtracts 20% of 100)

## Currency
- Prefix numbers with $ for currency formatting
- Supports thousands separators: $1,500.00

## Line References
- Use \1, \2, etc. to reference results from previous lines
- Example:
  - Line 1: 100 + 50 = 150
  - Line 2: \1 * 2 = 300

## Functions
- sin(), cos(), tan()
- sqrt(), abs()
- log(), ln()
- floor(), ceil(), round()
`)
	content.Wrapping = fyne.TextWrapWord

	d := dialog.NewCustom("SuperCalc Manual", "Close", content, window)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// ShowAboutDialog displays the about dialog.
func ShowAboutDialog(window fyne.Window) {
	dialog.ShowInformation("About SuperCalc",
		"SuperCalc v1.0\n\nA powerful calculator with support for:\n• Multi-line expressions\n• Line references\n• Currency formatting\n• Mathematical functions",
		window)
}
