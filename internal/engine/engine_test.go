package engine

import (
	"slices"
	"testing"

	"multiout/internal/audio"
)

func TestEngineFanOutAndControls(t *testing.T) {
	fakeBe := audio.NewFakeBackend()
	defer fakeBe.Close()

	cfg := audio.Config{SampleRate: 48000, Channels: 2}
	eng := NewEngine(fakeBe, cfg)
	defer eng.Close()

	// 1. Start Capture
	capDev := audio.Device{ID: "fake:capture:0", Name: "Monitor of Virtual Sink", Kind: audio.Capture}
	if err := eng.StartCapture(capDev); err != nil {
		t.Fatalf("start capture failed: %v", err)
	}
	if eng.CaptureDeviceID() != "fake:capture:0" {
		t.Fatalf("unexpected capture device ID %q", eng.CaptureDeviceID())
	}

	// 2. Add two output playback devices
	devA := audio.Device{ID: "fake:playback:0", Name: "Fake Headset", Kind: audio.Playback}
	devB := audio.Device{ID: "fake:playback:1", Name: "Fake Speakers", Kind: audio.Playback}

	pA, err := eng.AddOutput(devA)
	if err != nil {
		t.Fatalf("add output A failed: %v", err)
	}
	_, err = eng.AddOutput(devB)
	if err != nil {
		t.Fatalf("add output B failed: %v", err)
	}

	// 3. Set different volumes and mute states
	if err := eng.SetOutputVolume(devA.ID, 0.5); err != nil {
		t.Fatalf("set volume A failed: %v", err)
	}
	if err := eng.SetOutputDelay(devB.ID, 50); err != nil {
		t.Fatalf("set delay B failed: %v", err)
	}
	if err := eng.SetOutputMute(devB.ID, true); err != nil {
		t.Fatalf("set mute B failed: %v", err)
	}

	// 4. Verify status snapshot
	states := eng.Outputs()
	if len(states) != 2 {
		t.Fatalf("expected 2 output states, got %d", len(states))
	}

	var stateA, stateB OutputDeviceState
	for _, s := range states {
		if s.ID == devA.ID {
			stateA = s
		} else if s.ID == devB.ID {
			stateB = s
		}
	}

	if stateA.Volume != 0.5 {
		t.Fatalf("state A volume %f want 0.5", stateA.Volume)
	}
	if !stateB.Muted {
		t.Fatalf("state B should be muted")
	}
	if stateB.DelayMs != 50 {
		t.Fatalf("state B delay %d want 50", stateB.DelayMs)
	}

	// 5. Test audio fan-out through FakeBackend streams
	capStream := fakeBe.GetCaptureStream("fake:capture:0")
	playStreamA := fakeBe.GetPlaybackStream(devA.ID)
	playStreamB := fakeBe.GetPlaybackStream(devB.ID)

	// Feed test frames into capture
	chunk := make([]float32, 960) // 10ms
	for i := range chunk {
		chunk[i] = 1.0
	}
	capStream.TickCapture(chunk)

	// Pull output from both playback streams
	outA := make([]float32, 960)
	outB := make([]float32, 960)

	playStreamA.TickPlayback(480, outA)
	playStreamB.TickPlayback(480, outB)

	// B was muted, so its samples should all be 0
	for i, v := range outB {
		if v != 0 {
			t.Fatalf("muted device B played non-zero sample at %d: %f", i, v)
		}
	}

	// A was unmuted with volume 0.5, so its output should be non-zero
	hasAudio := false
	for _, v := range outA {
		if v > 0 {
			hasAudio = true
			break
		}
	}
	if !hasAudio && pA.Fill() > 0 {
		t.Fatalf("device A produced all zeroes")
	}

	// 6. Test removing output
	if err := eng.RemoveOutput(devA.ID); err != nil {
		t.Fatalf("remove output failed: %v", err)
	}
	if len(eng.Outputs()) != 1 {
		t.Fatalf("expected 1 output after removal, got %d", len(eng.Outputs()))
	}
}

func TestEngineOutputToggling(t *testing.T) {
	fakeBe := audio.NewFakeBackend()
	defer fakeBe.Close()

	eng := NewEngine(fakeBe, audio.Config{SampleRate: 48000, Channels: 2})
	defer eng.Close()

	dev := audio.Device{ID: "fake:playback:0", Name: "Fake Headset", Kind: audio.Playback}
	_, err := eng.AddOutput(dev)
	if err != nil {
		t.Fatalf("add output failed: %v", err)
	}

	if err := eng.SetOutputEnabled(dev.ID, false); err != nil {
		t.Fatalf("disable output failed: %v", err)
	}

	playStream := fakeBe.GetPlaybackStream(dev.ID)
	out := make([]float32, 20)
	for i := range out {
		out[i] = 0.99
	}
	playStream.TickPlayback(10, out)

	// Disabled stream must output silence (all zeroes)
	zeroes := make([]float32, 20)
	if !slices.Equal(out, zeroes) {
		t.Fatalf("disabled output should produce silence, got %v", out)
	}
}

func TestEngineFeedbackProtection(t *testing.T) {
	fakeBe := audio.NewFakeBackend()
	defer fakeBe.Close()

	eng := NewEngine(fakeBe, audio.Config{SampleRate: 48000, Channels: 2})
	defer eng.Close()

	headset := audio.Device{ID: "fake:playback:0", Name: "Fake Headset", Kind: audio.Playback}
	headsetMonitor := audio.Device{ID: "fake:capture:0", Name: "Monitor of Fake Headset", Kind: audio.Capture}

	// 1. Add output and enable it
	_, err := eng.AddOutput(headset)
	if err != nil {
		t.Fatalf("add output failed: %v", err)
	}

	// 2. Attempting to start capture from headset monitor while headset is output must fail
	err = eng.StartCapture(headsetMonitor)
	if err == nil {
		t.Fatalf("expected feedback protection error when capturing from monitor of enabled output")
	}

	// 3. Disable output, then starting capture should succeed
	if err := eng.SetOutputEnabled(headset.ID, false); err != nil {
		t.Fatalf("disable output failed: %v", err)
	}
	if err := eng.StartCapture(headsetMonitor); err != nil {
		t.Fatalf("start capture after disable failed: %v", err)
	}

	// 4. While capture is active, attempting to re-enable headset output must fail
	err = eng.SetOutputEnabled(headset.ID, true)
	if err == nil {
		t.Fatalf("expected feedback protection error when enabling output that matches active capture monitor")
	}
}
