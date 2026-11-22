package engine

import (
	"math"
	"sync/atomic"
)

// GainController manages volume and mute with linear ramping to prevent clicks/zipper noise.
// All parameter changes (SetVolume, SetMute) are safe to call from any goroutine.
// Apply is real-time safe and performs zero allocations.
type GainController struct {
	targetGain  atomic.Uint32 // float32 bits
	isMuted     atomic.Bool
	currentGain float32       // only mutated in audio callback (Apply)
	peak        atomic.Uint32 // float32 bits, peak level for metering
	rampStep    float32       // change in gain per frame
}

// NewGainController creates a controller with an initial volume (0.0 to 4.0, up to +12 dB)
// and a ramp duration (e.g., 10ms = 480 frames at 48kHz).
func NewGainController(initialVolume float32, rampFrames int) *GainController {
	if initialVolume < 0 {
		initialVolume = 0
	} else if initialVolume > 4.0 {
		initialVolume = 4.0
	}
	if rampFrames <= 0 {
		rampFrames = 480 // 10ms at 48kHz
	}

	gc := &GainController{
		currentGain: initialVolume,
		rampStep:    1.0 / float32(rampFrames),
	}
	gc.SetVolume(initialVolume)
	return gc
}

// SetVolume sets the target volume between 0.0 and 4.0 (up to +12 dB boost).
func (gc *GainController) SetVolume(v float32) {
	if v < 0 {
		v = 0
	} else if v > 4.0 {
		v = 4.0
	}
	gc.targetGain.Store(math.Float32bits(v))
}

// DBToGain converts a decibel level (-60.0 to +12.0 dB) to a linear amplitude gain.
// Values <= -59.5 dB are treated as silence (0.0).
func DBToGain(db float64) float32 {
	if db <= -59.5 {
		return 0.0
	}
	return float32(math.Pow(10, db/20.0))
}

// GainToDB converts a linear amplitude gain to a decibel level (-60.0 to +12.0 dB).
func GainToDB(gain float32) float64 {
	if gain <= 0.001 {
		return -60.0
	}
	db := 20.0 * math.Log10(float64(gain))
	if db > 12.0 {
		return 12.0
	}
	if db < -60.0 {
		return -60.0
	}
	return db
}

// Volume returns the current target volume.
func (gc *GainController) Volume() float32 {
	return math.Float32frombits(gc.targetGain.Load())
}

// SetMute toggles mute. When muted, gain ramps down to 0.
func (gc *GainController) SetMute(muted bool) {
	gc.isMuted.Store(muted)
}

// IsMuted returns whether audio is currently muted.
func (gc *GainController) IsMuted() bool {
	return gc.isMuted.Load()
}

// Peak returns the peak absolute sample value from the most recent callback (0.0 to 1.0).
func (gc *GainController) Peak() float32 {
	return math.Float32frombits(gc.peak.Load())
}

// Apply scales samples in-place by the smoothed gain and records the peak level.
// channels specifies the interleaving (e.g. 2 for stereo).
func (gc *GainController) Apply(samples []float32, channels int) {
	if len(samples) == 0 || channels <= 0 {
		return
	}

	target := math.Float32frombits(gc.targetGain.Load())
	if gc.isMuted.Load() {
		target = 0
	}

	cur := gc.currentGain
	step := gc.rampStep
	var maxAbs float32

	for i := 0; i < len(samples); i += channels {
		// Ramp towards target once per frame
		if cur < target {
			cur += step
			if cur > target {
				cur = target
			}
		} else if cur > target {
			cur -= step
			if cur < target {
				cur = target
			}
		}

		for ch := 0; ch < channels; ch++ {
			idx := i + ch
			if idx >= len(samples) {
				break
			}
			v := samples[idx] * cur
			// Clamp to [-1.0, 1.0]
			if v > 1.0 {
				v = 1.0
			} else if v < -1.0 {
				v = -1.0
			}
			samples[idx] = v

			abs := float32(math.Abs(float64(v)))
			if abs > maxAbs {
				maxAbs = abs
			}
		}
	}

	gc.currentGain = cur
	gc.peak.Store(math.Float32bits(maxAbs))
}
