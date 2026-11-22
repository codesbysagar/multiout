//go:build !(linux || windows) || !cgo

package audio

import "errors"

// MalgoBackend is a stub implementation for unsupported platforms or when CGO is disabled.
type MalgoBackend struct{}

// NewMalgoBackend returns an error on unsupported platforms.
func NewMalgoBackend() (Backend, error) {
	return nil, errors.New("malgo audio backend is only supported on Linux and Windows with CGO enabled")
}

func (b *MalgoBackend) Devices(kind Kind) ([]Device, error) {
	return nil, errors.New("unsupported platform")
}

func (b *MalgoBackend) OpenPlayback(deviceID string, cfg Config, fill FillFunc) (Stream, error) {
	return nil, errors.New("unsupported platform")
}

func (b *MalgoBackend) OpenCapture(deviceID string, cfg Config, onData CaptureFunc) (Stream, error) {
	return nil, errors.New("unsupported platform")
}

func (b *MalgoBackend) Close() error {
	return nil
}

