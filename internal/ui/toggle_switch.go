package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// DaisyUI Toggle Switch Dimensions (strictly bounded to prevent layout overflow)
const (
	switchWidth  = float32(40)
	switchHeight = float32(22)
	knobDiameter = float32(16)
	switchPad    = float32(3)
)

// ToggleSwitch is a modern pill-shaped toggle switch.
type ToggleSwitch struct {
	widget.BaseWidget

	Checked       bool
	OnChanged     func(bool)
	ActiveColor   color.Color
	InactiveColor color.Color
	KnobColor     color.Color
	Disabled      bool
}

// NewToggleSwitch creates a toggle switch.
func NewToggleSwitch(initial bool, onChanged func(bool)) *ToggleSwitch {
	sw := &ToggleSwitch{
		Checked:   initial,
		OnChanged: onChanged,
		KnobColor: color.White,
	}
	sw.ExtendBaseWidget(sw)
	return sw
}

// SetChecked programmatically updates the toggle state.
func (sw *ToggleSwitch) SetChecked(checked bool) {
	if sw.Checked == checked {
		return
	}
	sw.Checked = checked
	sw.Refresh()
	if sw.OnChanged != nil {
		sw.OnChanged(checked)
	}
}

// SetCheckedSilent updates state without triggering OnChanged callback.
func (sw *ToggleSwitch) SetCheckedSilent(checked bool) {
	if sw.Checked == checked {
		return
	}
	sw.Checked = checked
	sw.Refresh()
}

// Tapped handles click events on the switch.
func (sw *ToggleSwitch) Tapped(e *fyne.PointEvent) {
	if sw.Disabled {
		return
	}
	sw.SetChecked(!sw.Checked)
}

// MinSize reports the fixed dimensions of the switch.
func (sw *ToggleSwitch) MinSize() fyne.Size {
	return fyne.NewSize(switchWidth, switchHeight)
}

// CreateRenderer creates the visual canvas elements for the toggle switch.
func (sw *ToggleSwitch) CreateRenderer() fyne.WidgetRenderer {
	p := GetPalette()
	pill := canvas.NewRectangle(p.SwitchOffPill)
	pill.CornerRadius = switchHeight / 2
	pill.StrokeWidth = 1
	pill.StrokeColor = p.Border

	knob := canvas.NewCircle(color.White)

	return &toggleSwitchRenderer{
		sw:   sw,
		pill: pill,
		knob: knob,
	}
}

type toggleSwitchRenderer struct {
	sw   *ToggleSwitch
	pill *canvas.Rectangle
	knob *canvas.Circle
}

func (r *toggleSwitchRenderer) Destroy() {}

func (r *toggleSwitchRenderer) MinSize() fyne.Size {
	return fyne.NewSize(switchWidth, switchHeight)
}

func (r *toggleSwitchRenderer) Layout(size fyne.Size) {
	p := GetPalette()

	// Center the fixed-dimension switch inside whatever size the container allocated
	x := (size.Width - switchWidth) / 2
	if x < 0 {
		x = 0
	}
	y := (size.Height - switchHeight) / 2
	if y < 0 {
		y = 0
	}

	r.pill.Move(fyne.NewPos(x, y))
	r.pill.Resize(fyne.NewSize(switchWidth, switchHeight))
	r.pill.CornerRadius = switchHeight / 2

	var knobX float32
	if r.sw.Checked {
		knobX = x + switchWidth - knobDiameter - switchPad
	} else {
		knobX = x + switchPad
	}
	knobY := y + (switchHeight-knobDiameter)/2

	r.knob.Move(fyne.NewPos(knobX, knobY))
	r.knob.Resize(fyne.NewSize(knobDiameter, knobDiameter))

	if r.sw.Checked {
		if r.sw.ActiveColor != nil {
			r.pill.FillColor = r.sw.ActiveColor
		} else {
			r.pill.FillColor = p.Accent
		}
		r.pill.StrokeColor = r.pill.FillColor
	} else {
		if r.sw.InactiveColor != nil {
			r.pill.FillColor = r.sw.InactiveColor
		} else {
			r.pill.FillColor = p.SwitchOffPill
		}
		r.pill.StrokeColor = p.Border
	}

	if r.sw.KnobColor != nil {
		r.knob.FillColor = r.sw.KnobColor
	} else {
		r.knob.FillColor = color.White
	}

	r.pill.Refresh()
	r.knob.Refresh()
}

func (r *toggleSwitchRenderer) Refresh() {
	r.Layout(r.sw.Size())
	canvas.Refresh(r.sw)
}

func (r *toggleSwitchRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.pill, r.knob}
}
