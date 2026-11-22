// Package engine holds the hardware-independent audio logic: the shared ring
// buffer and (later) the resampler, drift controller and mixer. It works on
// plain interleaved []float32 samples so it can be tested without any audio
// device.
package engine

import "sync"

// Ring is a single-writer, multi-reader ring buffer of interleaved float32
// samples. The capture callback writes into it; each output device owns a
// Reader with its own cursor, so a slow device can never block the others.
//
// All operations take one mutex. Critical sections are short copies, which is
// fine for 10 ms audio periods. If profiling ever shows contention we can
// move to a lock-free design, but correctness first.
type Ring struct {
	mu      sync.Mutex
	buf     []float32
	written uint64 // total samples ever written (monotonic)
}

// NewRing creates a ring that holds capacity samples (not frames).
func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		panic("engine: ring capacity must be positive")
	}
	return &Ring{buf: make([]float32, capacity)}
}

// Cap returns the ring capacity in samples.
func (r *Ring) Cap() int { return len(r.buf) }

// Write appends samples. Old data is overwritten when the ring is full. If p
// is larger than the ring, only the newest Cap() samples are kept.
func (r *Ring) Write(p []float32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c := len(r.buf)
	if len(p) > c {
		skip := len(p) - c
		r.written += uint64(skip)
		p = p[skip:]
	}
	start := int(r.written % uint64(c))
	n := copy(r.buf[start:], p)
	copy(r.buf, p[n:])
	r.written += uint64(len(p))
}

// Reader is one consumer's cursor into a Ring. Create it with Ring.NewReader.
// A Reader is safe to use from one audio goroutine while another goroutine
// calls SetMinFill, Fill or Overruns.
type Reader struct {
	ring     *Ring
	pos      uint64 // next sample to read (monotonic, same scale as Ring.written)
	minFill  uint64 // samples held back; this is the per-device delay
	overruns uint64
}

// NewReader returns a reader positioned at the current write head, so it only
// sees audio written from now on.
func (r *Ring) NewReader() *Reader {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Reader{ring: r, pos: r.written}
}

// SetMinFill sets how many samples the reader holds back before returning
// data. This is how per-device delay is implemented: a larger value means the
// device plays later. Values above the ring capacity are clamped.
func (rd *Reader) SetMinFill(samples int) {
	rd.ring.mu.Lock()
	defer rd.ring.mu.Unlock()
	if samples < 0 {
		samples = 0
	}
	if samples > len(rd.ring.buf) {
		samples = len(rd.ring.buf)
	}
	rd.minFill = uint64(samples)
}

// Fill returns how many unread samples are buffered for this reader,
// including the held-back delay. The drift controller watches this value.
func (rd *Reader) Fill() int {
	rd.ring.mu.Lock()
	defer rd.ring.mu.Unlock()
	return int(rd.ring.written - rd.pos)
}

// Overruns returns how many times the writer lapped this reader (audio was
// lost because the reader fell too far behind).
func (rd *Reader) Overruns() uint64 {
	rd.ring.mu.Lock()
	defer rd.ring.mu.Unlock()
	return rd.overruns
}

// Read copies up to len(dst) samples into dst and returns how many were
// copied. It returns fewer (possibly zero) when the buffered data is within
// the held-back delay; the caller should fill the remainder with silence.
func (rd *Reader) Read(dst []float32) int {
	r := rd.ring
	r.mu.Lock()
	defer r.mu.Unlock()

	c := uint64(len(r.buf))
	fill := r.written - rd.pos
	if fill > c {
		// Writer lapped us: the oldest samples are gone. Resync so that the
		// reader sits minFill samples behind the write head again.
		rd.overruns++
		rd.pos = r.written - min(rd.minFill, c)
		fill = r.written - rd.pos
	}

	var avail uint64
	if fill > rd.minFill {
		avail = fill - rd.minFill
	}
	n := int(min(uint64(len(dst)), avail))
	if n == 0 {
		return 0
	}

	idx := int(rd.pos % c)
	k := copy(dst[:n], r.buf[idx:])
	copy(dst[k:n], r.buf)
	rd.pos += uint64(n)
	return n
}
