package audio

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// MalgoBackend implements Backend on top of miniaudio (via malgo). On Linux it
// talks to PulseAudio, which PipeWire also provides on current Ubuntu.
type MalgoBackend struct {
	ctx *malgo.AllocatedContext

	mu    sync.Mutex
	known map[string]malgo.DeviceInfo // our ID -> miniaudio device info
}

func NewMalgoBackend() (*MalgoBackend, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("init audio context: %w", err)
	}
	return &MalgoBackend{ctx: ctx, known: map[string]malgo.DeviceInfo{}}, nil
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

// Devices lists devices of one kind. IDs look like "playback:0" and are index
// based for now; stable IDs (for remembering per-device settings and handling
// hot-plug) come later.
func (b *MalgoBackend) Devices(kind Kind) ([]Device, error) {
	infos, err := b.ctx.Devices(malgoType(kind))
	if err != nil {
		return nil, fmt.Errorf("list %s devices: %w", kind, err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

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
