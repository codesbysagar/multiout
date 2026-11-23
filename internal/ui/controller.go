package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"multiout/internal/audio"
	"multiout/internal/config"
	"multiout/internal/engine"
)

// Controller manages the application workflow between audio backend, engine, and UI.
type Controller struct {
	mu         sync.Mutex
	backend    audio.Backend
	engine     *engine.Engine
	sinkMgr    audio.VirtualSinkManager
	cfg        *config.Config
	running    bool
	stopPollCh chan struct{}

	playbackDevs []audio.Device
	captureDevs  []audio.Device
	selectedCap  audio.Device

	cards       map[string]*DeviceCard
	sourceCards map[string]*SourceCard

	Logger *ExecutionLogger

	OnStatusChange          func(status string, isErr bool)
	OnSelectedCaptureChange func(dev audio.Device)
}

// NewController initializes a UI controller with execution logging.
func NewController(be audio.Backend, eng *engine.Engine, sinkMgr audio.VirtualSinkManager, cfg *config.Config) *Controller {
	ctrl := &Controller{
		backend:     be,
		engine:      eng,
		sinkMgr:     sinkMgr,
		cfg:         cfg,
		cards:       make(map[string]*DeviceCard),
		sourceCards: make(map[string]*SourceCard),
		Logger:      NewExecutionLogger(300),
	}
	ctrl.Logger.Log("[System] Multiout controller initialized.")
	return ctrl
}

// RefreshDevices queries audio endpoints and updates the internal cache.
func (c *Controller) RefreshDevices() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	pDevs, err := c.backend.Devices(audio.Playback)
	if err != nil {
		c.Logger.Log("[Audio] Error listing playback devices: %v", err)
		return fmt.Errorf("list playback devices: %w", err)
	}

	// Filter out the virtual sink from playback devices to prevent circular audio routing (FR-8)
	filteredPlayDevs := make([]audio.Device, 0, len(pDevs))
	for _, d := range pDevs {
		lower := strings.ToLower(d.Name)
		if strings.Contains(lower, "multiout") || strings.Contains(lower, "null output") {
			continue
		}
		filteredPlayDevs = append(filteredPlayDevs, d)
	}
	c.playbackDevs = filteredPlayDevs

	cDevs, err := c.backend.Devices(audio.Capture)
	if err != nil {
		c.Logger.Log("[Audio] Error listing capture devices: %v", err)
		return fmt.Errorf("list capture devices: %w", err)
	}
	c.captureDevs = cDevs

	// Default capture selection:
	if c.selectedCap.ID == "" && len(cDevs) > 0 {
		// Prefer monitor of default sink, system audio, or first monitor/loopback
		for _, d := range cDevs {
			lower := strings.ToLower(d.Name)
			if d.IsDefault || strings.Contains(lower, "system audio") || strings.Contains(lower, "monitor") {
				c.selectedCap = d
				break
			}
		}
		if c.selectedCap.ID == "" {
			c.selectedCap = cDevs[0]
		}
	}

	if vSink, found := audio.DetectVirtualSink(c.playbackDevs); found {
		if winSinkMgr, ok := c.sinkMgr.(*audio.WindowsVirtualSinkManager); ok {
			winSinkMgr.SetDetectedSink(vSink.Name)
		}
		c.Logger.Log("[Audio] Virtual audio sink detected: \"%s\"", vSink.Name)
	}

	c.Logger.Log("[Audio] Discovered %d playback and %d capture devices.", len(pDevs), len(cDevs))
	return nil
}

// HasVirtualSink reports whether a virtual audio sink (e.g. VB-Cable or PipeWire) is active or available.
func (c *Controller) HasVirtualSink() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sinkMgr != nil && c.sinkMgr.VirtualSinkName() != "" {
		return true
	}
	_, found := audio.DetectVirtualSink(c.playbackDevs)
	return found
}

