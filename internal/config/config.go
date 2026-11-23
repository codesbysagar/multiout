package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// DeviceConfig holds persisted settings for a single playback output device.
type DeviceConfig struct {
	Enabled bool    `json:"enabled"`
	Volume  float32 `json:"volume"`
	Muted   bool    `json:"muted"`
	DelayMs int     `json:"delay_ms"`
}

// Config stores application settings across runs (FR-10).
// Device settings are keyed by device name (not transient index).
type Config struct {
	mu            sync.RWMutex
	CaptureSource string                  `json:"capture_source"`
	ThemeMode     string                  `json:"theme_mode,omitempty"`
	Outputs       map[string]DeviceConfig `json:"outputs"`
}

// NewDefaultConfig returns an initialized Config with standard defaults.
func NewDefaultConfig() *Config {
	return &Config{
		ThemeMode: "dark",
		Outputs:   make(map[string]DeviceConfig),
	}
}

// ConfigFilePath returns the path to multiout/config.json in the user's config directory.
func ConfigFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "multiout", "config.json"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "multiout", "config.json"), nil
}

// Load reads the configuration from the user config directory.
// Returns a default Config if the file does not yet exist.
func Load() (*Config, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return NewDefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return NewDefaultConfig(), nil
		}
		return NewDefaultConfig(), err
	}

	cfg := NewDefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return NewDefaultConfig(), err
	}
	if cfg.Outputs == nil {
		cfg.Outputs = make(map[string]DeviceConfig)
	}
	return cfg, nil
}

// Save writes the configuration atomically to the user config directory.
func (c *Config) Save() error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write atomically via a temporary file in the same directory
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// GetThemeMode returns the configured theme ("dark", "light", or "system").
func (c *Config) GetThemeMode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.ThemeMode == "" {
		return "dark"
	}
	return c.ThemeMode
}

// SetThemeMode updates the active theme mode.
func (c *Config) SetThemeMode(mode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ThemeMode = mode
}

// GetDeviceConfig returns the saved configuration for a device, or defaults if not found.
func (c *Config) GetDeviceConfig(deviceName string) DeviceConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if dev, exists := c.Outputs[deviceName]; exists {
		return dev
	}
	return DeviceConfig{
		Enabled: true,
		Volume:  1.0,
		Muted:   false,
		DelayMs: 0,
	}
}

// SetDeviceConfig updates or adds settings for a specific device.
func (c *Config) SetDeviceConfig(deviceName string, dev DeviceConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Outputs == nil {
		c.Outputs = make(map[string]DeviceConfig)
	}
	c.Outputs[deviceName] = dev
}
