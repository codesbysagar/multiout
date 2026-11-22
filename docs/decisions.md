# Architecture & Design Decisions

## ADR-001: Backend Interface Extension for Audio Capture
- **Date:** 2026-11-22
- **Status:** Accepted
- **Context:** Milestone M2 requires capturing audio from a source (such as a monitor sink or WASAPI loopback) and piping it into the engine ring buffer.
- **Decision:** Extend `audio.Backend` with `OpenCapture(deviceID string, cfg Config, onData func(in []float32)) (Stream, error)`.
- **Consequences:** Platform-specific capture (malgo on Linux/Windows) is isolated under `internal/audio/`. The engine and spike tools can consume captured audio via this common interface.

## ADR-002: Audio Engine Pipeline Architecture (Resampler, Drift, Gain, Peek/Advance)
- **Date:** 2026-11-22
- **Status:** Accepted
- **Context:** Milestones M3 and M4 require multi-device fan-out, smooth gain and mute ramping, and drift compensation without dropping or double-counting samples.
- **Decision:**
  1. Extended `Reader` with `Peek(dst)` and `Advance(samples)` to allow lookahead for linear interpolation without discarding samples ahead of fractional boundaries.
  2. Implemented `LinearResampler` preserving fractional phase and single-frame boundary history across audio callbacks.
  3. Implemented `DriftController` using exponential moving average (EMA) low-pass filtering and a PI control loop clamped to $\pm 5000$ ppm ($\pm 0.5\%$).
  4. Implemented `GainController` with 10 ms linear ramping and atomic peak level calculation for metering.
  5. Assembled `DevicePipeline` which is 100% allocation-free (`testing.AllocsPerRun == 0.0`) in audio callbacks, satisfying NFR-4.
- **Consequences:** Zero underruns and bounded buffer fill verified across clock skew simulations (100 ppm drift).

## ADR-003: Linux Virtual Sink Management via PipeWire (`pw-cli` / `wpctl`)
- **Date:** 2026-11-22
- **Status:** Accepted
- **Context:** Milestone M5 (FR-7) requires creating a virtual null sink so laptop speakers can be silenced while audio is captured through its monitor. Modern Ubuntu systems run PipeWire where `pactl` (`pulseaudio-utils`) is not installed by default.
- **Decision:** Manage virtual sinks primarily using PipeWire's standard `pw-cli` (`pw-cli create-node adapter ... support.null-audio-sink`) and `wpctl set-default`, falling back to `pactl` when available.
- **Consequences:** Eliminates external runtime dependencies beyond standard desktop PipeWire tools.


