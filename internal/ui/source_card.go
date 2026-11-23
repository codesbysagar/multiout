package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multiout/internal/audio"
)

// SourceCard represents an input capture source item.
type SourceCard struct {
	dev        audio.Device
	container  *fyne.Container
	toggle     *ToggleSwitch
	nameLabel  *widget.Label
	bg         *canvas.Rectangle
	onSelected func(dev audio.Device)
}

// NewSourceCard builds an input source card with a single-select toggle switch.
func NewSourceCard(dev audio.Device, isSelected bool, onSelected func(dev audio.Device)) *SourceCard {
	card := &SourceCard{
		dev:        dev,
		onSelected: onSelected,
	}

	title := dev.Name
	if dev.IsDefault && !strings.Contains(strings.ToLower(title), "default") {
		title += " (Default)"
	}
	card.nameLabel = widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: isSelected})

	p := GetPalette()
	card.toggle = NewToggleSwitch(isSelected, func(checked bool) {
		if !checked {
			card.toggle.SetCheckedSilent(true)
			return
		}
		card.nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		card.nameLabel.Refresh()
		if card.onSelected != nil {
			card.onSelected(dev)
		}
	})
	card.toggle.ActiveColor = p.Accent

	card.bg = canvas.NewRectangle(p.CardBg)
	card.bg.CornerRadius = 8
	card.bg.StrokeWidth = 1
	if isSelected {
		card.bg.StrokeColor = p.Accent
	} else {
		card.bg.StrokeColor = p.CardBorder
	}

	row := container.NewBorder(
		nil, nil,
		nil,
		card.toggle,
		card.nameLabel,
	)

	card.container = container.NewStack(card.bg, container.NewPadded(row))
	return card
}

// Container returns the Fyne CanvasObject.
func (sc *SourceCard) Container() fyne.CanvasObject {
	return sc.container
}

// UpdateTheme refreshes the card's colors for dark/light mode.
func (sc *SourceCard) UpdateTheme() {
	p := GetPalette()
	sc.bg.FillColor = p.CardBg
	if sc.toggle.Checked {
		sc.bg.StrokeColor = p.Accent
		sc.toggle.ActiveColor = p.Accent
	} else {
		sc.bg.StrokeColor = p.CardBorder
	}
	sc.bg.Refresh()
	sc.toggle.Refresh()
}

// SetSelected updates visual highlight and toggle state.
func (sc *SourceCard) SetSelected(selected bool) {
	p := GetPalette()
	sc.toggle.SetCheckedSilent(selected)
	sc.nameLabel.TextStyle = fyne.TextStyle{Bold: selected}
	sc.nameLabel.Refresh()

	if selected {
		sc.bg.StrokeColor = p.Accent
	} else {
		sc.bg.StrokeColor = p.CardBorder
	}
	sc.bg.Refresh()
}
