package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// CustomEntry is a MultiLineEntry that allows custom copy behavior.
type CustomEntry struct {
	widget.Entry
	OnCopy func() // Called when copy is triggered (Ctrl+C or menu)
}

// NewCustomMultiLineEntry creates a new multi-line entry with custom copy support.
func NewCustomMultiLineEntry() *CustomEntry {
	e := &CustomEntry{}
	e.ExtendBaseWidget(e)
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapOff
	return e
}

// TypedShortcut intercepts shortcuts to handle custom copy.
func (e *CustomEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if _, ok := shortcut.(*fyne.ShortcutCopy); ok && e.OnCopy != nil {
		e.OnCopy()
		return
	}
	e.Entry.TypedShortcut(shortcut)
}
