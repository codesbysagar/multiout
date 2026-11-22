package ui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multiout/internal/audio"
	"multiout/internal/config"
	"multiout/internal/engine"
)

// DeviceCard represents an output playback device card.
type DeviceCard struct {
	dev        audio.Device
	controller *Controller

	container    *fyne.Container
	bg           *canvas.Rectangle
	nameLabel    *widget.Label
	toggle       *ToggleSwitch
	volumeSlider *widget.Slider
	volumeValue  *widget.Label
	delaySlider  *widget.Slider
	delayValue   *widget.Label
	meter        *LevelMeter
	warnLabel    *widget.Label

	delayMs int
	volume  float32
	enabled bool
	ready   bool
}

// NewDeviceCard builds an output device card.
func NewDeviceCard(dev audio.Device, initialCfg config.DeviceConfig, ctrl *Controller) *DeviceCard {
	card := &DeviceCard{
		dev:        dev,
		controller: ctrl,
		volume:     initialCfg.Volume,
		delayMs:    initialCfg.DelayMs,
		enabled:    initialCfg.Enabled,
	}

	p := GetPalette()

	// 1. Header Row: Output Device Name & Toggle Switch
	title := dev.Name
	if dev.IsDefault {
		title += " (Default)"
	}
	card.nameLabel = widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	card.toggle = NewToggleSwitch(initialCfg.Enabled, func(checked bool) {
		if !card.ready {
			return
		}
		card.enabled = checked
		card.updateBorderHighlight()
		card.syncConfig()
		if ctrl.IsRunning() {
			_ = ctrl.engine.SetOutputEnabled(dev.ID, checked)
			ctrl.Logger.Log("[Output] %s -> enabled: %v", dev.Name, checked)
		}
	})
	card.toggle.ActiveColor = p.Primary

	headerRow := container.NewBorder(
		nil, nil,
		nil,
		card.toggle,
		card.nameLabel,
	)

	// 2. Compact Volume Row with Decibel (dB) Control
	initialDB := engine.GainToDB(initialCfg.Volume)
	card.volumeValue = widget.NewLabelWithStyle(FormatDB(initialDB), fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})

	card.volumeSlider = widget.NewSlider(-60.0, 12.0)
	card.volumeSlider.Step = 0.5
	card.volumeSlider.Value = initialDB
	card.volumeSlider.OnChanged = func(val float64) {
		if !card.ready {
			return
		}
		card.volume = engine.DBToGain(val)
		card.volumeValue.SetText(FormatDB(val))
		card.syncConfig()
		if ctrl.IsRunning() {
			_ = ctrl.engine.SetOutputVolume(dev.ID, card.volume)
		}
	}

	volRow := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("Volume", fyne.TextAlignLeading, fyne.TextStyle{}),
		card.volumeValue,
		card.volumeSlider,
	)

	// 3. Compact Delay Row
	card.delayValue = widget.NewLabelWithStyle(fmt.Sprintf("%3d ms", initialCfg.DelayMs), fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})

	card.delaySlider = widget.NewSlider(0, 300)
	card.delaySlider.Step = 1.0
	card.delaySlider.Value = float64(initialCfg.DelayMs)
	card.delaySlider.OnChanged = func(val float64) {
		if !card.ready {
			return
		}
		card.delayMs = int(val)
		card.delayValue.SetText(fmt.Sprintf("%3d ms", card.delayMs))
		card.syncConfig()
		if ctrl.IsRunning() {
			_ = ctrl.engine.SetOutputDelay(dev.ID, card.delayMs)
		}
	}

	delayRow := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("Delay", fyne.TextAlignLeading, fyne.TextStyle{}),
		card.delayValue,
		card.delaySlider,
	)

	// 4. Slim Level Meter & Optional Warning
	card.meter = NewLevelMeter()
	card.warnLabel = widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	meterRow := container.NewBorder(
		nil, nil,
		nil,
		card.warnLabel,
		card.meter,
	)

	card.ready = true

	// Card frame: rounded 10px, subtle visible border
	card.bg = canvas.NewRectangle(p.CardBg)
	card.bg.CornerRadius = 10
	card.bg.StrokeWidth = 1
	card.updateBorderHighlight()

	cardContent := container.NewVBox(
		headerRow,
		volRow,
		delayRow,
		meterRow,
	)

	card.container = container.NewStack(
		card.bg,
		container.NewPadded(cardContent),
	)

	return card
}

func (c *DeviceCard) updateBorderHighlight() {
	p := GetPalette()
	if c.enabled {
		c.bg.StrokeColor = p.Primary
		c.toggle.ActiveColor = p.Primary
	} else {
		c.bg.StrokeColor = p.CardBorder
	}
	c.bg.Refresh()
}

// UpdateTheme refreshes the card's colors for dark/light mode.
func (c *DeviceCard) UpdateTheme() {
	p := GetPalette()
	c.bg.FillColor = p.CardBg
	c.updateBorderHighlight()
	c.toggle.Refresh()
}

// Container returns the Fyne CanvasObject for layout.
func (c *DeviceCard) Container() fyne.CanvasObject {
	return c.container
}

// UpdateStatus updates the meter and warnings from live engine metrics.
func (c *DeviceCard) UpdateStatus(peak float32, underruns, overruns uint64, active bool) {
	c.meter.SetLevel(peak)
	if !active {
		c.warnLabel.SetText("")
		return
	}
	if underruns > 0 {
		c.warnLabel.SetText(fmt.Sprintf("⚠️ %d underrun(s)", underruns))
	} else {
		c.warnLabel.SetText("")
	}
}

func (c *DeviceCard) syncConfig() {
	cfg := config.DeviceConfig{
		Enabled: c.enabled,
		Volume:  c.volume,
		Muted:   false,
		DelayMs: c.delayMs,
	}
	c.controller.UpdateDeviceConfig(c.dev.Name, cfg)
}

// FormatDB formats decibels for UI display (e.g., "-inf dB", "+3.5 dB", " 0.0 dB").
func FormatDB(db float64) string {
	if db <= -59.5 {
		return "-inf dB"
	}
	if math.Abs(db) < 0.05 {
		return " 0.0 dB"
	}
	if db > 0 {
		return fmt.Sprintf("+%.1f dB", db)
	}
	return fmt.Sprintf("%.1f dB", db)
}
