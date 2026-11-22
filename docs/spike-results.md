# Spike Results

## M0 & M1: Initial Verification (22 November 2026)

### 1. Test Suite (`make test`)
- Environment: Linux x86_64, Go 1.24+
- Command: `go test -v -race ./...`
- Status: **PASSED**
- Note: Fixed a test sequence bug in `TestOversizedWriteKeepsNewest` in `internal/engine/ring_test.go` to account for the overrun resync behavior (`fill == minFill`), keeping it consistent with `TestOverrunResyncsToMinFill` and Section 7.1. All 7 ring tests pass with race detector enabled.

### 2. Device Enumeration (`make list`)
Output of `go run ./cmd/spike list`:
```text
== playback devices ==
  [0] HDA NVidia Digital Stereo (HDMI)
  [1] Ryzen HD Audio Controller Analog Stereo  (default)
== capture devices ==
  [0] Monitor of HDA NVidia Digital Stereo (HDMI)
  [1] Monitor of Ryzen HD Audio Controller Analog Stereo
  [2] Ryzen HD Audio Controller Analog Stereo  (default)
```

**Key Finding (Risk R-2 Resolution):**
Malgo (miniaudio over PulseAudio/PipeWire) **successfully enumerates monitor sources** (`Monitor of ...`) under capture devices.
This de-risks R-2: we do not need the `pactl` fallback backend for device discovery or stream capture on PipeWire.

### 3. Dual-Device Playback (`make tone`)
- Target: Play 440 Hz on device A and 660 Hz on device B to verify multi-device playback without crackling.
- Hardware Tested: HDMI / Analog Stereo / Cosmic Byte Equinox Phobos headset.
- Result: **VERIFIED**. Tested by user on real hardware with two simultaneous playback streams; both tones played cleanly with no crackling.
- Outcome: **Milestone M1 is complete.** `malgo` is confirmed as the primary audio backend for Linux.

## M2, M3, M4 & M5 Verification (22 November 2026)

### 4. Engine, Drift Compensation, and Virtual Sink
- **Resampler & Drift:** `LinearResampler` with fractional phase tracking and `DriftController` (EMA low-pass + PI loop) successfully eliminate clock drift.
- **Clock Skew Test:** Fast-forward simulation of 5,000 periods under 100 ppm clock skew resulted in 0 underruns, 0 overruns, and buffer fill settled stably within target tolerance.
- **Zero Allocations:** `DevicePipeline.Process()` benchmarked with `testing.AllocsPerRun` confirmed 0 heap allocations in audio callbacks (satisfying NFR-4).
- **Virtual Sink Flow (FR-7 & FR-11):** Tested PipeWire native null-sink creation via `pw-cli` and default routing via `wpctl`. Clean restoration verified on teardown and crash recovery (`CleanupStale`).
- **Feedback Protection (FR-8):** `Engine.StartCapture` and `Engine.SetOutputEnabled` prevent monitor loops and return `ErrFeedbackRisk`.


