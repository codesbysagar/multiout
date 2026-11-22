package engine

import (
	"math"
)

// LinearResampler resamples interleaved audio frames using linear interpolation.
// It maintains fractional phase across callbacks to ensure seamless continuity
// at buffer boundaries without audio clicks or zipper artifacts.
type LinearResampler struct {
	channels int
	phase    float64   // fractional position in [0.0, 1.0)
	history  []float32 // holds the last frame from previous callback [channels]
	hasHist  bool
}

// NewLinearResampler creates a resampler for the given channel count (e.g. 2 for stereo).
func NewLinearResampler(channels int) *LinearResampler {
	if channels <= 0 {
		channels = 2
	}
	return &LinearResampler{
		channels: channels,
		history:  make([]float32, channels),
	}
}

// Reset clears the resampler's internal phase and history.
func (r *LinearResampler) Reset() {
	r.phase = 0.0
	r.hasHist = false
	for i := range r.history {
		r.history[i] = 0
	}
}

// FramesNeeded calculates how many input frames are required to generate
// outFrames at the given resample ratio.
func (r *LinearResampler) FramesNeeded(outFrames int, ratio float64) int {
	if outFrames <= 0 {
		return 0
	}
	totalPos := r.phase + float64(outFrames)*ratio
	needed := int(math.Ceil(totalPos))
	if needed < 0 {
		return 0
	}
	return needed
}

// Resample produces outFrames into out from in, according to ratio.
// Returns the number of input frames consumed.
// out must have length >= outFrames * channels.
// in must have length >= FramesNeeded(outFrames, ratio) * channels.
func (r *LinearResampler) Resample(in []float32, out []float32, outFrames int, ratio float64) int {
	if outFrames <= 0 || len(out) < outFrames*r.channels {
		return 0
	}
	ch := r.channels

	// If no history yet, seed history with the first frame of input
	if !r.hasHist && len(in) >= ch {
		copy(r.history, in[:ch])
		r.hasHist = true
	}

	phase := r.phase
	inFrames := len(in) / ch

	for i := 0; i < outFrames; i++ {
		pos := phase + float64(i)*ratio
		idx := int(math.Floor(pos))
		frac := float32(pos - float64(idx))

		outBase := i * ch
		for c := 0; c < ch; c++ {
			var left, right float32

			// Left sample
			if idx < 0 {
				left = r.history[c]
			} else if idx < inFrames {
				left = in[idx*ch+c]
			} else if inFrames > 0 {
				left = in[(inFrames-1)*ch+c]
			}

			// Right sample
			nextIdx := idx + 1
			if nextIdx < 0 {
				right = r.history[c]
			} else if nextIdx < inFrames {
				right = in[nextIdx*ch+c]
			} else if inFrames > 0 {
				right = in[(inFrames-1)*ch+c]
			}

			out[outBase+c] = left*(1.0-frac) + right*frac
		}
	}

	// Update phase and determine frames consumed
	totalPos := phase + float64(outFrames)*ratio
	consumed := int(math.Floor(totalPos))
	if consumed > inFrames {
		consumed = inFrames
	}

	r.phase = totalPos - float64(consumed)

	// Save history for next callback
	if consumed > 0 && inFrames > 0 {
		lastIdx := consumed - 1
		if lastIdx >= inFrames {
			lastIdx = inFrames - 1
		}
		copy(r.history, in[lastIdx*ch:(lastIdx+1)*ch])
	}

	return consumed
}
