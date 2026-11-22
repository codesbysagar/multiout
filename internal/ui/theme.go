package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeMode specifies the active appearance mode.
type ThemeMode string

const (
	ThemeModeDark   ThemeMode = "dark"
	ThemeModeLight  ThemeMode = "light"
	ThemeModeSystem ThemeMode = "system"
)

var currentThemeMode = ThemeModeDark

// SetActiveThemeMode updates the active theme mode globally.
func SetActiveThemeMode(mode ThemeMode) {
	if mode == "" {
		mode = ThemeModeDark
	}
	currentThemeMode = mode
}

// GetActiveThemeMode returns the current active theme mode.
func GetActiveThemeMode() ThemeMode {
	return currentThemeMode
}

// Palette holds the semantic color scheme for the active mode.
// Styled after modern desktop app references (Downloader & OpenLAMT).
type Palette struct {
	WindowBg      color.Color
	ContainerBg   color.Color
	CardBg        color.Color
	WellBg        color.Color
	Border        color.Color
	CardBorder    color.Color
	TextPrimary   color.Color
	TextMuted     color.Color
	Primary       color.Color
	PrimaryHover  color.Color
	Accent        color.Color
	AccentHover   color.Color
	Error         color.Color
	SwitchOffPill color.Color
	IconBoxBg     color.Color
}

// GetPalette returns the color palette corresponding to the current theme mode.
func GetPalette() Palette {
	return GetPaletteForMode(currentThemeMode)
}

// GetPaletteForMode returns the palette for a specific mode.
func GetPaletteForMode(mode ThemeMode) Palette {
	isDark := true
	switch mode {
	case ThemeModeLight:
		isDark = false
	case ThemeModeDark:
		isDark = true
	case ThemeModeSystem:
		if fyne.CurrentApp() != nil {
			isDark = (fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark)
		}
	}

	if isDark {
		return Palette{
			WindowBg:      color.NRGBA{R: 0x13, G: 0x14, B: 0x1F, A: 0xFF}, // #13141F
			ContainerBg:   color.NRGBA{R: 0x18, G: 0x1A, B: 0x27, A: 0xFF}, // #181A27
			CardBg:        color.NRGBA{R: 0x20, G: 0x23, B: 0x36, A: 0xFF}, // #202336
			WellBg:        color.NRGBA{R: 0x14, G: 0x17, B: 0x24, A: 0xFF}, // #141724
			Border:        color.NRGBA{R: 0x26, G: 0x29, B: 0x3D, A: 0xFF}, // #26293D
			CardBorder:    color.NRGBA{R: 0x2F, G: 0x34, B: 0x4C, A: 0xFF}, // #2F344C
			TextPrimary:   color.NRGBA{R: 0xF3, G: 0xF4, B: 0xF6, A: 0xFF}, // #F3F4F6
			TextMuted:     color.NRGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF}, // #9CA3AF
			Primary:       color.NRGBA{R: 0x8B, G: 0x5C, B: 0xF6, A: 0xFF}, // #8B5CF6 (Purple)
			PrimaryHover:  color.NRGBA{R: 0x7C, G: 0x3A, B: 0xED, A: 0xFF},
			Accent:        color.NRGBA{R: 0x14, G: 0xB8, B: 0xA6, A: 0xFF}, // #14B8A6 (Teal)
			AccentHover:   color.NRGBA{R: 0x0D, G: 0x94, B: 0x88, A: 0xFF},
			Error:         color.NRGBA{R: 0xEF, G: 0x44, B: 0x44, A: 0xFF},
			SwitchOffPill: color.NRGBA{R: 0x2C, G: 0x30, B: 0x44, A: 0xFF},
			IconBoxBg:     color.NRGBA{R: 0x26, G: 0x22, B: 0x42, A: 0xFF}, // Purple-tinted icon box
		}
	}

	// Light Mode Palette (Clean, spacious, modern light)
	return Palette{
		WindowBg:      color.NRGBA{R: 0xF4, G: 0xF5, B: 0xFB, A: 0xFF}, // #F4F5FB
		ContainerBg:   color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, // #FFFFFF
		CardBg:        color.NRGBA{R: 0xF9, G: 0xFA, B: 0xFB, A: 0xFF}, // #F9FAFB
		WellBg:        color.NRGBA{R: 0x14, G: 0x17, B: 0x24, A: 0xFF}, // Keep dark console for logs
		Border:        color.NRGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 0xFF}, // #E5E7EB
		CardBorder:    color.NRGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 0xFF}, // #E5E7EB
		TextPrimary:   color.NRGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF}, // #111827
		TextMuted:     color.NRGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF}, // #6B7280
		Primary:       color.NRGBA{R: 0x7C, G: 0x3A, B: 0xED, A: 0xFF}, // #7C3AED
		PrimaryHover:  color.NRGBA{R: 0x6D, G: 0x28, B: 0xD9, A: 0xFF},
		Accent:        color.NRGBA{R: 0x0D, G: 0x94, B: 0x88, A: 0xFF}, // #0D9488
		AccentHover:   color.NRGBA{R: 0x0F, G: 0x76, B: 0x6E, A: 0xFF},
		Error:         color.NRGBA{R: 0xDC, G: 0x26, B: 0x26, A: 0xFF},
		SwitchOffPill: color.NRGBA{R: 0xD1, G: 0xD5, B: 0xDB, A: 0xFF},
		IconBoxBg:     color.NRGBA{R: 0xED, G: 0xE9, B: 0xFE, A: 0xFF}, // Light purple tint
	}
}

// MultioutTheme implements fyne.Theme.
type MultioutTheme struct {
	mode ThemeMode
}

// NewMultioutTheme creates a new MultioutTheme with the specified mode.
func NewMultioutTheme(mode ThemeMode) *MultioutTheme {
	if mode == "" {
		mode = ThemeModeDark
	}
	SetActiveThemeMode(mode)
	return &MultioutTheme{mode: mode}
}

func (t *MultioutTheme) IsDark(variant fyne.ThemeVariant) bool {
	switch t.mode {
	case ThemeModeLight:
		return false
	case ThemeModeDark:
		return true
	case ThemeModeSystem:
		return variant == theme.VariantDark
	default:
		return true
	}
}

func (t *MultioutTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	p := GetPaletteForMode(t.mode)
	if t.mode == ThemeModeSystem && variant != theme.VariantDark {
		p = GetPaletteForMode(ThemeModeLight)
	}

	switch name {
	case theme.ColorNamePrimary:
		return p.Primary

	case theme.ColorNameFocus:
		return p.Accent

	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x7C, G: 0x3A, B: 0xED, A: 0x33}

	case theme.ColorNameBackground:
		return p.WindowBg

	case theme.ColorNameInputBackground:
		return p.ContainerBg

	case theme.ColorNameButton:
		return p.CardBg

	case theme.ColorNameForeground:
		return p.TextPrimary

	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x14, G: 0xB8, B: 0xA6, A: 0x66}

	case theme.ColorNameSeparator:
		return p.Border

	case theme.ColorNameSuccess:
		return p.Accent

	case theme.ColorNameError:
		return p.Error

	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *MultioutTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *MultioutTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *MultioutTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 6.0
	}
	if name == theme.SizeNameScrollBar {
		return 8.0
	}
	return theme.DefaultTheme().Size(name)
}
