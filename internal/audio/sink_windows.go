//go:build windows

package audio

import (
	"strings"
	"sync"
)

// WindowsVirtualSinkManager detects and coordinates virtual audio devices (such as VB-Audio Virtual Cable) on Windows.
type WindowsVirtualSinkManager struct {
	mu       sync.Mutex
	active   bool
	sinkName string
}

// NewVirtualSinkManager creates the Windows virtual sink manager.
func NewVirtualSinkManager() VirtualSinkManager {
	return &WindowsVirtualSinkManager{}
}

func (m *WindowsVirtualSinkManager) VirtualSinkName() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sinkName
}

func (m *WindowsVirtualSinkManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

func (m *WindowsVirtualSinkManager) CleanupStale() error {
	return nil
}

// SetDetectedSink records the detected virtual sink name.
func (m *WindowsVirtualSinkManager) SetDetectedSink(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sinkName = name
}

func (m *WindowsVirtualSinkManager) Setup() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = true
	return nil
}

func (m *WindowsVirtualSinkManager) Restore() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = false
	return nil
}

// DetectVirtualSink checks a list of playback devices and returns the first virtual audio cable found.
func DetectVirtualSink(devs []Device) (Device, bool) {
	virtualKeywords := []string{
		"cable input",
		"vb-audio",
		"virtual audio",
		"virtual cable",
		"multiout",
	}
	for _, d := range devs {
		lower := strings.ToLower(d.Name)
		for _, kw := range virtualKeywords {
			if strings.Contains(lower, kw) {
				return d, true
			}
		}
	}
	return Device{}, false
}
