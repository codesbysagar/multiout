<p align="center">
  <img src="internal/ui/multiout-logo-rounded.png" width="128" height="128" alt="Multiout Logo" />
</p>

<h1 align="center">Multiout</h1>

<p align="center">
  <b>Simultaneous Multi-Device Audio Output with Real-Time Drift Compensation</b>
</p>

<p align="center">
  <a href="https://github.com/codesbysagar/multiout/releases"><img src="https://img.shields.io/github/v/release/codesbysagar/multiout?color=6366f1&style=flat-square" alt="GitHub Release"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="#"><img src="https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-blue?style=flat-square" alt="Platform"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg?style=flat-square" alt="License: MIT"></a>
  <a href="https://github.com/codesbysagar/multiout/actions"><img src="https://img.shields.io/badge/Tests-Passing-brightgreen?style=flat-square" alt="Tests"></a>
</p>

<p align="center">
  Share audio seamlessly across multiple earphones, headphones, and external speakers (e.g. <b>3.5 mm jack</b> + <b>USB/Type-C DAC</b> + <b>Bluetooth headset</b>) at the same time.<br>
  Built with <b>Go</b>, <b>Fyne</b>, and <b>miniaudio</b>, featuring real-time closed-loop clock drift compensation, per-device logarithmic decibel controls, millisecond latency alignment, and loopback feedback protection.
</p>

---

## 📑 Table of Contents

