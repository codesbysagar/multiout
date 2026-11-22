package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaultsAndRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := NewDefaultConfig()
	cfg.CaptureSource = "Test Monitor"
	cfg.SetDeviceConfig("Earphones A", DeviceConfig{
		Enabled: true,
		Volume:  0.8,
		Muted:   false,
		DelayMs: 45,
	})
	cfg.SetDeviceConfig("Headset B", DeviceConfig{
		Enabled: false,
		Volume:  0.4,
		Muted:   true,
		DelayMs: 120,
	})

	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was written
	expectedPath := filepath.Join(tmpDir, "multiout", "config.json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}

	// Load config back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.CaptureSource != "Test Monitor" {
		t.Fatalf("expected CaptureSource %q, got %q", "Test Monitor", loaded.CaptureSource)
	}

	devA := loaded.GetDeviceConfig("Earphones A")
	if !devA.Enabled || devA.Volume != 0.8 || devA.Muted || devA.DelayMs != 45 {
		t.Fatalf("unexpected devA config: %+v", devA)
	}

	devB := loaded.GetDeviceConfig("Headset B")
	if devB.Enabled || devB.Volume != 0.4 || !devB.Muted || devB.DelayMs != 120 {
		t.Fatalf("unexpected devB config: %+v", devB)
	}

	// Non-existent device returns sane default
	devC := loaded.GetDeviceConfig("Unseen Device")
	if !devC.Enabled || devC.Volume != 1.0 || devC.Muted || devC.DelayMs != 0 {
		t.Fatalf("unexpected devC default config: %+v", devC)
	}
}
