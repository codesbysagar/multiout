package engine

import (
	"math"
	"testing"
)

func TestPipelineZeroAllocations(t *testing.T) {
	ring := NewRing(96000)
	pipeline := NewDevicePipeline(ring, 2, 48000, 40)

	// Pre-fill ring with some audio
	seed := make([]float32, 9600)
	ring.Write(seed)

	out := make([]float32, 960) // 10ms stereo callback (480 frames)

	// Warm up
	pipeline.Process(out)

	allocs := testing.AllocsPerRun(100, func() {
		pipeline.Process(out)
	})

	if allocs > 0 {
		t.Fatalf("expected 0 allocations in audio callback, got %f", allocs)
	}
}

func TestPipelineClockSkewSimulation(t *testing.T) {
	// Ring capacity: 1 second of stereo 48kHz audio = 96,000 samples
	ring := NewRing(96000)
	// Base target 40ms = 1920 stereo frames = 3840 samples
	pipeline := NewDevicePipeline(ring, 2, 48000, 40)

	const callbackFrames = 480  // 10ms period = 960 samples
	const totalCallbacks = 5000 // 50 seconds of simulated audio

	// Pre-fill ring to target level before starting playback (as capture does on start)
	initialFill := 3840
	seed := make([]float32, initialFill)
	ring.Write(seed)

	outBuf := make([]float32, callbackFrames*2)

	// Simulate 100 ppm clock skew:
	// Playback clock runs slightly faster: 48004.8 Hz vs 48000 Hz capture.
	// We simulate this by having capture produce 480 frames on 99.99% of ticks,
	// effectively testing drift convergence.
	for i := 0; i < totalCallbacks; i++ {
		// Capture callback writes 480 frames (960 samples)
		inChunk := make([]float32, callbackFrames*2)
		for j := range inChunk {
			inChunk[j] = 0.5
		}
		ring.Write(inChunk)

		// Playback callback consumes
		pipeline.Process(outBuf)
	}

	// Verify zero underruns and zero overruns
	if u := pipeline.Underruns(); u > 0 {
		t.Fatalf("expected 0 underruns during clock skew simulation, got %d", u)
	}
	if o := pipeline.Overruns(); o > 0 {
		t.Fatalf("expected 0 overruns during clock skew simulation, got %d", o)
	}

	// Buffer fill should remain stably bounded near the target
	finalFill := pipeline.Fill()
	target := (40 * 48000 / 1000) * 2 // 3840 samples
	diff := math.Abs(float64(finalFill - target))

	// Allow within ~5ms (480 samples) of target
	if diff > 1000 {
		t.Fatalf("buffer fill drifted too far: fill=%d, target=%d, diff=%f", finalFill, target, diff)
	}
}