- [The Problem & Multiout's Solution](#-the-problem--multiouts-solution)
- [Key Features](#-key-features)
- [Supported Platforms](#-supported-platforms)
- [Installation & Downloads](#-installation--downloads)
  - [Linux (Debian / Ubuntu .deb)](#linux-debian--ubuntu-deb)
  - [Windows (.exe Standalone)](#windows-exe-standalone)
  - [Install from Source (Linux)](#install-from-source-linux)
- [Building from Source](#-building-from-source)
- [Technical Architecture](#-technical-architecture)
- [User Guide](#-user-guide)
- [Troubleshooting & FAQ](#-troubleshooting--faq)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## ⚡ The Problem & Multiout's Solution

### The Challenge with Multi-Device Audio
When you connect two independent audio devices (e.g., wired headphones plugged into a 3.5 mm audio jack and a pair of wireless Bluetooth earbuds), both soundcards run on separate, unsynchronized physical crystal oscillators.
- **Clock Drift:** One soundcard consumes samples at $48,001\text{ Hz}$ while the other consumes at $47,998\text{ Hz}$. Over minutes, audio drifts apart, causing audible echoes, buffer underruns (clicks, pops), or buffer overruns (dropped audio).
- **Latency Discrepancy:** Bluetooth codecs (AAC, SBC, LDAC) introduce $80\text{--}200\text{ ms}$ of transmission delay, whereas analog jacks have near-zero latency, resulting in an unbearable echo chamber when listening together.
- **Volume Mismatch:** Different earphones have different impedances and sensitivities. A standard volume slider makes one pair deafening while the other is whisper-quiet.

### How Multiout Solves It
1. **Closed-Loop PI Drift Controller:** Dynamically monitors the fill level of each output's ring buffer with an Exponential Moving Average (EMA) low-pass filter and continuously steers a Proportional-Integral (PI) control loop to adjust resampling ratios smoothly within $\pm 5000\text{ ppm}$ ($\pm 0.5\%$), keeping all soundcards sample-synchronized indefinitely without audible pitch changes.
2. **Sub-Sample Fractional Linear Resampler:** Adjusts playback speeds smoothly in real time without dropping frames or introducing harmonic distortion.
3. **Per-Device Latency Delay Sync ($0\text{ to }300\text{ ms}$):** Add millisecond delay buffers to low-latency devices to match higher-latency Bluetooth or wireless outputs perfectly.
4. **Per-Device Logarithmic Decibel Controls ($-60.0\text{ dB}$ to $+12.0\text{ dB}$):** Calibrated logarithmic decibel scaling with $+12\text{ dB}$ headroom boost ensures optimal loudness matching across high-impedance headphones and sensitive IEMs.
5. **Zero-Allocation Real-Time Audio Engine:** Audio callbacks operate on lock-free atomic ring buffers with zero dynamic heap allocations ($0.0\text{ allocs/run}$), eliminating garbage collection pauses during playback.
6. **Automatic Loopback Protection:** Prevents accidental feedback loops by validating input/output pairings before routing.

---

## ✨ Key Features

- **Simultaneous Audio Fan-Out:** Route system audio to any number of connected output devices simultaneously.
- **Real-Time Drift Compensation:** Automated closed-loop resampler maintains sync across independent hardware DAC clocks.
- **Precision Decibel Volume:** Individual sliders from $-60.0\text{ dB}$ down to mute, up to $+12.0\text{ dB}$ boost in $0.5\text{ dB}$ steps.
- **Custom Delay Offset:** Fine-tune latency between $0\text{ ms}$ and $300\text{ ms}$ per device to eliminate Bluetooth echo.
- **Modern DaisyUI & OpenLAMT Aesthetic:** Clean, distraction-free desktop interface built with custom Fyne widgets.
- **Dynamic Theme Switcher:** Instant switching between **Dark** (`#13141F`), **Light** (`#F4F5FB`), and **System Default** modes.
- **Peak Level Meters:** Real-time dual-gradient visual audio meters for each active output.
- **Execution Log Terminal:** Built-in live logging panel to monitor device events, underruns, and streaming health.
- **Seamless PipeWire / PulseAudio Integration:** Automatically provisions a dedicated virtual null sink on Linux and restores default routing on exit.

---

## 🖥️ Supported Platforms

| Platform | Status | Backend / Technology |
| :--- | :---: | :--- |
| **Linux (Ubuntu, Debian, Fedora, Arch)** | **Supported** | PipeWire (`pw-cli` / `wpctl`) & ALSA via `malgo` |
| **Windows (10 / 11 64-bit)** | **Supported** | Native WASAPI Loopback Capture & Playback |
| **macOS (Apple Silicon & Intel)** | *Planned* | CoreAudio Aggregate Device / Virtual Audio Driver |

---

## 📦 Installation & Downloads

Pre-built binaries and packages for **Linux** and **Windows** are available on the [**GitHub Releases page**](https://github.com/codesbysagar/multiout/releases).

### Linux (Debian / Ubuntu .deb)

1. Download the latest `.deb` package from the [Releases](https://github.com/codesbysagar/multiout/releases) page:
   ```bash
   wget https://github.com/codesbysagar/multiout/releases/download/v1.0.0/multiout_1.0.0_amd64.deb
   ```
2. Install the package:
   ```bash
   sudo dpkg -i multiout_1.0.0_amd64.deb
   ```
3. If any dependencies are missing, resolve them automatically:
   ```bash
   sudo apt-get install -f
   ```
4. Launch **Multiout** from your desktop application menu or by running `multiout` in terminal.

### Windows (.exe Standalone & Installer)

**Option A: Standalone Executable (Portable)**
1. Download `multiout.exe` from the [Releases](https://github.com/codesbysagar/multiout/releases) page.
2. Place `multiout.exe` in any folder of your choice and double-click to run.

**Option B: Windows Installer Setup**
1. Download `multiout_setup_x64.exe` from [Releases](https://github.com/codesbysagar/multiout/releases).
2. Run the wizard to install Multiout to your system with Start Menu and Desktop shortcuts.

**Option C: PowerShell Setup (From Source)**
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1 -Build
```

> [!TIP]
> **Windows Audio Setup Tip (Avoid Dual-Voice Echo):**
> On Windows, audio plays directly to your default hardware soundcard. If you route system audio using Multiout to the same device, Windows will play raw audio *plus* Multiout's resampled audio, causing an echo.
> To route clean audio:
> 1. Install a free Virtual Audio Cable, such as **[VB-Audio Cable](https://vb-audio.com/Cable/)** (or select a secondary monitor display output).
> 2. Set **CABLE Input** (or your monitor) as the default Windows Output.
> 3. In Multiout, select **CABLE Output** (or your monitor's loopback) as the input source, and toggle ON all the physical earphones/headphones you want sound to play through. Multiout automatically detects your virtual cable!

### Install from Source (Linux)

If you have Go installed, you can build and install Multiout directly into your user environment (`~/.local/bin` and desktop application launcher):

```bash
git clone https://github.com/codesbysagar/multiout.git
cd multiout
make install
```

---

## 🛠️ Building from Source

### Prerequisites

#### Linux (Ubuntu / Debian)
- **Go 1.24+**: [Download Go](https://golang.org/dl/)
- **C Compiler & Graphics Libraries:**
  ```bash
  sudo apt update
  sudo apt install -y build-essential libgl1-mesa-dev xorg-dev pipewire wireplumber
  ```
- *(Optional, for cross-compiling Windows on Linux)*:
  ```bash
  sudo apt install -y gcc-mingw-w64-x86-64
  ```

#### Windows
- **Go 1.24+**
- **MinGW-w64 (GCC)** (e.g. via MSYS2 or WinLibs)

### Available Make Targets

| Command | Description |
| :--- | :--- |
| `make build` | Compiles Linux executable to `bin/multiout` |
| `make run` | Runs Multiout directly with `go run ./cmd/multiout` |
| `make deb` | Packages standard Debian `.deb` installer into `bin/` |
| `make exe` | Cross-compiles Windows 64-bit GUI executable (`bin/multiout.exe`) |
| `make install` | Installs binary, `.desktop` launcher, and icons to `~/.local/` |
| `make test` | Executes all unit and drift simulation tests with race detection (`-race`) |
| `make vet` | Runs `go vet ./...` static analysis |

---

## 🏛️ Technical Architecture

```
                    ┌──────────────────────────────┐
                    │ System Audio / Media Players │
                    └──────────────┬───────────────┘
                                   │
                                   ▼
                   ┌──────────────────────────────┐
                   │    Multiout Virtual Sink     │
                   └──────────────┬───────────────┘
                                  │ (Capture Callback)
                                  ▼
                   ┌──────────────────────────────┐
                   │   Fan-Out Broadcaster        │
                   └──────┬────────────────┬──────┘
                          │                │
            ┌─────────────┴─────┐    ┌─────┴─────────────┐
            ▼                   ▼    ▼                   ▼
     ┌──────────────┐                 ┌──────────────┐
     │ Ring Buffer  │                 │ Ring Buffer  │
     │  (Device A)  │                 │  (Device B)  │
     └──────┬───────┘                 └──────┬───────┘
            │                                │
     ┌──────┴───────┐                 ┌──────┴───────┐
     │ EMA Monitor  │                 │ EMA Monitor  │
     └──────┬───────┘                 └──────┬───────┘
            │ (Fill level)                   │ (Fill level)
            ▼                                ▼
     ┌──────────────┐                 ┌──────────────┐
     │ PI Controller│                 │ PI Controller│
     └──────┬───────┘                 └──────┬───────┘
            │ (Resample ratio)               │ (Resample ratio)
            ▼                                ▼
     ┌──────────────┐                 ┌──────────────┐
     │  Resampler   │                 │  Resampler   │
     └──────┬───────┘                 └──────┬───────┘
            │                                │
     ┌──────┴───────┐                 ┌──────┴───────┐
     │ Gain (dB)    │                 │ Gain (dB)    │
     └──────┬───────┘                 └──────┬───────┘
            │                                │
     ┌──────┴───────┐                 ┌──────┴───────┐
     │ Delay Buffer │                 │ Delay Buffer │
     └──────┬───────┘                 └──────┬───────┘
            │                                │
            ▼                                ▼
     ┌──────────────┐                 ┌──────────────┐
     │ DAC Output A │                 │ DAC Output B │
     │ (3.5 mm Jack)│                 │ (Bluetooth)  │
     └──────────────┘                 └──────────────┘
```

### Architectural Pipeline
1. **Audio Capture:** Captures audio from the system monitor or virtual sink at $48\text{ kHz}$ stereo 32-bit float.
2. **Lock-Free Atomic Ring Buffer:** An allocation-free multi-channel ring buffer decouples the input capture thread from individual output audio callbacks.
3. **EMA Filter & PI Controller:** Buffer fill level is measured every callback and smoothed via an exponential moving average. The proportional-integral controller adjusts the target resampler speed by $\pm 5000\text{ ppm}$ to eliminate drift.
4. **Fractional Linear Resampler:** Computes sub-sample interpolated audio frames while maintaining boundary history between callback chunks.
5. **Logarithmic Gain Controller:** Applies volume scaling using $10^{\text{dB} / 20}$ with a 10 ms linear anti-pop ramp whenever sliders move.
6. **Delay Buffer:** A circular sample delay line delays output by $0\text{--}300\text{ ms}$ for phase alignment.

---

## 📖 User Guide

1. **Launch Multiout:** Start the application via your desktop launcher or terminal (`multiout`).
2. **Select Input Source:**
   - **Linux:** Multiout automatically provisions and selects the **Multiout Virtual Sink**. All desktop audio routes through it seamlessly.
   - **Windows:** Select your **Virtual Cable** (e.g. `CABLE Output (VB-Audio Virtual Cable)`) or a secondary output's loopback (`System Audio (Default Output)` or monitor loopback).
3. **Enable Output Devices:**
   - Toggle the switch for each speaker, headphone, or Bluetooth headset you want to output audio to.
   - *Note on Windows:* If you capture an active output device's loopback, routing audio back into that same physical device is automatically blocked or alerted to prevent acoustic feedback loops.
4. **Calibrate Volume & Delay:**
   - Adjust the **Volume** slider (calibrated in **dB**) for each device to find comfortable listening levels.
   - If one device is delayed (such as Bluetooth), increase the **Delay** slider (in **ms**) on the faster device until the echo disappears.
5. **Start Playback:**
   - Toggle the **start / stop** switch in the top header.
   - The status badge will change to `ACTIVE`, and the real-time peak meters will indicate live signal levels.
6. **Switch Themes:**
   - Click the theme toggle button in the header (`🌙 Dark` / `☀️ Light` / `🖥️ System`) to adapt to your desktop preference.

---

## ❓ Troubleshooting & FAQ

#### Q: On Windows, I hear a duplicate voice / echo on one of my headphones.
- This occurs when Windows' default playback device is set directly to one of your physical headphones. Windows sends raw audio straight to that headphone's hardware while Multiout also sends resampled audio to it.
- **Solution:** Set your Windows Default Playback Device to a Virtual Audio Cable (like **VB-Cable**) or a secondary monitor. Then in Multiout, choose that input source and toggle ON your headphones.

#### Q: I don't hear any sound after starting Multiout.
- Verify that your system's default sound output device is set to **Multiout Virtual Sink** (Linux) or your **Virtual Audio Cable / Monitor** (Windows) in your OS sound settings. Click the `⚙️ Sound Settings` button in Multiout to open your system settings directly.
- Ensure at least one output device in Multiout has its toggle switch enabled and its volume is above $-60\text{ dB}$.

#### Q: I hear a slight echo between my Bluetooth headset and wired earphones.
- Bluetooth headphones introduce intrinsic latency due to wireless encoding. Slide the **Delay** slider on your **wired earphones** up (typically $100\text{--}180\text{ ms}$) until the sound from both aligns simultaneously.

#### Q: I plugged in a new headset while Multiout was running.
- Click the **🔄 Refresh** button in the top header to discover newly connected playback and capture devices instantly without restarting.

#### Q: How do I inspect diagnostics or underruns?
- The built-in **Execution logs** terminal at the bottom of the window displays real-time connection status, device enumeration, and buffer underrun warnings. Click **Clear** to reset the view.

---

## 🗺️ Roadmap

- [x] Multi-device simultaneous audio output
- [x] Closed-loop PI drift compensation and fractional resampling
- [x] Logarithmic decibel volume control ($-60.0\text{ dB}$ to $+12.0\text{ dB}$)
- [x] Per-device delay synchronization ($0\text{--}300\text{ ms}$)
- [x] Modern Fyne UI with Dark / Light / System theme support
- [x] Linux `.deb` package and Windows `.exe` standalone compilation
- [ ] System tray / background minimization
- [ ] Per-application audio stream routing
- [ ] macOS CoreAudio Aggregate Device / Virtual Audio Driver integration

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

Developed with ❤️ by [Sagar](https://github.com/codesbysagar).
