//go:build !windows

package ui

import (
	"os/exec"
)

// OpenSystemSoundSettings opens the system sound settings page on Linux/other platforms.
func OpenSystemSoundSettings() error {
	if _, err := exec.LookPath("pavucontrol"); err == nil {
		return exec.Command("pavucontrol").Start()
	}
	return exec.Command("gnome-control-center", "sound").Start()
}
