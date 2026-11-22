package engine

import (
	"sync/atomic"
)

// DevicePipeline manages the end-to-end playback chain for a single output device:
// reading from the Ring, compensating clock drift, resampling, smoothing gain/mute,
// and metering peak volume.
//
// Process is real-time safe (zero heap allocations, no I/O, no logging).
type DevicePipeline struct {
	reader    *Reader
	resampler *LinearResampler
	drift     *DriftController
	gain      *GainController

	channels   int
	sampleRate int

	baseTargetFrames int
	delayFrames      int

	// Scratch buffer for reading from Ring into resampler (pre-allocated to avoid allocation in callback)
	scratchIn []float32

	underruns atomic.Uint64
}

// NewDevicePipeline creates a pipeline for an output device attached to ring.
// baseTargetMs is the baseline buffer fill target (e.g. 40ms).
func NewDevicePipeline(ring *Ring, channels, sampleRate, baseTargetMs int) *DevicePipeline {
	if channels <= 0 {
		channels = 2
	}
	if sampleRate <= 0 {
		sampleRate = 48000
	}
	if baseTargetMs <= 0 {
		baseTargetMs = 40
	}

	baseTargetFrames := (baseTargetMs * sampleRate) / 1000
	baseTargetSamples := baseTargetFrames * channels

	reader := ring.NewReader()
	// Pre-allocate scratch buffer to hold up to 200ms of input frames
	maxScratchFrames := sampleRate / 5
	scratchIn := make([]float32, maxScratchFrames*channels)

	return &DevicePipeline{
		reader:           reader,
		resampler:        NewLinearResampler(channels),
		drift:            NewDriftController(baseTargetSamples),
		gain:             NewGainController(1.0, (10*sampleRate)/1000), // 10ms ramp
		channels:         channels,
		sampleRate:       sampleRate,
		baseTargetFrames: baseTargetFrames,
		scratchIn:        scratchIn,
	}
}

// SetDelay sets the delay for this device in milliseconds (0 to 300 ms).
func (p *DevicePipeline) SetDelay(delayMs int) {
	if delayMs < 0 {
		delayMs = 0
	} else if delayMs > 300 {
		delayMs = 300
	}

	p.delayFrames = (delayMs * p.sampleRate) / 1000
	delaySamples := p.delayFrames * p.channels
	p.reader.SetMinFill(delaySamples)

	totalTargetSamples := (p.baseTargetFrames + p.delayFrames) * p.channels
	p.drift.SetTarget(totalTargetSamples)
}

// SetVolume sets target gain in range [0.0, 1.0].
func (p *DevicePipeline) SetVolume(vol float32) {
	p.gain.SetVolume(vol)
}

// Volume returns current target gain.
func (p *DevicePipeline) Volume() float32 {
	return p.gain.Volume()
}

// SetMute toggles mute.
func (p *DevicePipeline) SetMute(mute bool) {
	p.gain.SetMute(mute)
}

// IsMuted reports mute state.
func (p *DevicePipeline) IsMuted() bool {
	return p.gain.IsMuted()
}

// Peak returns the peak volume level recorded during the last callback.
func (p *DevicePipeline) Peak() float32 {
	return p.gain.Peak()
}

// Underruns returns the count of times audio samples were missing for a callback.
func (p *DevicePipeline) Underruns() uint64 {
	return p.underruns.Load()
}

// Overruns returns the count of times the writer lapped this reader.
func (p *DevicePipeline) Overruns() uint64 {
	return p.reader.Overruns()
}

// Fill returns the current unread samples buffered for this device.
func (p *DevicePipeline) Fill() int {
	return p.reader.Fill()
}

// Process is the audio callback function that produces out samples.
// Real-time safe: zero memory allocations.
func (p *DevicePipeline) Process(out []float32) {
	outSamples := len(out)
	if outSamples == 0 {
		return
	}
	outFrames := outSamples / p.channels

	// 1. Update drift controller based on current fill
	fill := p.reader.Fill()
	ratio := p.drift.Update(fill)

	// 2. Determine how many input frames are needed for the resampler
	neededFrames := p.resampler.FramesNeeded(outFrames, ratio)
	neededSamples := neededFrames * p.channels

	// Ensure scratch buffer is large enough (should always be preallocated)
	scratch := p.scratchIn
	if len(scratch) < neededSamples {
		scratch = make([]float32, neededSamples)
		p.scratchIn = scratch
	}

	// 3. Peek needed samples from Ring without advancing reader cursor
	readSamples := p.reader.Peek(scratch[:neededSamples])

	// If underrun, pad remainder of input with silence and increment counter
	if readSamples < neededSamples {
		p.underruns.Add(1)
		for i := readSamples; i < neededSamples; i++ {
			scratch[i] = 0
		}
	}

	// 4. Resample into out buffer and determine actual consumed frames
	consumedFrames := p.resampler.Resample(scratch[:neededSamples], out, outFrames, ratio)

	// 5. Advance reader cursor by the number of samples actually consumed by resampler
	consumedSamples := min(consumedFrames*p.channels, readSamples)
	p.reader.Advance(consumedSamples)

	// 6. Apply volume / mute ramp and record peak level
	p.gain.Apply(out, p.channels)
}
