package engine

import (
	"sync/atomic"
)

// DriftController monitors reader buffer fill and adjusts the resample ratio
// using a low-pass filter and a PI controller to maintain target latency and eliminate clock drift.
type DriftController struct {
	targetFill    atomic.Int64 // in samples
	filteredFill  float64      // exponential moving average
	integral      float64      // accumulated error for integral term
	alpha         float64      // low-pass filter smoothing factor [0.0, 1.0]
	kp            float64      // proportional gain
	ki            float64      // integral gain
	maxRatioDelta float64      // maximum deviation from 1.0 (default 0.005 = 5000 ppm)
	initialized   bool
}

// NewDriftController creates a drift controller with the given target fill (in samples).
func NewDriftController(targetFill int) *DriftController {
	dc := &DriftController{
		alpha:         0.05,  // smooth over ~20 callbacks (~200ms at 10ms periods)
		kp:            1e-5,  // responsive proportional control
		ki:            5e-7,  // integral control to eliminate steady-state error
		maxRatioDelta: 0.005, // +- 5000 ppm
	}
	dc.SetTarget(targetFill)
	return dc
}

// SetTarget updates the target buffer fill in samples (baseTarget + delay).
func (dc *DriftController) SetTarget(samples int) {
	if samples < 0 {
		samples = 0
	}
	dc.targetFill.Store(int64(samples))
}

// Target returns the current target buffer fill in samples.
func (dc *DriftController) Target() int {
	return int(dc.targetFill.Load())
}

// FilteredFill returns the low-pass filtered fill estimate in samples.
func (dc *DriftController) FilteredFill() float64 {
	return dc.filteredFill
}

// Reset clears the filter history and integrator.
func (dc *DriftController) Reset() {
	dc.initialized = false
	dc.filteredFill = 0
	dc.integral = 0
}

// Update measures currentFill, updates the low-pass filter and PI loop,
// and returns the target resample ratio (centered around 1.0).
func (dc *DriftController) Update(currentFill int) float64 {
	fill := float64(currentFill)

	if !dc.initialized {
		dc.filteredFill = fill
		dc.initialized = true
	} else {
		// Low-pass filter: EMA
		dc.filteredFill = dc.alpha*fill + (1.0-dc.alpha)*dc.filteredFill
	}

	target := float64(dc.targetFill.Load())
	err := dc.filteredFill - target

	// Accumulate integral with anti-windup clamping
	dc.integral += err
	maxIntegral := dc.maxRatioDelta / dc.ki
	if dc.integral > maxIntegral {
		dc.integral = maxIntegral
	} else if dc.integral < -maxIntegral {
		dc.integral = -maxIntegral
	}

	// Ratio calculation:
	// If filteredFill > target (buffer growing): we want to consume more input -> ratio > 1.0.
	// If filteredFill < target (buffer shrinking): consume less input -> ratio < 1.0.
	delta := dc.kp*err + dc.ki*dc.integral

	// Clamp to +- maxRatioDelta
	if delta > dc.maxRatioDelta {
		delta = dc.maxRatioDelta
	} else if delta < -dc.maxRatioDelta {
		delta = -dc.maxRatioDelta
	}

	return 1.0 + delta
}
