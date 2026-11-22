package audio

import (
	"fmt"
	"sync"
)

// FakeStream represents an in-memory simulated audio stream.
type FakeStream struct {
	devID  string
	kind   Kind
	cfg    Config
	fill   FillFunc
	onData CaptureFunc

	mu      sync.Mutex
	running bool
	closed  bool
}

func (s *FakeStream) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("stream closed")
	}
	s.running = true
	return nil
}

func (s *FakeStream) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	return nil
}

func (s *FakeStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.closed = true
	return nil
}

// IsRunning reports whether the stream is active.
func (s *FakeStream) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// TickPlayback invokes the playback callback with a buffer of length frames * channels.
func (s *FakeStream) TickPlayback(frames int, buf []float32) {
	if !s.IsRunning() || s.fill == nil {
		return
	}
	n := frames * s.cfg.Channels
	if len(buf) < n {
		buf = make([]float32, n)
	}
	s.fill(buf[:n])
}

// TickCapture feeds the given samples into the capture callback.
func (s *FakeStream) TickCapture(samples []float32) {
	if !s.IsRunning() || s.onData == nil {
		return
	}
	s.onData(samples)
}

// FakeBackend is an in-memory Backend for testing.
type FakeBackend struct {
	mu              sync.Mutex
	devices         map[Kind][]Device
	playbackStreams map[string]*FakeStream
	captureStreams  map[string]*FakeStream
	closed          bool
}

func NewFakeBackend() *FakeBackend {
	return &FakeBackend{
		devices: map[Kind][]Device{
			Playback: {
				{ID: "fake:playback:0", Name: "Fake Headset", Kind: Playback, IsDefault: true},
				{ID: "fake:playback:1", Name: "Fake Speakers", Kind: Playback, IsDefault: false},
			},
			Capture: {
				{ID: "fake:capture:0", Name: "Monitor of Fake Headset", Kind: Capture, IsDefault: true},
				{ID: "fake:capture:1", Name: "Fake Microphone", Kind: Capture, IsDefault: false},
			},
		},
		playbackStreams: make(map[string]*FakeStream),
		captureStreams:  make(map[string]*FakeStream),
	}
}

func (f *FakeBackend) Devices(kind Kind) ([]Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, fmt.Errorf("backend closed")
	}
	devs := f.devices[kind]
	out := make([]Device, len(devs))
	copy(out, devs)
	return out, nil
}

func (f *FakeBackend) SetDevices(kind Kind, devs []Device) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.devices[kind] = devs
}

func (f *FakeBackend) OpenPlayback(deviceID string, cfg Config, fill FillFunc) (Stream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, fmt.Errorf("backend closed")
	}
	s := &FakeStream{
		devID: deviceID,
		kind:  Playback,
		cfg:   cfg,
		fill:  fill,
	}
	f.playbackStreams[deviceID] = s
	return s, nil
}

func (f *FakeBackend) OpenCapture(deviceID string, cfg Config, onData CaptureFunc) (Stream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, fmt.Errorf("backend closed")
	}
	s := &FakeStream{
		devID:  deviceID,
		kind:   Capture,
		cfg:    cfg,
		onData: onData,
	}
	f.captureStreams[deviceID] = s
	return s, nil
}

func (f *FakeBackend) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	for _, s := range f.playbackStreams {
		_ = s.Close()
	}
	for _, s := range f.captureStreams {
		_ = s.Close()
	}
	return nil
}

func (f *FakeBackend) GetPlaybackStream(deviceID string) *FakeStream {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.playbackStreams[deviceID]
}

func (f *FakeBackend) GetCaptureStream(deviceID string) *FakeStream {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.captureStreams[deviceID]
}
