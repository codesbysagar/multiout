// Package audio defines the platform-neutral audio backend interface. The
// engine only talks to this interface, so the malgo backend can be swapped for
// a native one (e.g. PipeWire/pactl on Linux) without touching engine code.
package audio

// Kind is the direction of an audio device.
type Kind int

const (
	Playback Kind = iota
	Capture
)

func (k Kind) String() string {
	if k == Playback {
		return "playback"
	}
	return "capture"
}

// Device describes one audio endpoint. ID is only valid for the Backend that
// produced it, and only until the next call to Devices.
type Device struct {
	ID        string
	Name      string
	Kind      Kind
	IsDefault bool
}

// Config is the stream format. Samples are always interleaved float32.
type Config struct {
	SampleRate int
	Channels   int
}

// FillFunc is called on the audio thread to produce samples for playback.
// Keep it fast and allocation-free: copy from a buffer, nothing else.
type FillFunc func(out []float32)

// CaptureFunc is called on the audio thread when captured samples are available.
// Keep it fast and allocation-free: copy to a buffer, nothing else.
type CaptureFunc func(in []float32)

// Stream is an open stream (playback or capture).
type Stream interface {
	Start() error
	Stop() error
	Close() error
}

// Backend abstracts the OS audio system.
type Backend interface {
	Devices(kind Kind) ([]Device, error)
	OpenPlayback(deviceID string, cfg Config, fill FillFunc) (Stream, error)
	OpenCapture(deviceID string, cfg Config, onData CaptureFunc) (Stream, error)
	Close() error
}
