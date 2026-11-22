package engine

import (
	"math"
	"testing"
)

func TestGainControllerVolumeRamping(t *testing.T) {
	// 10 frames to ramp from 0 to 1
	gc := NewGainController(0.0, 10)
	gc.SetVolume(1.0)

	samples := make([]float32, 20) // 10 stereo frames (20 samples) of value 1.0
	for i := range samples {
		samples[i] = 1.0
	}

	gc.Apply(samples, 2)

	// After 10 frames, current gain should reach 1.0
	if math.Abs(float64(gc.currentGain-1.0)) > 1e-4 {
		t.Fatalf("expected current gain 1.0, got %f", gc.currentGain)
	}

	// First frame should have been scaled by ~0.1
	if math.Abs(float64(samples[0]-0.1)) > 1e-4 {
		t.Fatalf("expected samples[0] ~ 0.1, got %f", samples[0])
	}
	// Last frame should have been scaled by 1.0
	if math.Abs(float64(samples[18]-1.0)) > 1e-4 {
		t.Fatalf("expected samples[18] ~ 1.0, got %f", samples[18])
	}
}

func TestGainControllerMute(t *testing.T) {
	gc := NewGainController(1.0, 5)
	gc.SetMute(true)

	samples := make([]float32, 20) // 10 stereo frames
	for i := range samples {
		samples[i] = 1.0
	}

	gc.Apply(samples, 2)

	// After 5 frames, should be muted (gain 0)
	if gc.currentGain != 0.0 {
		t.Fatalf("expected gain 0 after mute, got %f", gc.currentGain)
	}
	// Samples at the end must be 0
	for i := 10; i < len(samples); i++ {
		if samples[i] != 0.0 {
			t.Fatalf("expected sample %d to be 0, got %f", i, samples[i])
		}
	}
}

func TestGainControllerClampingAndPeak(t *testing.T) {
	gc := NewGainController(1.0, 1)

	samples := []float32{1.5, -2.0, 0.5, -0.2}
	gc.Apply(samples, 2)

	if samples[0] != 1.0 || samples[1] != -1.0 {
		t.Fatalf("expected clamped values [1.0, -1.0], got [%f, %f]", samples[0], samples[1])
	}

	if gc.Peak() != 1.0 {
		t.Fatalf("expected peak 1.0, got %f", gc.Peak())
	}
}

func TestDBConversionsAndBoost(t *testing.T) {
	// 0 dB is 1.0 unity gain
	if math.Abs(float64(DBToGain(0.0)-1.0)) > 1e-3 {
		t.Errorf("expected 0 dB to be 1.0, got %f", DBToGain(0.0))
	}
	if math.Abs(GainToDB(1.0)-0.0) > 1e-3 {
		t.Errorf("expected 1.0 gain to be 0 dB, got %f", GainToDB(1.0))
	}

	// +6 dB is approx 2.0x gain
	gain6 := DBToGain(6.0)
	if math.Abs(float64(gain6-1.995)) > 0.05 {
		t.Errorf("expected +6 dB to be ~2.0, got %f", gain6)
	}

	// +12 dB is approx 4.0x gain
	gain12 := DBToGain(12.0)
	if math.Abs(float64(gain12-3.981)) > 0.05 {
		t.Errorf("expected +12 dB to be ~3.98, got %f", gain12)
	}

	// -60 dB is silence
	if DBToGain(-60.0) != 0.0 {
		t.Errorf("expected -60 dB to be 0.0, got %f", DBToGain(-60.0))
	}

	// Verify gain controller supports +6 dB boost
	gc := NewGainController(2.0, 1)
	samples := []float32{0.25, -0.25}
	gc.Apply(samples, 2)
	if math.Abs(float64(samples[0]-0.5)) > 1e-3 {
		t.Errorf("expected sample 0.25 boosted 2x to be 0.5, got %f", samples[0])
	}
}