// VirtualSinkName returns the name of any detected or active virtual audio sink.
func (c *Controller) VirtualSinkName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sinkMgr != nil && c.sinkMgr.VirtualSinkName() != "" {
		return c.sinkMgr.VirtualSinkName()
	}
	if vSink, found := audio.DetectVirtualSink(c.playbackDevs); found {
		return vSink.Name
	}
	return ""
}

// OpenSoundSettings launches the system sound settings dialog.
func (c *Controller) OpenSoundSettings() error {
	return OpenSystemSoundSettings()
}

// PlaybackDevices returns currently enumerated playback devices.
func (c *Controller) PlaybackDevices() []audio.Device {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]audio.Device, len(c.playbackDevs))
	copy(out, c.playbackDevs)
	return out
}

// CaptureDevices returns currently enumerated capture devices.
func (c *Controller) CaptureDevices() []audio.Device {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]audio.Device, len(c.captureDevs))
	copy(out, c.captureDevs)
	return out
}

// SelectedCapture returns the active capture device.
func (c *Controller) SelectedCapture() audio.Device {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.selectedCap
}

// SetSelectedCapture sets the active capture source and updates UI switches.
func (c *Controller) SetSelectedCapture(dev audio.Device) {
	c.mu.Lock()
	c.selectedCap = dev
	c.cfg.CaptureSource = dev.Name
	_ = c.cfg.Save()

	// Update all source cards so only this one is selected
	for id, card := range c.sourceCards {
		card.SetSelected(id == dev.ID)
	}
	notifyFn := c.OnSelectedCaptureChange
	c.mu.Unlock()

	c.Logger.Log("[Input] Active source selected: \"%s\"", dev.Name)
	if notifyFn != nil {
		notifyFn(dev)
	}
}

// RegisterSourceCard tracks an input source card for radio single-selection.
func (c *Controller) RegisterSourceCard(deviceID string, card *SourceCard) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sourceCards[deviceID] = card
}

// ClearSourceCards clears registered source cards.
func (c *Controller) ClearSourceCards() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sourceCards = make(map[string]*SourceCard)
}

// IsRunning reports whether multi-output streaming is active.
func (c *Controller) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// UpdateDeviceConfig persists changes from UI cards.
func (c *Controller) UpdateDeviceConfig(name string, dcfg config.DeviceConfig) {
	c.cfg.SetDeviceConfig(name, dcfg)
	_ = c.cfg.Save()
}

// ThemeMode returns the saved theme mode preference.
func (c *Controller) ThemeMode() ThemeMode {
	return ThemeMode(c.cfg.GetThemeMode())
}

// SetThemeMode updates and persists the theme mode preference.
func (c *Controller) SetThemeMode(mode ThemeMode) {
	c.cfg.SetThemeMode(string(mode))
	_ = c.cfg.Save()
}

