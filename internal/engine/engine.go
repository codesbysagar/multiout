package engine

import (
	"fmt"
	"sync"
	"sync/atomic"

	"multiout/internal/audio"
)

// OutputDeviceState contains a thread-safe snapshot of an output device's runtime metrics.
type OutputDeviceState struct {
	ID        string
	Name      string
	Enabled   bool
	Volume    float32
	Muted     bool
	DelayMs   int
	Peak      float32
	Underruns uint64
	Overruns  uint64
	Fill      int
}

type outputTrack struct {
	dev      audio.Device
	pipeline *DevicePipeline
	stream   audio.Stream
	enabled  atomic.Bool
	delayMs  atomic.Int64
}

// Engine coordinates system audio capture and fan-out to multiple playback outputs.
// It is thread-safe and isolates audio callback logic from UI and control operations.
type Engine struct {
	backend audio.Backend
	cfg     audio.Config
	ring    *Ring

	mu            sync.RWMutex
	captureStream audio.Stream
	captureDevID  string
	outputs       map[string]*outputTrack
	closed        bool
}

// NewEngine creates an Engine using the provided audio backend and configuration.
// Default configuration: 48000 Hz, stereo (channels = 2), 1-second ring buffer.
func NewEngine(backend audio.Backend, cfg audio.Config) *Engine {
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = 48000
	}
	if cfg.Channels <= 0 {
		cfg.Channels = 2
	}

	// 1 second ring capacity (48000 frames * 2 channels = 96000 samples)
	ringCapacity := cfg.SampleRate * cfg.Channels
	return &Engine{
		backend: backend,
		cfg:     cfg,
		ring:    NewRing(ringCapacity),
		outputs: make(map[string]*outputTrack),
	}
}

// StartCapture initiates capture from the specified device ID (e.g. monitor sink).
func (e *Engine) StartCapture(deviceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return fmt.Errorf("engine is closed")
	}

	if e.captureStream != nil {
		if e.captureDevID == deviceID {
			return nil
		}
		_ = e.captureStream.Stop()
		_ = e.captureStream.Close()
		e.captureStream = nil
	}

	stream, err := e.backend.OpenCapture(deviceID, e.cfg, func(in []float32) {
		e.ring.Write(in)
	})
	if err != nil {
		return fmt.Errorf("open capture %s: %w", deviceID, err)
	}

	if err := stream.Start(); err != nil {
		_ = stream.Close()
		return fmt.Errorf("start capture %s: %w", deviceID, err)
	}

	e.captureStream = stream
	e.captureDevID = deviceID
	return nil
}

// StopCapture stops the active capture stream.
func (e *Engine) StopCapture() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.captureStream != nil {
		_ = e.captureStream.Stop()
		err := e.captureStream.Close()
		e.captureStream = nil
		e.captureDevID = ""
		return err
	}
	return nil
}

// CaptureDeviceID returns the ID of the active capture device.
func (e *Engine) CaptureDeviceID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.captureDevID
}

// AddOutput registers and opens an output playback device.
func (e *Engine) AddOutput(dev audio.Device) (*DevicePipeline, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, fmt.Errorf("engine is closed")
	}

	if track, exists := e.outputs[dev.ID]; exists {
		return track.pipeline, nil
	}

	pipeline := NewDevicePipeline(e.ring, e.cfg.Channels, e.cfg.SampleRate, 40) // 40ms base target

	track := &outputTrack{
		dev:      dev,
		pipeline: pipeline,
	}

	stream, err := e.backend.OpenPlayback(dev.ID, e.cfg, func(out []float32) {
		if !track.enabled.Load() {
			for i := range out {
				out[i] = 0
			}
			return
		}
		pipeline.Process(out)
	})
	if err != nil {
		return nil, fmt.Errorf("open playback %s: %w", dev.ID, err)
	}

	if err := stream.Start(); err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("start playback %s: %w", dev.ID, err)
	}

	track.stream = stream
	track.enabled.Store(true)
	e.outputs[dev.ID] = track

	return pipeline, nil
}

// RemoveOutput closes and unregisters an output device.
func (e *Engine) RemoveOutput(deviceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	track, exists := e.outputs[deviceID]
	if !exists {
		return nil
	}

	delete(e.outputs, deviceID)
	track.enabled.Store(false)
	_ = track.stream.Stop()
	return track.stream.Close()
}

// SetOutputEnabled toggles playback for a specific output device.
func (e *Engine) SetOutputEnabled(deviceID string, enabled bool) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	track, exists := e.outputs[deviceID]
	if !exists {
		return fmt.Errorf("output device %s not found", deviceID)
	}
	track.enabled.Store(enabled)
	return nil
}

// SetOutputVolume sets the target volume (0.0 to 1.0) for a specific output device.
func (e *Engine) SetOutputVolume(deviceID string, volume float32) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	track, exists := e.outputs[deviceID]
	if !exists {
		return fmt.Errorf("output device %s not found", deviceID)
	}
	track.pipeline.SetVolume(volume)
	return nil
}

// SetOutputMute toggles mute for a specific output device.
func (e *Engine) SetOutputMute(deviceID string, muted bool) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	track, exists := e.outputs[deviceID]
	if !exists {
		return fmt.Errorf("output device %s not found", deviceID)
	}
	track.pipeline.SetMute(muted)
	return nil
}

// SetOutputDelay sets delay in milliseconds (0 to 300 ms) for a specific output device.
func (e *Engine) SetOutputDelay(deviceID string, delayMs int) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	track, exists := e.outputs[deviceID]
	if !exists {
		return fmt.Errorf("output device %s not found", deviceID)
	}
	track.delayMs.Store(int64(delayMs))
	track.pipeline.SetDelay(delayMs)
	return nil
}

// Outputs returns a snapshot of all registered output devices and their runtime status.
// Safe to poll periodically from the UI goroutine.
func (e *Engine) Outputs() []OutputDeviceState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	res := make([]OutputDeviceState, 0, len(e.outputs))
	for id, track := range e.outputs {
		p := track.pipeline
		res = append(res, OutputDeviceState{
			ID:        id,
			Name:      track.dev.Name,
			Enabled:   track.enabled.Load(),
			Volume:    p.Volume(),
			Muted:     p.IsMuted(),
			DelayMs:   int(track.delayMs.Load()),
			Peak:      p.Peak(),
			Underruns: p.Underruns(),
			Overruns:  p.Overruns(),
			Fill:      p.Fill(),
		})
	}
	return res
}

// Close stops all streams, cleans up devices, and closes the engine.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}
	e.closed = true

	if e.captureStream != nil {
		_ = e.captureStream.Stop()
		_ = e.captureStream.Close()
		e.captureStream = nil
	}

	for _, track := range e.outputs {
		track.enabled.Store(false)
		_ = track.stream.Stop()
		_ = track.stream.Close()
	}
	e.outputs = make(map[string]*outputTrack)

	return nil
}

