# multiout

Play one laptop's audio to several output devices at once (e.g. USB + 3.5 mm
earphones), with per-device volume and delay. Linux + Windows first, macOS later.

## Setup (Ubuntu)

    sudo apt install build-essential     # cgo needs gcc
    go get github.com/gen2brain/malgo@latest
    go mod tidy

## Milestone 1 spike

    make test            # engine unit tests, with the race detector
    make list            # list playback + capture devices
    make tone A=0 B=1    # 440 Hz on device 0, 660 Hz on device 1

The capture list should include "Monitor of ..." entries. If it does not, the
malgo-on-PipeWire plan needs rethinking; see the Linux pactl alternative.

## Layout

    cmd/spike/         throwaway experiments
    internal/audio/    Backend interface + malgo implementation
    internal/engine/   ring buffer (now), resampler + drift control (next)
