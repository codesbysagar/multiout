package engine

import (
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func readN(rd *Reader, max int) []float32 {
	buf := make([]float32, max)
	return buf[:rd.Read(buf)]
}

func TestRoundTrip(t *testing.T) {
	r := NewRing(8)
	rd := r.NewReader()

	r.Write([]float32{1, 2, 3})
	if got := readN(rd, 5); !slices.Equal(got, []float32{1, 2, 3}) {
		t.Fatalf("got %v", got)
	}
	if got := readN(rd, 5); len(got) != 0 {
		t.Fatalf("expected empty read, got %v", got)
	}
}

func TestWraparound(t *testing.T) {
	r := NewRing(4)
	rd := r.NewReader()

	r.Write([]float32{1, 2, 3})
	readN(rd, 3)
	r.Write([]float32{4, 5, 6}) // wraps around the end of the buffer
	if got := readN(rd, 3); !slices.Equal(got, []float32{4, 5, 6}) {
		t.Fatalf("got %v", got)
	}
}

func TestIndependentReaders(t *testing.T) {
	r := NewRing(16)
	a := r.NewReader()

	r.Write([]float32{1, 2, 3, 4})
	if got := readN(a, 2); !slices.Equal(got, []float32{1, 2}) {
		t.Fatalf("a first read: %v", got)
	}

	b := r.NewReader() // starts at the current head
	if b.Fill() != 0 {
		t.Fatalf("new reader should start empty, fill=%d", b.Fill())
	}

	r.Write([]float32{5})
	if got := readN(b, 10); !slices.Equal(got, []float32{5}) {
		t.Fatalf("b: %v", got)
	}
	if got := readN(a, 10); !slices.Equal(got, []float32{3, 4, 5}) {
		t.Fatalf("a second read: %v", got)
	}
}

func TestMinFillHoldsBackDelay(t *testing.T) {
	r := NewRing(16)
	rd := r.NewReader()
	rd.SetMinFill(3)

	r.Write([]float32{1, 2})
	if got := readN(rd, 10); len(got) != 0 {
		t.Fatalf("expected nothing while within delay, got %v", got)
	}

	r.Write([]float32{3, 4, 5})
	if got := readN(rd, 10); !slices.Equal(got, []float32{1, 2}) {
		t.Fatalf("got %v", got)
	}
	if rd.Fill() != 3 {
		t.Fatalf("fill should settle at minFill, got %d", rd.Fill())
	}
}

func TestOverrunResyncsToMinFill(t *testing.T) {
	r := NewRing(4)
	rd := r.NewReader()
	rd.SetMinFill(2)

	r.Write([]float32{1, 2, 3})
	r.Write([]float32{4, 5, 6}) // reader is now 6 behind a 4-sample ring

	if got := readN(rd, 10); len(got) != 0 {
		t.Fatalf("expected nothing right after resync, got %v", got)
	}
	if rd.Overruns() != 1 {
		t.Fatalf("overruns=%d, want 1", rd.Overruns())
	}
	if rd.Fill() != 2 {
		t.Fatalf("fill=%d, want 2", rd.Fill())
	}

	r.Write([]float32{7})
	if got := readN(rd, 10); !slices.Equal(got, []float32{5}) {
		t.Fatalf("got %v", got)
	}
}

func TestOversizedWriteKeepsNewest(t *testing.T) {
	r := NewRing(4)
	rd := r.NewReader()
	rd.SetMinFill(2)

	r.Write([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) // bigger than the ring
	readN(rd, 10)                                     // resyncs to minFill behind head
	r.Write([]float32{11, 12})

	if got := readN(rd, 10); !slices.Equal(got, []float32{9, 10}) {
		t.Fatalf("got %v", got)
	}
}

// Run with -race. One writer and one reader hammer the ring; values must
// never be reordered or repeated, though gaps are allowed after an overrun.
func TestConcurrentWriterReader(t *testing.T) {
	const total = 200000
	r := NewRing(1 << 12)
	rd := r.NewReader()

	var done atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		chunk := make([]float32, 64)
		next := 1
		for next <= total {
			n := min(len(chunk), total-next+1)
			for i := 0; i < n; i++ {
				chunk[i] = float32(next + i)
			}
			r.Write(chunk[:n])
			next += n
			runtime.Gosched()
		}
		done.Store(true)
	}()

	buf := make([]float32, 100)
	var last float32
	for {
		n := rd.Read(buf)
		for _, v := range buf[:n] {
			if v <= last {
				t.Fatalf("out of order: %v after %v", v, last)
			}
			last = v
		}
		if n == 0 {
			if done.Load() && rd.Fill() == 0 {
				break
			}
			runtime.Gosched()
		}
	}
	wg.Wait()
}

func TestReaderPeekAndAdvance(t *testing.T) {
	r := NewRing(16)
	rd := r.NewReader()

	r.Write([]float32{10, 20, 30, 40})

	peekBuf := make([]float32, 2)
	n := rd.Peek(peekBuf)
	if n != 2 || !slices.Equal(peekBuf, []float32{10, 20}) {
		t.Fatalf("peek got %v", peekBuf)
	}

	// Peek should not have moved cursor
	if rd.Fill() != 4 {
		t.Fatalf("fill should be 4, got %d", rd.Fill())
	}

	// Second peek should return identical data
	n2 := rd.Peek(peekBuf)
	if n2 != 2 || !slices.Equal(peekBuf, []float32{10, 20}) {
		t.Fatalf("second peek got %v", peekBuf)
	}

	// Advance by 2
	rd.Advance(2)
	if rd.Fill() != 2 {
		t.Fatalf("fill should be 2 after advance, got %d", rd.Fill())
	}

	// Peek now returns next samples
	rd.Peek(peekBuf)
	if !slices.Equal(peekBuf, []float32{30, 40}) {
		t.Fatalf("peek after advance got %v", peekBuf)
	}
}
