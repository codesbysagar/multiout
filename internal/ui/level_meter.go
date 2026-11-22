package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// LevelMeter is a sleek, lightweight horizontal audio peak meter.
type LevelMeter struct {
	widget.BaseWidget
	level float32 // 0.0 to 1.0
}

// NewLevelMeter creates a new LevelMeter widget.
func NewLevelMeter() *LevelMeter {
	m := &LevelMeter{level: 0.0}
	m.ExtendBaseWidget(m)
	return m
}

// SetLevel updates the current meter peak level (0.0 to 1.0).
func (m *LevelMeter) SetLevel(level float32) {
	if level < 0 {
		level = 0
	} else if level > 1 {
		level = 1
	}
	if m.level == level {
		return
	}
	m.level = level
	m.Refresh()
}

// CreateRenderer creates the canvas objects for the LevelMeter.
func (m *LevelMeter) CreateRenderer() fyne.WidgetRenderer {
	p := GetPalette()
	bg := canvas.NewRectangle(p.WellBg)
	bg.CornerRadius = 2

	bar := canvas.NewRectangle(p.Accent)
	bar.CornerRadius = 2

	return &levelMeterRenderer{
		meter: m,
		bg:    bg,
		bar:   bar,
	}
}

type levelMeterRenderer struct {
	meter *LevelMeter
	bg    *canvas.Rectangle
	bar   *canvas.Rectangle
}

func (r *levelMeterRenderer) Destroy() {}

func (r *levelMeterRenderer) Layout(size fyne.Size) {
	p := GetPalette()
	r.bg.Resize(size)
	r.bg.FillColor = p.WellBg

	barWidth := size.Width * float32(r.meter.level)
	if barWidth > size.Width {
		barWidth = size.Width
	}
	r.bar.Resize(fyne.NewSize(barWidth, size.Height))

	// Color grading based on level:
	if r.meter.level > 0.90 {
		r.bar.FillColor = color.NRGBA{R: 0xEF, G: 0x44, B: 0x44, A: 0xFF} // Red
	} else if r.meter.level > 0.70 {
		r.bar.FillColor = p.Primary // Accent Purple
	} else {
		r.bar.FillColor = p.Accent // Accent Teal
	}
	r.bar.Refresh()
	r.bg.Refresh()
}

func (r *levelMeterRenderer) MinSize() fyne.Size {
	return fyne.NewSize(40, 4)
}

func (r *levelMeterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.bar}
}

func (r *levelMeterRenderer) Refresh() {
	r.Layout(r.meter.Size())
	canvas.Refresh(r.meter)
}
