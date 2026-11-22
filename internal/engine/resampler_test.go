package engine

import (
	"math"
	"testing"
)

func TestLinearResamplerPassthrough(t *testing.T) {
	resampler := NewLinearResampler(2)
	inFrames := 100
	in := make([]float32, inFrames*2)
	for i := range in {
		in[i] = float32(i)
	}

	out := make([]float32, inFrames*2)
	consumed := resampler.Resample(in, out, inFrames, 1.0)

	if consumed != inFrames {
		t.Fatalf("expected consumed %d, got %d", inFrames, consumed)
	}

	for i := range out {
		if math.Abs(float64(out[i]-in[i])) > 1e-4 {
			t.Fatalf("at sample %d, got %f want %f", i, out[i], in[i])
		}
	}
}

func TestLinearResamplerContinuityAcrossCallbacks(t *testing.T) {
	// Generate a continuous sine wave at 480 Hz (period = 100 frames at 48 kHz).
	// Feed it in chunks with ratio 1.002 (clock drift simulation).
	// Verify that the resampled output does not have jumps > max delta at chunk boundaries.
	resampler := NewLinearResampler(2)
	ratio := 1.002
	chunkOut := 480 // 10ms callback

	var allOut []float32
	timeStep := 2 * math.Pi * 480.0 / 48000.0
	currentInputFrame := 0

	for cb := 0; cb < 10; cb++ {
		needed := resampler.FramesNeeded(chunkOut, ratio) + 10
		in := make([]float32, needed*2)
		for f := 0; f < needed; f++ {
			v := float32(math.Sin(float64(currentInputFrame+f) * timeStep))
			in[f*2] = v
			in[f*2+1] = v
		}

		out := make([]float32, chunkOut*2)
		consumed := resampler.Resample(in, out, chunkOut, ratio)
		currentInputFrame += consumed
		allOut = append(allOut, out...)
	}

	// Verify waveform continuity: maximum step between consecutive frames should be smooth
	maxAllowedStep := float32(2*math.Pi*480.0/48000.0*1.5) + 0.01 // approx derivative * dt
	for i := 2; i < len(allOut); i += 2 {
		step := float32(math.Abs(float64(allOut[i] - allOut[i-2])))
		if step > maxAllowedStep {
			t.Fatalf("discontinuity detected at frame %d: step %f > allowed %f", i/2, step, maxAllowedStep)
		}
	}
}

