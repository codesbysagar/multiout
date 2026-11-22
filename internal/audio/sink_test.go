package audio

import (
	"testing"
)

func TestCheckFeedbackRisk(t *testing.T) {
	playDev := Device{ID: "playback:1", Name: "Ryzen HD Audio Controller Analog Stereo", Kind: Playback}

	// 1. Direct monitor of playback device: feedback risk
	capMonitorDirect := Device{ID: "capture:1", Name: "Monitor of Ryzen HD Audio Controller Analog Stereo", Kind: Capture}
	if !CheckFeedbackRisk(capMonitorDirect, playDev) {
		t.Fatalf("expected feedback risk for %q and %q", capMonitorDirect.Name, playDev.Name)
	}

	// 2. Monitor of a different device (e.g. HDMI): no feedback risk
	capMonitorOther := Device{ID: "capture:0", Name: "Monitor of HDA NVidia Digital Stereo (HDMI)", Kind: Capture}
	if CheckFeedbackRisk(capMonitorOther, playDev) {
		t.Fatalf("expected no feedback risk for %q and %q", capMonitorOther.Name, playDev.Name)
	}

	// 3. Monitor of virtual sink: no feedback risk
	capVirtualSink := Device{ID: "capture:5", Name: "Monitor of Multiout Virtual Sink", Kind: Capture}
	if CheckFeedbackRisk(capVirtualSink, playDev) {
		t.Fatalf("expected no feedback risk for virtual sink monitor")
	}

	// 4. Microphone: no feedback risk
	capMic := Device{ID: "capture:2", Name: "Ryzen HD Audio Controller Analog Stereo", Kind: Capture}
	if CheckFeedbackRisk(capMic, playDev) {
		t.Fatalf("expected no feedback risk for microphone capture")
	}
}

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

