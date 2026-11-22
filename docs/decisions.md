# Architecture & Design Decisions

## ADR-001: Backend Interface Extension for Audio Capture
- **Date:** 2026-11-22
- **Status:** Proposed / In-progress
- **Context:** Milestone M2 requires capturing audio from a source (such as a monitor sink or WASAPI loopback) and piping it into the engine ring buffer.
- **Decision:** Extend `audio.Backend` with `OpenCapture(deviceID string, cfg Config, onData func(in []float32)) (Stream, error)`.
- **Consequences:** Platform-specific capture (malgo on Linux/Windows) is isolated under `internal/audio/`. The engine and spike tools can consume captured audio via this common interface.

