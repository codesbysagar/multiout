package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multiout/internal/audio"
)

// MainWindow builds the main multiout application window.
type MainWindow struct {
	app        fyne.App
	win        fyne.Window
	ctrl       *Controller
	sourceBox  *fyne.Container
	outputBox  *fyne.Container
	startStop  *ToggleSwitch
	statusPill *widget.Label
	themeBtn   *widget.Button
	logConsole *LogConsole

	titleText    *canvas.Text
	subTitleText *canvas.Text
	startPill    *canvas.Rectangle
	inputBg      *canvas.Rectangle
	outputBg     *canvas.Rectangle
	cards        []*DeviceCard
	srcCards     []*SourceCard
}

// NewMainWindow creates and initializes the user interface.
func NewMainWindow(app fyne.App, ctrl *Controller) *MainWindow {
	win := app.NewWindow("Multiout")
	win.Resize(fyne.NewSize(960, 720))
	win.SetMaster()
	win.SetIcon(LogoResource)

	m := &MainWindow{
		app:       app,
		win:       win,
		ctrl:      ctrl,
		sourceBox: container.NewVBox(),
		outputBox: container.NewVBox(),
	}

	// Apply configured theme
	currentMode := ctrl.ThemeMode()
	SetActiveThemeMode(currentMode)
	app.Settings().SetTheme(NewMultioutTheme(currentMode))
	p := GetPalette()

	// -------------------------------------------------------------
	// 1. Creative, Bold Header with Rounded Logo & Uniform Alignment
	// -------------------------------------------------------------
	appIconBox := m.buildAppIcon()

	m.titleText = canvas.NewText("Multiout", p.TextPrimary)
	m.titleText.TextSize = 24
	m.titleText.TextStyle = fyne.TextStyle{Bold: true}

	m.subTitleText = canvas.NewText("Simultaneous Multi-Device Audio Output", p.TextMuted)
	m.subTitleText.TextSize = 13
	m.subTitleText.TextStyle = fyne.TextStyle{Italic: true}

	titleCol := container.NewVBox(m.titleText, m.subTitleText)
	leftBrand := container.NewHBox(appIconBox, container.NewCenter(titleCol))

	// Theme mode switcher
	m.themeBtn = widget.NewButton(m.themeButtonLabel(currentMode), func() {
		nextMode := m.nextThemeMode(GetActiveThemeMode())
		m.applyThemeMode(nextMode)
	})
	m.themeBtn.Importance = widget.LowImportance

	// Start / Stop switch & status badge inside a clean framed pill
	m.statusPill = widget.NewLabelWithStyle("○ STOPPED", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	m.startStop = NewToggleSwitch(false, func(start bool) {
		if start {
			if err := m.ctrl.Start(); err != nil {
				m.startStop.SetCheckedSilent(false)
				m.statusPill.SetText("⚠️ ERROR")
				return
			}
			m.statusPill.SetText("● RUNNING")
		} else {
			m.ctrl.Stop()
			m.statusPill.SetText("○ STOPPED")
		}
	})
	m.startStop.ActiveColor = p.Accent

	startPillRow := container.NewHBox(
		widget.NewLabelWithStyle("start / stop:", fyne.TextAlignTrailing, fyne.TextStyle{}),
		m.statusPill,
		container.NewCenter(m.startStop),
	)

	m.startPill = canvas.NewRectangle(p.CardBg)
	m.startPill.CornerRadius = 18
	m.startPill.StrokeWidth = 1
	m.startPill.StrokeColor = p.Border
	startPillBox := container.NewStack(m.startPill, container.NewPadded(startPillRow))

	rightControls := container.NewHBox(
		container.NewCenter(m.themeBtn),
		container.NewCenter(startPillBox),
	)

	headerRow := container.NewBorder(
		nil, nil,
		leftBrand,
		rightControls,
	)
	headerContainer := container.NewVBox(
		container.NewPadded(headerRow),
		widget.NewSeparator(),
	)

	// -------------------------------------------------------------
	// 2. Middle Body: Two Columns with Top/Bottom Margins
	// -------------------------------------------------------------
	// Left Column: Input Sources
	inputHeader := container.NewVBox(
		widget.NewLabelWithStyle("Input sources", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Only one switch can be enabled at a time", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewSeparator(),
	)

	inputScrollContent := container.NewBorder(
		m.createVerticalSpacer(10),
		m.createVerticalSpacer(10),
		m.createHorizontalSpacer(8),
		m.createHorizontalSpacer(8),
		m.sourceBox,
	)
	inputScroll := container.NewVScroll(inputScrollContent)

	inputContainer := container.NewBorder(
		inputHeader, nil, nil, nil,
		inputScroll,
	)
	m.inputBg = canvas.NewRectangle(p.ContainerBg)
	m.inputBg.CornerRadius = 12
	m.inputBg.StrokeColor = p.Border
	m.inputBg.StrokeWidth = 1
	leftCol := container.NewStack(m.inputBg, container.NewPadded(inputContainer))

	// Right Column: Output Sources (scrollable with margins)
	outputHeader := container.NewVBox(
		widget.NewLabelWithStyle("Output sources", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Multiple switches can be enabled in output sources", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewSeparator(),
	)

	outputScrollContent := container.NewBorder(
		m.createVerticalSpacer(10),
		m.createVerticalSpacer(10),
		m.createHorizontalSpacer(8),
		m.createHorizontalSpacer(8),
		m.outputBox,
	)
	outputScroll := container.NewVScroll(outputScrollContent)

	outputContainer := container.NewBorder(
		outputHeader, nil, nil, nil,
		outputScroll,
	)
	m.outputBg = canvas.NewRectangle(p.ContainerBg)
	m.outputBg.CornerRadius = 12
	m.outputBg.StrokeColor = p.Border
	m.outputBg.StrokeWidth = 1
	rightCol := container.NewStack(m.outputBg, container.NewPadded(outputContainer))

	bodySplit := container.NewHSplit(leftCol, rightCol)
	bodySplit.SetOffset(0.38)

	// -------------------------------------------------------------
	// 3. Bottom Panel: Execution Logs with Clear Button
	// -------------------------------------------------------------
	logTitle := widget.NewLabelWithStyle("Execution logs:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	clearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		ctrl.Logger.Clear()
	})
	clearBtn.Importance = widget.LowImportance

	logHeader := container.NewBorder(
		nil, nil,
		logTitle,
		clearBtn,
	)

	m.logConsole = NewLogConsole(ctrl.Logger)
	logPanelContent := container.NewBorder(
		logHeader, nil, nil, nil,
		m.logConsole.Container(),
	)

	mainSplit := container.NewVSplit(bodySplit, logPanelContent)
	mainSplit.SetOffset(0.68)

	windowContent := container.NewBorder(
		headerContainer,
		nil, nil, nil,
		mainSplit,
	)

	win.SetContent(container.NewPadded(windowContent))

	// Clean exit on window close
	win.SetCloseIntercept(func() {
		ctrl.Stop()
		win.Close()
	})

	// Initial device discovery & card population
	_ = ctrl.RefreshDevices()
	m.rebuildSourceCards()
	m.rebuildDeviceCards()

	return m
}

func (m *MainWindow) createVerticalSpacer(height float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(1, height))
	return rect
}

func (m *MainWindow) createHorizontalSpacer(width float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(width, 1))
	return rect
}

