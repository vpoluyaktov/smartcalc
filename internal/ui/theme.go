package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// LargerTextTheme wraps the default theme with a larger text size.
type LargerTextTheme struct {
	fyne.Theme
}

// NewLargerTextTheme creates a theme with larger text size.
func NewLargerTextTheme() *LargerTextTheme {
	return &LargerTextTheme{Theme: theme.DefaultTheme()}
}

func (t *LargerTextTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 16 // Increased from default ~14
	}
	return t.Theme.Size(name)
}

func (t *LargerTextTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, variant)
}

func (t *LargerTextTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.Theme.Font(style)
}

func (t *LargerTextTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.Theme.Icon(name)
}
