//go:build windows

package ui

import (
	"os/exec"
)

// OpenSystemSoundSettings opens the Windows Sound Settings page.
func OpenSystemSoundSettings() error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", "ms-settings:sound").Start()
}
