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

