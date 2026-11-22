//go:build !linux

package audio

// NonLinuxVirtualSinkManager is a no-op manager for Windows / macOS
// where virtual sinks are not managed through pulse/pipewire.
type NonLinuxVirtualSinkManager struct{}

func NewVirtualSinkManager() VirtualSinkManager {
	return &NonLinuxVirtualSinkManager{}
}

func (m *NonLinuxVirtualSinkManager) VirtualSinkName() string {
	return ""
}

func (m *NonLinuxVirtualSinkManager) IsActive() bool {
	return false
}

func (m *NonLinuxVirtualSinkManager) Setup() error {
	return nil
}

func (m *NonLinuxVirtualSinkManager) Restore() error {
	return nil
}

func (m *NonLinuxVirtualSinkManager) CleanupStale() error {
	return nil
}

