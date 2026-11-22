//go:build linux

package audio

import (
	"testing"
)

func TestParseDefaultSinkFromWpctl(t *testing.T) {
	sampleWpctlOutput := `
Audio
 ├─ Devices:
 │      48. HDA NVidia                          [alsa]
 │      52. Ryzen HD Audio Controller           [alsa]
 │      95. Equinox Phobos                      [alsa]
 │  
 ├─ Sinks:
 │      58. HDA NVidia Digital Stereo (HDMI)    [vol: 0.40]
 │      59. Ryzen HD Audio Controller Analog Stereo [vol: 0.57]
 │  *   99. Equinox Phobos Analog Stereo        [vol: 0.62]
 │  
 ├─ Sources:
 │  *   60. Ryzen HD Audio Controller Analog Stereo [vol: 0.27]
 │     100. Equinox Phobos Mono                 [vol: 1.00]
`
	id, err := parseDefaultSinkFromWpctl(sampleWpctlOutput)
	if err != nil {
		t.Fatalf("unexpected error parsing wpctl output: %v", err)
	}
	if id != 99 {
		t.Fatalf("expected default sink ID 99, got %d", id)
	}
}

func TestParseNodeIDFromPwCli(t *testing.T) {
	samplePwCliOutput := `
	id 58, type PipeWire:Interface:Node/3
 		node.name = "alsa_output.pci-0000_01_00.1.hdmi-stereo"
	id 77, type PipeWire:Interface:Node/3
 		node.name = "multiout"
	id 99, type PipeWire:Interface:Node/3
 		node.name = "alsa_output.usb-CHIYINFOX_Co._Ltd._Equinox_Phobos_000000000000-00.analog-stereo"
`
	id, err := parseNodeIDFromPwCli(samplePwCliOutput, "multiout")
	if err != nil {
		t.Fatalf("unexpected error parsing pw-cli node ID: %v", err)
	}
	if id != 77 {
		t.Fatalf("expected node ID 77, got %d", id)
	}
}
