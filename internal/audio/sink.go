package audio

import (
	"errors"
	"strings"
)

// ErrFeedbackRisk is returned when a capture source is the monitor of an enabled playback device.
var ErrFeedbackRisk = errors.New("feedback loop detected: capture source is the monitor of an enabled output device")

// VirtualSinkManager manages the lifecycle of a virtual audio output sink (FR-7, FR-11).
// On Linux, it routes desktop audio into a null sink monitor and restores defaults on cleanup.
type VirtualSinkManager interface {
	// Setup creates the virtual sink, records the previous default sink,
	// and routes system audio to the virtual sink.
	Setup() error

	// Restore sets the system default sink back to the original sink
	// and removes the virtual sink.
	Restore() error

	// VirtualSinkName returns the name of the created virtual sink (e.g. "multiout").
	VirtualSinkName() string

	// IsActive reports whether the virtual sink is currently active.
	IsActive() bool

	// CleanupStale removes any leftover virtual sink nodes from a previous abnormal termination.
	CleanupStale() error
}

// CheckFeedbackRisk verifies whether a capture device is the monitor or loopback source
// of a specified playback device.
func CheckFeedbackRisk(captureDev Device, playbackDev Device) bool {
	capName := strings.ToLower(captureDev.Name)
	playName := strings.ToLower(playbackDev.Name)

	// 1. Linux monitor check
	if strings.Contains(capName, "monitor") {
		return strings.Contains(capName, playName)
	}

	// 2. Windows device-specific loopback check
	if strings.Contains(capName, "loopback") {
		cleanCap := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(capName, "(loopback)", ""), "[loopback]", ""))
		if cleanCap != "" && (strings.Contains(cleanCap, playName) || strings.Contains(playName, cleanCap)) {
			return true
		}
	}

	// 3. System Audio Default loopback check
	if captureDev.ID == "loopback:default" && playbackDev.IsDefault {
		return true
	}

	return false
}
