//go:build linux

package audio

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	VirtualSinkName        = "multiout"
	VirtualSinkDescription = "Multiout Virtual Sink"
)

// LinuxVirtualSinkManager implements VirtualSinkManager on Linux using
// PipeWire native tools (pw-cli / wpctl) with fallback to pactl.
type LinuxVirtualSinkManager struct {
	mu           sync.Mutex
	active       bool
	origSinkID   int
	origSinkName string
	sinkNodeID   int
	usePulse     bool // true if using pactl instead of PipeWire
}

// NewVirtualSinkManager creates a platform-appropriate VirtualSinkManager.
func NewVirtualSinkManager() VirtualSinkManager {
	_, err := exec.LookPath("wpctl")
	usePulse := err != nil

	return &LinuxVirtualSinkManager{
		usePulse: usePulse,
	}
}

func (m *LinuxVirtualSinkManager) VirtualSinkName() string {
	return VirtualSinkName
}

func (m *LinuxVirtualSinkManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// CleanupStale searches for and destroys any leftover multiout sink from a previous crash.
func (m *LinuxVirtualSinkManager) CleanupStale() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.usePulse {
		// PipeWire cleanup: find node ID of multiout
		id, err := findPipeWireNodeID(VirtualSinkName)
		if err == nil && id > 0 {
			_ = exec.Command("pw-cli", "destroy", strconv.Itoa(id)).Run()
		}
	} else {
		// PulseAudio cleanup: search pactl modules
		out, err := exec.Command("pactl", "list", "short", "modules").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, VirtualSinkName) {
					fields := strings.Fields(line)
					if len(fields) > 0 {
						_ = exec.Command("pactl", "unload-module", fields[0]).Run()
					}
				}
			}
		}
	}
	return nil
}

// Setup creates the multiout virtual sink, remembers the default sink, and sets multiout as default.
func (m *LinuxVirtualSinkManager) Setup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active {
		return nil
	}

	// 1. Cleanup any stale sink from past crashes
	if !m.usePulse {
		id, err := findPipeWireNodeID(VirtualSinkName)
		if err == nil && id > 0 {
			_ = exec.Command("pw-cli", "destroy", strconv.Itoa(id)).Run()
		}
	}

	// 2. Remember original default sink
	if !m.usePulse {
		origID, err := getPipeWireDefaultSinkID()
		if err != nil {
			return fmt.Errorf("detect default sink via wpctl: %w", err)
		}
		m.origSinkID = origID

		// 3. Create PipeWire null sink
		cmd := exec.Command("pw-cli", "create-node", "adapter", fmt.Sprintf(
			"{ factory.name=support.null-audio-sink node.name=%s node.description=\"%s\" media.class=Audio/Sink audio.position=[FL,FR] object.linger=true }",
			VirtualSinkName, VirtualSinkDescription,
		))
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("create pipewire null sink: %s (%w)", string(out), err)
		}

		// 4. Find created node ID
		newNodeID, err := findPipeWireNodeID(VirtualSinkName)
		if err != nil {
			return fmt.Errorf("locate created multiout sink node: %w", err)
		}
		m.sinkNodeID = newNodeID

		// 5. Set multiout as default sink
		if err := exec.Command("wpctl", "set-default", strconv.Itoa(newNodeID)).Run(); err != nil {
			// Rollback
			_ = exec.Command("pw-cli", "destroy", strconv.Itoa(newNodeID)).Run()
			return fmt.Errorf("set multiout as default sink: %w", err)
		}
	} else {
		// PulseAudio flow
		origNameOut, err := exec.Command("pactl", "get-default-sink").Output()
		if err != nil {
			return fmt.Errorf("get default sink via pactl: %w", err)
		}
		m.origSinkName = strings.TrimSpace(string(origNameOut))

		// Load module-null-sink
		cmd := exec.Command("pactl", "load-module", "module-null-sink",
			"sink_name="+VirtualSinkName,
			fmt.Sprintf("sink_properties=device.description=\"%s\"", VirtualSinkDescription),
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("load module-null-sink: %s (%w)", string(out), err)
		}
		modID, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		m.sinkNodeID = modID

		// Set default
		if err := exec.Command("pactl", "set-default-sink", VirtualSinkName).Run(); err != nil {
			_ = exec.Command("pactl", "unload-module", strconv.Itoa(modID)).Run()
			return fmt.Errorf("set-default-sink multiout: %w", err)
		}
	}

	m.active = true
	return nil
}

// Restore restores the original default sink and destroys the virtual sink.
func (m *LinuxVirtualSinkManager) Restore() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return nil
	}

	var restoreErr error
	if !m.usePulse {
		// Restore default sink
		if m.origSinkID > 0 {
			if err := exec.Command("wpctl", "set-default", strconv.Itoa(m.origSinkID)).Run(); err != nil {
				restoreErr = fmt.Errorf("restore original default sink %d: %w", m.origSinkID, err)
			}
		}

		// Destroy virtual sink node
		if m.sinkNodeID > 0 {
			_ = exec.Command("pw-cli", "destroy", strconv.Itoa(m.sinkNodeID)).Run()
		}
	} else {
		if m.origSinkName != "" {
			_ = exec.Command("pactl", "set-default-sink", m.origSinkName).Run()
		}
		if m.sinkNodeID > 0 {
			_ = exec.Command("pactl", "unload-module", strconv.Itoa(m.sinkNodeID)).Run()
		}
	}

	m.active = false
	m.sinkNodeID = 0
	return restoreErr
}

// Helper: parse wpctl status to find current default sink ID
func getPipeWireDefaultSinkID() (int, error) {
	out, err := exec.Command("wpctl", "status").Output()
	if err != nil {
		return 0, err
	}
	return parseDefaultSinkFromWpctl(string(out))
}

var sinkRegex = regexp.MustCompile(`\*\s+(\d+)\.`)

func parseDefaultSinkFromWpctl(statusOutput string) (int, error) {
	inSinks := false
	for _, line := range strings.Split(statusOutput, "\n") {
		if strings.Contains(line, "Sinks:") {
			inSinks = true
			continue
		}
		if inSinks {
			if strings.Contains(line, "Sources:") || strings.Contains(line, "Filters:") {
				break
			}
			if strings.Contains(line, "*") {
				matches := sinkRegex.FindStringSubmatch(line)
				if len(matches) >= 2 {
					return strconv.Atoi(matches[1])
				}
			}
		}
	}
	return 0, fmt.Errorf("no default sink found in wpctl status")
}

// Helper: find PipeWire node ID for a given node.name
func findPipeWireNodeID(nodeName string) (int, error) {
	out, err := exec.Command("pw-cli", "list-objects", "Node").Output()
	if err != nil {
		return 0, err
	}
	return parseNodeIDFromPwCli(string(out), nodeName)
}

func parseNodeIDFromPwCli(output, nodeName string) (int, error) {
	nodeIDRegex := regexp.MustCompile(`id\s+(\d+),\s+type\s+PipeWire:Interface:Node`)
	currentID := 0

	for _, line := range strings.Split(output, "\n") {
		if m := nodeIDRegex.FindStringSubmatch(line); len(m) >= 2 {
			currentID, _ = strconv.Atoi(m[1])
			continue
		}
		if currentID > 0 && strings.Contains(line, "node.name") && strings.Contains(line, nodeName) {
			return currentID, nil
		}
	}
	return 0, fmt.Errorf("node %q not found", nodeName)
}