func (m *MainWindow) themeButtonLabel(mode ThemeMode) string {
	switch mode {
	case ThemeModeLight:
		return "☀️ Light"
	case ThemeModeSystem:
		return "🖥️ System"
	default:
		return "🌙 Dark"
	}
}

func (m *MainWindow) nextThemeMode(current ThemeMode) ThemeMode {
	switch current {
	case ThemeModeDark:
		return ThemeModeLight
	case ThemeModeLight:
		return ThemeModeSystem
	default:
		return ThemeModeDark
	}
}

func (m *MainWindow) applyThemeMode(mode ThemeMode) {
	SetActiveThemeMode(mode)
	m.ctrl.SetThemeMode(mode)
	m.app.Settings().SetTheme(NewMultioutTheme(mode))
	m.themeBtn.SetText(m.themeButtonLabel(mode))

	// Dynamically re-color all container backgrounds and borders
	p := GetPalette()
	m.titleText.Color = p.TextPrimary
	m.titleText.Refresh()

	m.subTitleText.Color = p.TextMuted
	m.subTitleText.Refresh()

	m.startPill.FillColor = p.CardBg
	m.startPill.StrokeColor = p.Border
	m.startPill.Refresh()

	m.inputBg.FillColor = p.ContainerBg
	m.inputBg.StrokeColor = p.Border
	m.inputBg.Refresh()

	m.outputBg.FillColor = p.ContainerBg
	m.outputBg.StrokeColor = p.Border
	m.outputBg.Refresh()

	m.logConsole.UpdateTheme()
	m.startStop.Refresh()

	for _, sc := range m.srcCards {
		sc.UpdateTheme()
	}
	for _, dc := range m.cards {
		dc.UpdateTheme()
	}

	m.win.Content().Refresh()
}

// ShowAndRun displays the window.
func (m *MainWindow) ShowAndRun() {
	m.win.ShowAndRun()
}

func (m *MainWindow) buildAppIcon() fyne.CanvasObject {
	img := canvas.NewImageFromResource(LogoResource)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(48, 48))
	return img
}

func (m *MainWindow) rebuildSourceCards() {
	m.sourceBox.Objects = nil
	m.srcCards = nil
	m.ctrl.ClearSourceCards()

	caps := m.ctrl.CaptureDevices()
	if len(caps) == 0 {
		m.sourceBox.Add(widget.NewLabel("No capture/monitor devices detected."))
		m.sourceBox.Refresh()
		return
	}

	selectedCap := m.ctrl.SelectedCapture()
	for _, dev := range caps {
		isSelected := dev.ID == selectedCap.ID
		card := NewSourceCard(dev, isSelected, func(d audio.Device) {
			m.ctrl.SetSelectedCapture(d)
		})
		m.ctrl.RegisterSourceCard(dev.ID, card)
		m.srcCards = append(m.srcCards, card)
		m.sourceBox.Add(card.Container())
	}
	m.sourceBox.Refresh()
}

func (m *MainWindow) rebuildDeviceCards() {
	m.outputBox.Objects = nil
	m.cards = nil

	playDevs := m.ctrl.PlaybackDevices()
	if len(playDevs) == 0 {
		m.outputBox.Add(widget.NewLabel("No playback devices detected."))
		m.outputBox.Refresh()
		return
	}

	for _, dev := range playDevs {
		cardCfg := m.ctrl.cfg.GetDeviceConfig(dev.Name)
		card := NewDeviceCard(dev, cardCfg, m.ctrl)
		m.ctrl.RegisterCard(dev.ID, card)
		m.cards = append(m.cards, card)
		m.outputBox.Add(card.Container())
	}
	m.outputBox.Refresh()
}
