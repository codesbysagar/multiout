package engine

import (
	"math"
	"testing"
)

func TestDriftControllerClamping(t *testing.T) {
	dc := NewDriftController(1000)

	// Massive buffer overrun condition: fill is 100,000 samples
	for i := 0; i < 100; i++ {
		ratio := dc.Update(100000)
		if ratio > 1.005001 {
			t.Fatalf("ratio %f exceeded max clamp 1.005", ratio)
		}
	}

	// Massive buffer starvation condition: fill is 0
	for i := 0; i < 100; i++ {
		ratio := dc.Update(0)
		if ratio < 0.994999 {
			t.Fatalf("ratio %f went below min clamp 0.995", ratio)
		}
	}
}

func TestDriftControllerConvergence(t *testing.T) {
	// Target is 1000 samples.
	target := 1000
	dc := NewDriftController(target)

	// Simulate clock skew:
	// At ratio 1.0, capture writes 480 samples, but playback consumes 480.05 samples (100 ppm skew).
	// Without drift control, the buffer would drain.
	// With drift control, the ratio should adapt and stabilize fill near target.

	simulatedFill := float64(target)
	clockSkew := 0.0001 // 100 ppm playback faster than capture

	for step := 0; step < 5000; step++ {
		ratio := dc.Update(int(simulatedFill))
		// Net change in fill per period = captureIn - playbackConsumed * ratio
		// Let nominal period = 480 frames
		inFrames := 480.0
		consumedFrames := 480.0 * (1.0 + clockSkew) * ratio
		simulatedFill += (inFrames - consumedFrames)
	}

	// After settling, fill should be close to target (+- 50 samples = ~1ms at 48kHz)
	diff := math.Abs(simulatedFill - float64(target))
	if diff > 100.0 {
		t.Fatalf("drift controller failed to settle near target: fill=%f, target=%d, diff=%f", simulatedFill, target, diff)
	}
}

