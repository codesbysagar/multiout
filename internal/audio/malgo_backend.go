//go:build (linux || windows) && cgo

package audio

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// MalgoBackend implements Backend on top of miniaudio (via malgo). On Linux it
// talks to PulseAudio, which PipeWire also provides on current Ubuntu.
type MalgoBackend struct {
	ctx *malgo.AllocatedContext

	mu       sync.Mutex
	known    map[string]malgo.DeviceInfo // our ID -> miniaudio device info
	loopback map[string]bool             // our ID -> true if loopback device
}

func NewMalgoBackend() (*MalgoBackend, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("init audio context: %w", err)
	}
	return &MalgoBackend{
		ctx:      ctx,
		known:    map[string]malgo.DeviceInfo{},
		loopback: map[string]bool{},
	}, nil
}

func (b *MalgoBackend) Close() error {
	err := b.ctx.Uninit()
	b.ctx.Free()
	return err
}

func malgoType(k Kind) malgo.DeviceType {
	if k == Playback {
		return malgo.Playback
	}
	return malgo.Capture
}

// Devices lists devices of one kind. On Windows, Capture devices include
// WASAPI loopback endpoints (System Audio and render device loopbacks)
// so system output and media players can be captured.
func (b *MalgoBackend) Devices(kind Kind) ([]Device, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if kind == Capture && runtime.GOOS == "windows" {
		var out []Device

		// 1. On Windows, enumerate playback endpoints as Loopback capture sources
		pInfos, err := b.ctx.Devices(malgo.Playback)
		if err == nil {
			// Virtual entry for default system playback
			defaultID := "loopback:default"
			b.loopback[defaultID] = true
			out = append(out, Device{
				ID:        defaultID,
				Name:      "System Audio (Default Output)",
				Kind:      Capture,
				IsDefault: true,
			})

			for i := range pInfos {
				id := fmt.Sprintf("loopback:%d", i)
				b.known[id] = pInfos[i]
				b.loopback[id] = true
				out = append(out, Device{
					ID:        id,
					Name:      fmt.Sprintf("%s (Loopback)", pInfos[i].Name()),
					Kind:      Capture,
					IsDefault: false,
				})
			}
		}

		// 2. Enumerate physical capture devices (microphones)
		cInfos, err := b.ctx.Devices(malgo.Capture)
		if err != nil {
			return nil, fmt.Errorf("list capture devices: %w", err)
		}
		for i := range cInfos {
			id := fmt.Sprintf("capture:%d", i)
			b.known[id] = cInfos[i]
			out = append(out, Device{
				ID:        id,
				Name:      fmt.Sprintf("%s (Mic)", cInfos[i].Name()),
				Kind:      Capture,
				IsDefault: false,
			})
		}
		return out, nil
	}

	infos, err := b.ctx.Devices(malgoType(kind))
	if err != nil {
		return nil, fmt.Errorf("list %s devices: %w", kind, err)
	}

	out := make([]Device, 0, len(infos))
	for i := range infos {
		id := fmt.Sprintf("%s:%d", kind, i)
		b.known[id] = infos[i]
		out = append(out, Device{
			ID:        id,
			Name:      infos[i].Name(),
			Kind:      kind,
			IsDefault: infos[i].IsDefault != 0,
		})
	}
	return out, nil
}

type malgoStream struct{ dev *malgo.Device }

func (s *malgoStream) Start() error { return s.dev.Start() }
func (s *malgoStream) Stop() error  { return s.dev.Stop() }
func (s *malgoStream) Close() error { s.dev.Uninit(); return nil }

func (b *MalgoBackend) OpenPlayback(deviceID string, cfg Config, fill FillFunc) (Stream, error) {
	b.mu.Lock()
	info, ok := b.known[deviceID]
	b.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown device %q (call Devices first)", deviceID)
	}

	dc := malgo.DefaultDeviceConfig(malgo.Playback)
	dc.Playback.Format = malgo.FormatF32
	dc.Playback.Channels = uint32(cfg.Channels)
	dc.SampleRate = uint32(cfg.SampleRate)
	dc.Playback.DeviceID = info.ID.Pointer()

	channels := cfg.Channels
	callbacks := malgo.DeviceCallbacks{
		Data: func(out, _ []byte, frames uint32) {
			n := int(frames) * channels
			if n == 0 || len(out) < n*4 {
				return
			}
			// Zero-copy view of the output buffer as float32.
			fill(unsafe.Slice((*float32)(unsafe.Pointer(&out[0])), n))
		},
	}

	dev, err := malgo.InitDevice(b.ctx.Context, dc, callbacks)
	if err != nil {
		return nil, fmt.Errorf("open playback %q: %w", deviceID, err)
	}
	return &malgoStream{dev: dev}, nil
}

func (b *MalgoBackend) OpenCapture(deviceID string, cfg Config, onData CaptureFunc) (Stream, error) {
	b.mu.Lock()
	isLoopback := b.loopback[deviceID]
	info, hasInfo := b.known[deviceID]
	b.mu.Unlock()

	var dc malgo.DeviceConfig
	if isLoopback {
		dc = malgo.DefaultDeviceConfig(malgo.Loopback)
		dc.Capture.Format = malgo.FormatF32
		dc.Capture.Channels = uint32(cfg.Channels)
		dc.SampleRate = uint32(cfg.SampleRate)
		if deviceID == "loopback:default" {
			dc.Capture.DeviceID = nil // default render endpoint in WASAPI
		} else if hasInfo {
			dc.Capture.DeviceID = info.ID.Pointer()
		} else {
			return nil, fmt.Errorf("unknown loopback device %q (call Devices first)", deviceID)
		}
	} else {
		if !hasInfo {
			return nil, fmt.Errorf("unknown device %q (call Devices first)", deviceID)
		}
		dc = malgo.DefaultDeviceConfig(malgo.Capture)
		dc.Capture.Format = malgo.FormatF32
		dc.Capture.Channels = uint32(cfg.Channels)
		dc.SampleRate = uint32(cfg.SampleRate)
		dc.Capture.DeviceID = info.ID.Pointer()
	}

	channels := cfg.Channels
	callbacks := malgo.DeviceCallbacks{
		Data: func(_, in []byte, frames uint32) {
			n := int(frames) * channels
			if n == 0 || len(in) < n*4 {
				return
			}
			// Zero-copy view of the input buffer as float32.
			onData(unsafe.Slice((*float32)(unsafe.Pointer(&in[0])), n))
		},
	}

	dev, err := malgo.InitDevice(b.ctx.Context, dc, callbacks)
	if err != nil {
		return nil, fmt.Errorf("open capture %q: %w", deviceID, err)
	}
	return &malgoStream{dev: dev}, nil
}
