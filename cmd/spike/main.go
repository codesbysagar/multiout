// Command spike is the milestone-1 experiment: list audio devices and play a
// different test tone to two playback devices at the same time.
//
//	go run ./cmd/spike list
//	go run ./cmd/spike tone <playbackIndexA> <playbackIndexB>
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"multiout/internal/audio"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	be, err := audio.NewMalgoBackend()
	if err != nil {
		fatal(err)
	}
	defer be.Close()

	switch os.Args[1] {
	case "list":
		list(be)
	case "tone":
		if len(os.Args) != 4 {
			usage()
		}
		tone(be, os.Args[2], os.Args[3])
	default:
		usage()
	}
}

func list(be audio.Backend) {
	for _, kind := range []audio.Kind{audio.Playback, audio.Capture} {
		devs, err := be.Devices(kind)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("== %s devices ==\n", kind)
		for i, d := range devs {
			def := ""
			if d.IsDefault {
				def = "  (default)"
			}
			fmt.Printf("  [%d] %s%s\n", i, d.Name, def)
		}
	}
	fmt.Println("\nLook for entries containing \"Monitor\" under capture: that is the system-audio source.")
}

func tone(be audio.Backend, a, b string) {
	if _, err := be.Devices(audio.Playback); err != nil { // populate IDs
		fatal(err)
	}
	cfg := audio.Config{SampleRate: 48000, Channels: 2}

	var streams []audio.Stream
	for i, idx := range []string{a, b} {
		if _, err := strconv.Atoi(idx); err != nil {
			fatal(fmt.Errorf("device index %q is not a number", idx))
		}
		freq := []float64{440, 660}[i] // different pitch per device
		s, err := be.OpenPlayback("playback:"+idx, cfg, sine(freq, cfg, 0.1))
		if err != nil {
			fatal(err)
		}
		defer s.Close()
		streams = append(streams, s)
	}
	for _, s := range streams {
		if err := s.Start(); err != nil {
			fatal(err)
		}
	}
	fmt.Println("Playing 440 Hz on device A and 660 Hz on device B for 10 seconds...")
	time.Sleep(10 * time.Second)
}

func sine(freq float64, cfg audio.Config, gain float32) audio.FillFunc {
	phase := 0.0
	step := 2 * math.Pi * freq / float64(cfg.SampleRate)
	ch := cfg.Channels
	return func(out []float32) {
		for i := 0; i+ch <= len(out); i += ch {
			v := float32(math.Sin(phase)) * gain
			phase += step
			if phase > 2*math.Pi {
				phase -= 2 * math.Pi
			}
			for c := 0; c < ch; c++ {
				out[i+c] = v
			}
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: spike list | spike tone <playbackIndexA> <playbackIndexB>")
	os.Exit(2)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
