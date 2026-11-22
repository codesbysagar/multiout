package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"multiout/internal/audio"
	"multiout/internal/config"
	"multiout/internal/engine"
)

func TestExecutionLogger(t *testing.T) {
	logger := NewExecutionLogger(10)
	logger.Log("Test log %d", 1)
	logger.Log("Test log %d", 2)

	lines := logger.AllLines()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] == "" || lines[1] == "" {
		t.Errorf("expected non-empty lines, got %v", lines)
	}

	// Verify ring buffer capping
	for i := 3; i <= 15; i++ {
		logger.Log("Line %d", i)
	}
	lines = logger.AllLines()
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines capped, got %d", len(lines))
	}
}

func TestToggleSwitchState(t *testing.T) {
	test.NewApp()
	var notified bool
	var state bool
	sw := NewToggleSwitch(false, func(val bool) {
		notified = true
		state = val
	})

	if sw.Checked {
		t.Errorf("expected initial checked=false")
	}

	sw.SetChecked(true)
	if !notified || !state || !sw.Checked {
		t.Errorf("expected switch to be true and notified")
	}

	sw.SetCheckedSilent(false)
	if sw.Checked {
		t.Errorf("expected silent check to be false")
	}
}

func TestControllerSourceSingleSelect(t *testing.T) {
	test.NewApp()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fakeBe := audio.NewFakeBackend()
	cfg := config.NewDefaultConfig()
	eng := engine.NewEngine(fakeBe, audio.Config{SampleRate: 48000, Channels: 2})
	ctrl := NewController(fakeBe, eng, audio.NewVirtualSinkManager(), cfg)

	if err := ctrl.RefreshDevices(); err != nil {
		t.Fatalf("refresh devices failed: %v", err)
	}

	caps := ctrl.CaptureDevices()
	if len(caps) < 2 {
		t.Skip("not enough capture devices for test")
	}

	cardA := NewSourceCard(caps[0], true, func(d audio.Device) { ctrl.SetSelectedCapture(d) })
	cardB := NewSourceCard(caps[1], false, func(d audio.Device) { ctrl.SetSelectedCapture(d) })
	ctrl.RegisterSourceCard(caps[0].ID, cardA)
	ctrl.RegisterSourceCard(caps[1].ID, cardB)

	// Select B
	ctrl.SetSelectedCapture(caps[1])

	if !cardB.toggle.Checked {
		t.Errorf("expected card B to be checked")
	}
	if cardA.toggle.Checked {
		t.Errorf("expected card A to be unchecked (single-select rule)")
	}
}

func TestVirtualSinkExcludedFromPlayback(t *testing.T) {
	test.NewApp()
	fakeBe := audio.NewFakeBackend()
	// Add a virtual sink to playback devices
	fakeBe.SetDevices(audio.Playback, []audio.Device{
		{ID: "dev:1", Name: "Ryzen Audio Controller", Kind: audio.Playback},
		{ID: "dev:2", Name: "Equinox Phobos", Kind: audio.Playback},
		{ID: "dev:3", Name: "Multiout Virtual Sink", Kind: audio.Playback},
		{ID: "dev:4", Name: "Null Output", Kind: audio.Playback},
	})

	cfg := config.NewDefaultConfig()
	eng := engine.NewEngine(fakeBe, audio.Config{SampleRate: 48000, Channels: 2})
	ctrl := NewController(fakeBe, eng, audio.NewVirtualSinkManager(), cfg)

	if err := ctrl.RefreshDevices(); err != nil {
		t.Fatalf("refresh devices failed: %v", err)
	}

	playDevs := ctrl.PlaybackDevices()
	if len(playDevs) != 2 {
		t.Fatalf("expected 2 playback devices, got %d", len(playDevs))
	}
	for _, d := range playDevs {
		if d.Name == "Multiout Virtual Sink" || d.Name == "Null Output" {
			t.Errorf("virtual sink %s should have been filtered out of playback devices", d.Name)
		}
	}
}

func TestThemeModeSwitching(t *testing.T) {
	test.NewApp()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fakeBe := audio.NewFakeBackend()
	cfg := config.NewDefaultConfig()
	eng := engine.NewEngine(fakeBe, audio.Config{SampleRate: 48000, Channels: 2})
	ctrl := NewController(fakeBe, eng, audio.NewVirtualSinkManager(), cfg)

	if ctrl.ThemeMode() != ThemeModeDark {
		t.Errorf("expected default theme dark, got %v", ctrl.ThemeMode())
	}

	ctrl.SetThemeMode(ThemeModeLight)
	if ctrl.ThemeMode() != ThemeModeLight {
		t.Errorf("expected theme light, got %v", ctrl.ThemeMode())
	}

	themeDark := NewMultioutTheme(ThemeModeDark)
	themeLight := NewMultioutTheme(ThemeModeLight)
	if !themeDark.IsDark(0) {
		t.Errorf("expected themeDark to report dark")
	}
	if themeLight.IsDark(0) {
		t.Errorf("expected themeLight to report light")
	}
}
