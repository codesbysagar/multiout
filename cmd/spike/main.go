// Command spike is the milestone-1/milestone-2 experiment: list audio devices,
// play a test tone to two playback devices, or pipe captured audio into a playback device.
//
//	go run ./cmd/spike list
//	go run ./cmd/spike tone <playbackIndexA> <playbackIndexB>
//	go run ./cmd/spike capture-play <captureIndex> <playbackIndex>
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"multiout/internal/audio"
	"multiout/internal/engine"
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
	case "capture-play":
		if len(os.Args) != 4 {
			usage()
		}
		capturePlay(be, os.Args[2], os.Args[3])
	case "sink-play":
		if len(os.Args) < 3 {
			usage()
		}
		sinkPlay(be, os.Args[2:])
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
	devs, err := be.Devices(audio.Playback)
	if err != nil { // populate IDs
		fatal(err)
	}
	cfg := audio.Config{SampleRate: 48000, Channels: 2}

	var streams []audio.Stream
	for i, idxStr := range []string{a, b} {
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			fatal(fmt.Errorf("device index %q is not a number", idxStr))
		}
		if idx < 0 || idx >= len(devs) {
			fatal(fmt.Errorf("device index %d is out of range: only %d playback device(s) available (0 to %d). Run 'make list' to inspect", idx, len(devs), len(devs)-1))
		}
		freq := []float64{440, 660}[i] // different pitch per device
		s, err := be.OpenPlayback("playback:"+idxStr, cfg, sine(freq, cfg, 0.1))
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

func capturePlay(be audio.Backend, capIdxStr, playIdxStr string) {
	capDevs, err := be.Devices(audio.Capture)
	if err != nil {
		fatal(err)
	}
	playDevs, err := be.Devices(audio.Playback)
	if err != nil {
		fatal(err)
	}

	capIdx, err := strconv.Atoi(capIdxStr)
	if err != nil {
		fatal(fmt.Errorf("capture device index %q is not a number", capIdxStr))
	}
	if capIdx < 0 || capIdx >= len(capDevs) {
		fatal(fmt.Errorf("capture index %d out of range: %d capture device(s) found (run 'make list')", capIdx, len(capDevs)))
	}

	playIdx, err := strconv.Atoi(playIdxStr)
	if err != nil {
		fatal(fmt.Errorf("playback device index %q is not a number", playIdxStr))
	}
	if playIdx < 0 || playIdx >= len(playDevs) {
		fatal(fmt.Errorf("playback index %d out of range: %d playback device(s) found (run 'make list')", playIdx, len(playDevs)))
	}

	cfg := audio.Config{SampleRate: 48000, Channels: 2}
	// 2 seconds capacity
	r := engine.NewRing(cfg.SampleRate * cfg.Channels * 2)
	rd := r.NewReader()

	capStream, err := be.OpenCapture("capture:"+capIdxStr, cfg, func(in []float32) {
		r.Write(in)
	})
	if err != nil {
		fatal(fmt.Errorf("open capture: %w", err))
	}
	defer capStream.Close()

	playStream, err := be.OpenPlayback("playback:"+playIdxStr, cfg, func(out []float32) {
		n := rd.Read(out)
		for i := n; i < len(out); i++ {
			out[i] = 0 // fill underrun with silence
		}
	})
	if err != nil {
		fatal(fmt.Errorf("open playback: %w", err))
	}
	defer playStream.Close()

	if err := capStream.Start(); err != nil {
		fatal(err)
	}
	if err := playStream.Start(); err != nil {
		fatal(err)
	}

	fmt.Printf("Streaming capture:%s -> playback:%s. Press Ctrl+C to stop.\n", capIdxStr, playIdxStr)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Printf("\nStopping... Overruns: %d\n", rd.Overruns())
			return
		case <-ticker.C:
			fmt.Printf("Status: buffer fill=%d samples, overruns=%d\n", rd.Fill(), rd.Overruns())
		}
	}
}

func sinkPlay(be audio.Backend, outputIndices []string) {
	sinkMgr := audio.NewVirtualSinkManager()
	if err := sinkMgr.Setup(); err != nil {
		fatal(fmt.Errorf("setup virtual sink: %w", err))
	}
	defer func() {
		fmt.Println("\nRestoring original default audio sink...")
		_ = sinkMgr.Restore()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Re-enumerate capture devices after virtual sink setup
	capDevs, err := be.Devices(audio.Capture)
	if err != nil {
		fatal(err)
	}

	var virtualSinkMonitor audio.Device
	for _, d := range capDevs {
		if strings.Contains(strings.ToLower(d.Name), "monitor") && strings.Contains(strings.ToLower(d.Name), "multiout") {
			virtualSinkMonitor = d
			break
		}
	}
	if virtualSinkMonitor.ID == "" {
		fatal(fmt.Errorf("could not find monitor of virtual sink in capture devices"))
	}

	playDevs, err := be.Devices(audio.Playback)
	if err != nil {
		fatal(err)
	}

	eng := engine.NewEngine(be, audio.Config{SampleRate: 48000, Channels: 2})
	defer eng.Close()

	if err := eng.StartCapture(virtualSinkMonitor); err != nil {
		fatal(fmt.Errorf("start capture from virtual sink: %w", err))
	}

	for _, idxStr := range outputIndices {
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 || idx >= len(playDevs) {
			fatal(fmt.Errorf("invalid playback device index %q: must be between 0 and %d", idxStr, len(playDevs)-1))
		}
		if _, err := eng.AddOutput(playDevs[idx]); err != nil {
			fatal(fmt.Errorf("add output %s: %w", playDevs[idx].Name, err))
		}
		fmt.Printf("Routing audio to output [%d]: %s\n", idx, playDevs[idx].Name)
	}

	fmt.Println("\nVirtual sink 'multiout' active. Play video/audio in browser or VLC!")
	fmt.Println("Audio will play through your selected earphones simultaneously while laptop speakers stay silent.")
	fmt.Println("Press Ctrl+C to stop and restore normal desktop audio.")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			return
		case <-ticker.C:
			states := eng.Outputs()
			for _, s := range states {
				fmt.Printf("[%s] fill=%d peak=%.2f underruns=%d overruns=%d\n", s.Name, s.Fill, s.Peak, s.Underruns, s.Overruns)
			}
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: spike list | spike tone <A> <B> | spike capture-play <cap> <play> | spike sink-play <out1> <out2> ...")
	os.Exit(2)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
