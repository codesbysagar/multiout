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

	// 5. Windows direct loopback: feedback risk
	capLoopbackDirect := Device{ID: "loopback:1", Name: "Ryzen HD Audio Controller Analog Stereo (Loopback)", Kind: Capture}
	if !CheckFeedbackRisk(capLoopbackDirect, playDev) {
		t.Fatalf("expected feedback risk for Windows loopback on same device")
	}

	// 6. Windows other loopback: no feedback risk
	capLoopbackOther := Device{ID: "loopback:0", Name: "HDA NVidia Digital Stereo (HDMI) (Loopback)", Kind: Capture}
	if CheckFeedbackRisk(capLoopbackOther, playDev) {
		t.Fatalf("expected no feedback risk for Windows loopback on different device")
	}

	// 7. Windows default system audio loopback: risk if playback device is default
	playDevDefault := Device{ID: "playback:1", Name: "Ryzen HD Audio Controller Analog Stereo", Kind: Playback, IsDefault: true}
	capSystemDefault := Device{ID: "loopback:default", Name: "System Audio (Default Output)", Kind: Capture, IsDefault: true}
	if !CheckFeedbackRisk(capSystemDefault, playDevDefault) {
		t.Fatalf("expected feedback risk for System Audio default loopback and default playback device")
	}
	if CheckFeedbackRisk(capSystemDefault, playDev) {
		t.Fatalf("expected no feedback risk for System Audio default loopback and non-default playback device")
	}
}
