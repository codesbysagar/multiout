package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed multiout-logo-rounded.png
var LogoData []byte

// LogoResource is the embedded static resource for the rounded application logo.
var LogoResource = fyne.NewStaticResource("multiout-logo.png", LogoData)