// Start begins capture and routes audio to enabled devices.
func (c *Controller) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if c.sinkMgr.VirtualSinkName() != "" {
		c.Logger.Log("[Engine] Setting up %s virtual sink...", c.sinkMgr.VirtualSinkName())
	} else {
		c.Logger.Log("[Engine] Initializing audio loopback capture...")
	}
	// 1. Setup virtual sink (FR-7)
	if err := c.sinkMgr.Setup(); err != nil {
		c.Logger.Log("[Error] Virtual sink setup failed: %v", err)
		if c.OnStatusChange != nil {
			c.OnStatusChange(fmt.Sprintf("Failed to setup virtual sink: %v", err), true)
		}
		return err
	}

	// 2. Refresh devices so the virtual sink monitor appears
	_ = c.RefreshDevices()

	c.mu.Lock()
	captureToUse := c.selectedCap
	for _, d := range c.captureDevs {
		if strings.Contains(strings.ToLower(d.Name), "monitor") && strings.Contains(strings.ToLower(d.Name), "multiout") {
			captureToUse = d
			break
		}
	}

	c.Logger.Log("[Engine] Starting capture on \"%s\"...", captureToUse.Name)
	// 3. Start engine capture
	if err := c.engine.StartCapture(captureToUse); err != nil {
		c.mu.Unlock()
		_ = c.sinkMgr.Restore()
		c.Logger.Log("[Error] Capture startup failed: %v", err)
		if c.OnStatusChange != nil {
			c.OnStatusChange(fmt.Sprintf("Capture error: %v", err), true)
		}
		return err
	}

	// 4. Attach outputs
	activeCount := 0
	for _, dev := range c.playbackDevs {
		lower := strings.ToLower(dev.Name)
		if strings.Contains(lower, "multiout") || strings.Contains(lower, "null output") || audio.CheckFeedbackRisk(captureToUse, dev) {
			c.Logger.Log("[Warning] Skipping output \"%s\" to prevent feedback loop", dev.Name)
			continue
		}
		cardCfg := c.cfg.GetDeviceConfig(dev.Name)
		pipeline, err := c.engine.AddOutput(dev)
		if err != nil {
			c.Logger.Log("[Warning] Failed to add output \"%s\": %v", dev.Name, err)
			continue
		}
		pipeline.SetVolume(cardCfg.Volume)
		pipeline.SetMute(cardCfg.Muted)
		pipeline.SetDelay(cardCfg.DelayMs)
		_ = c.engine.SetOutputEnabled(dev.ID, cardCfg.Enabled)
		if cardCfg.Enabled {
			activeCount++
			c.Logger.Log("[Output] Attached \"%s\" (vol=%.1f dB, delay=%dms)", dev.Name, engine.GainToDB(cardCfg.Volume), cardCfg.DelayMs)
		}
	}

	c.running = true
	c.stopPollCh = make(chan struct{})
	c.mu.Unlock()

	// Start ~20Hz status / meter polling loop
	go c.pollStatsLoop(c.stopPollCh)

	c.Logger.Log("[Engine] Multiout active. Streaming to %d enabled outputs.", activeCount)
	if c.OnStatusChange != nil {
		c.OnStatusChange(fmt.Sprintf("Running (%d outputs active)", activeCount), false)
	}
	return nil
}

// Stop cleanly terminates streaming, restores desktop audio, and saves settings.
func (c *Controller) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	if c.stopPollCh != nil {
		close(c.stopPollCh)
		c.stopPollCh = nil
	}
	c.mu.Unlock()

	c.Logger.Log("[Engine] Stopping streaming engine...")
	_ = c.engine.StopCapture()
	_ = c.sinkMgr.Restore()
	_ = c.cfg.Save()
	c.Logger.Log("[System] Stopped. Audio routing restored to default.")

	// Reset card meters
	resetFn := func() {
		for _, card := range c.cards {
			card.UpdateStatus(0, 0, 0, false)
		}
		if c.OnStatusChange != nil {
			c.OnStatusChange("Stopped", false)
		}
	}
	if fyne.CurrentApp() != nil {
		fyne.Do(resetFn)
	} else {
		resetFn()
	}
}

func (c *Controller) pollStatsLoop(stopCh <-chan struct{}) {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 Hz
	defer ticker.Stop()

	var lastReportedUnderruns uint64

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			states := c.engine.Outputs()
			fyne.Do(func() {
				var totalUnderruns uint64
				for _, s := range states {
					if card, ok := c.cards[s.ID]; ok {
						card.UpdateStatus(s.Peak, s.Underruns, s.Overruns, s.Enabled)
					}
					totalUnderruns += s.Underruns
				}
				if totalUnderruns > lastReportedUnderruns {
					c.Logger.Log("[Warning] Underrun detected (%d total)", totalUnderruns)
					lastReportedUnderruns = totalUnderruns
				}
				if totalUnderruns > 0 && c.OnStatusChange != nil {
					c.OnStatusChange(fmt.Sprintf("Running (%d underruns detected)", totalUnderruns), false)
				}
			})
		}
	}
}

// RegisterCard associates an output device ID with its UI card.
func (c *Controller) RegisterCard(deviceID string, card *DeviceCard) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cards[deviceID] = card
}
