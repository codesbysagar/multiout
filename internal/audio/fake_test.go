package audio

import (
	"slices"
	"testing"
)

func TestFakeBackendDevices(t *testing.T) {
	be := NewFakeBackend()
	defer be.Close()

	playDevs, err := be.Devices(Playback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(playDevs) != 2 {
		t.Fatalf("expected 2 playback devices, got %d", len(playDevs))
	}

	capDevs, err := be.Devices(Capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capDevs) != 2 {
		t.Fatalf("expected 2 capture devices, got %d", len(capDevs))
	}
}

func TestFakeBackendStreamIO(t *testing.T) {
	be := NewFakeBackend()
	defer be.Close()

	cfg := Config{SampleRate: 48000, Channels: 2}

	var captured []float32
	capStream, err := be.OpenCapture("fake:capture:0", cfg, func(in []float32) {
		captured = append(captured, in...)
	})
	if err != nil {
		t.Fatalf("open capture failed: %v", err)
	}
	if err := capStream.Start(); err != nil {
		t.Fatalf("start capture failed: %v", err)
	}

	fakeCap := be.GetCaptureStream("fake:capture:0")
	inputData := []float32{0.1, 0.2, 0.3, 0.4}
	fakeCap.TickCapture(inputData)

	if !slices.Equal(captured, inputData) {
		t.Fatalf("captured %v, want %v", captured, inputData)
	}

	var generated []float32
	playStream, err := be.OpenPlayback("fake:playback:0", cfg, func(out []float32) {
		for i := range out {
			out[i] = float32(i + 1)
		}
	})
	if err != nil {
		t.Fatalf("open playback failed: %v", err)
	}
	if err := playStream.Start(); err != nil {
		t.Fatalf("start playback failed: %v", err)
	}

	fakePlay := be.GetPlaybackStream("fake:playback:0")
	outBuf := make([]float32, 4)
	fakePlay.TickPlayback(2, outBuf)
	generated = outBuf

	expected := []float32{1, 2, 3, 4}
	if !slices.Equal(generated, expected) {
		t.Fatalf("generated %v, want %v", generated, expected)
	}
}

