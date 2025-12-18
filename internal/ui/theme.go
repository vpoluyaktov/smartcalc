package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// SmartCalcTheme is a modern, visually appealing theme for SmartCalc.
type SmartCalcTheme struct {
	fyne.Theme
}

// NewLargerTextTheme creates a theme with larger text size and custom colors.
func NewLargerTextTheme() *SmartCalcTheme {
	return &SmartCalcTheme{Theme: theme.DefaultTheme()}
}

// Color palette - Modern dark theme with vibrant accents
var (
	// Primary colors
	primaryColor   = color.NRGBA{R: 99, G: 102, B: 241, A: 255} // Indigo-500
	secondaryColor = color.NRGBA{R: 139, G: 92, B: 246, A: 255} // Violet-500
	accentColor    = color.NRGBA{R: 34, G: 211, B: 238, A: 255} // Cyan-400

	// Background colors
	bgDark       = color.NRGBA{R: 15, G: 23, B: 42, A: 255}  // Slate-900
	bgMedium     = color.NRGBA{R: 30, G: 41, B: 59, A: 255}  // Slate-800
	bgLight      = color.NRGBA{R: 51, G: 65, B: 85, A: 255}  // Slate-700
	bgHover      = color.NRGBA{R: 71, G: 85, B: 105, A: 255} // Slate-600
	bgInputField = color.NRGBA{R: 24, G: 34, B: 52, A: 255}  // Custom dark

	// Text colors
	textPrimary   = color.NRGBA{R: 248, G: 250, B: 252, A: 255} // Slate-50
	textSecondary = color.NRGBA{R: 148, G: 163, B: 184, A: 255} // Slate-400
	textMuted     = color.NRGBA{R: 100, G: 116, B: 139, A: 255} // Slate-500

	// Status colors
	successColor = color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // Green-500
	errorColor   = color.NRGBA{R: 239, G: 68, B: 68, A: 255}  // Red-500
	warningColor = color.NRGBA{R: 251, G: 191, B: 36, A: 255} // Amber-400

	// Light theme variants
	bgLightTheme       = color.NRGBA{R: 248, G: 250, B: 252, A: 255} // Slate-50
	bgLightMedium      = color.NRGBA{R: 241, G: 245, B: 249, A: 255} // Slate-100
	bgLightInput       = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White
	textLightPrimary   = color.NRGBA{R: 15, G: 23, B: 42, A: 255}    // Slate-900
	textLightSecondary = color.NRGBA{R: 71, G: 85, B: 105, A: 255}   // Slate-600
)

func (t *SmartCalcTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 15
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameScrollBar:
		return 14
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 4
	}
	return t.Theme.Size(name)
}

func (t *SmartCalcTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return t.lightColor(name)
	}
	return t.darkColor(name)
}

func (t *SmartCalcTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return bgDark
	case theme.ColorNameButton:
		return bgLight
	case theme.ColorNameDisabledButton:
		return bgMedium
	case theme.ColorNameDisabled:
		return textMuted
	case theme.ColorNameError:
		return errorColor
	case theme.ColorNameFocus:
		return primaryColor
	case theme.ColorNameForeground:
		return textPrimary
	case theme.ColorNameForegroundOnError:
		return textPrimary
	case theme.ColorNameForegroundOnPrimary:
		return textPrimary
	case theme.ColorNameForegroundOnSuccess:
		return textPrimary
	case theme.ColorNameForegroundOnWarning:
		return bgDark
	case theme.ColorNameHeaderBackground:
		return bgMedium
	case theme.ColorNameHover:
		return bgHover
	case theme.ColorNameHyperlink:
		return accentColor
	case theme.ColorNameInputBackground:
		return bgInputField
	case theme.ColorNameInputBorder:
		return bgLight
	case theme.ColorNameMenuBackground:
		return bgMedium
	case theme.ColorNameOverlayBackground:
		return bgMedium
	case theme.ColorNamePlaceHolder:
		return textMuted
	case theme.ColorNamePressed:
		return secondaryColor
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameScrollBar:
		return bgLight
	case theme.ColorNameSelection:
		return color.NRGBA{R: 99, G: 102, B: 241, A: 100}
	case theme.ColorNameSeparator:
		return bgLight
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameSuccess:
		return successColor
	case theme.ColorNameWarning:
		return warningColor
	}
	return t.Theme.Color(name, theme.VariantDark)
}

func (t *SmartCalcTheme) lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return bgLightTheme
	case theme.ColorNameButton:
		return bgLightMedium
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 226, G: 232, B: 240, A: 255}
	case theme.ColorNameDisabled:
		return textLightSecondary
	case theme.ColorNameError:
		return errorColor
	case theme.ColorNameFocus:
		return primaryColor
	case theme.ColorNameForeground:
		return textLightPrimary
	case theme.ColorNameForegroundOnError:
		return textPrimary
	case theme.ColorNameForegroundOnPrimary:
		return textPrimary
	case theme.ColorNameForegroundOnSuccess:
		return textPrimary
	case theme.ColorNameForegroundOnWarning:
		return textLightPrimary
	case theme.ColorNameHeaderBackground:
		return bgLightMedium
	case theme.ColorNameHover:
		return color.NRGBA{R: 226, G: 232, B: 240, A: 255}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 79, G: 70, B: 229, A: 255}
	case theme.ColorNameInputBackground:
		return bgLightInput
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 203, G: 213, B: 225, A: 255}
	case theme.ColorNameMenuBackground:
		return bgLightInput
	case theme.ColorNameOverlayBackground:
		return bgLightInput
	case theme.ColorNamePlaceHolder:
		return textLightSecondary
	case theme.ColorNamePressed:
		return secondaryColor
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 203, G: 213, B: 225, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 99, G: 102, B: 241, A: 80}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 226, G: 232, B: 240, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 40}
	case theme.ColorNameSuccess:
		return successColor
	case theme.ColorNameWarning:
		return warningColor
	}
	return t.Theme.Color(name, theme.VariantLight)
}

func (t *SmartCalcTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.Theme.Font(style)
}

func (t *SmartCalcTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.Theme.Icon(name)
}
