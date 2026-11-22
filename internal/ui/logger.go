package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ExecutionLogger manages an in-memory ring buffer of timestamped log messages.
type ExecutionLogger struct {
	mu             sync.Mutex
	lines          []string
	maxLines       int
	listeners      []func(line string)
	clearListeners []func()
}

// NewExecutionLogger creates a thread-safe execution logger.
func NewExecutionLogger(maxLines int) *ExecutionLogger {
	if maxLines <= 0 {
		maxLines = 300
	}
	return &ExecutionLogger{
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

// Log appends a new formatted log line with a high-precision timestamp.
func (l *ExecutionLogger) Log(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s", timestamp, msg)

	l.mu.Lock()
	if len(l.lines) >= l.maxLines {
		l.lines = l.lines[1:]
	}
	l.lines = append(l.lines, line)
	listeners := make([]func(string), len(l.listeners))
	copy(listeners, l.listeners)
	l.mu.Unlock()

	if len(listeners) == 0 {
		return
	}
	dispatch := func() {
		for _, listener := range listeners {
			listener(line)
		}
	}
	if fyne.CurrentApp() != nil {
		fyne.Do(dispatch)
	} else {
		dispatch()
	}
}

// Clear clears all stored logs and informs listeners.
func (l *ExecutionLogger) Clear() {
	l.mu.Lock()
	l.lines = nil
	clearListeners := make([]func(), len(l.clearListeners))
	copy(clearListeners, l.clearListeners)
	l.mu.Unlock()

	dispatch := func() {
		for _, fn := range clearListeners {
			fn()
		}
	}
	if fyne.CurrentApp() != nil {
		fyne.Do(dispatch)
	} else {
		dispatch()
	}
}

// AddListener registers a callback that runs on the UI goroutine for every new log.
func (l *ExecutionLogger) AddListener(fn func(string)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.listeners = append(l.listeners, fn)
}

// AddClearListener registers a callback for clear events.
func (l *ExecutionLogger) AddClearListener(fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clearListeners = append(l.clearListeners, fn)
}

// AllLines returns a copy of all current lines in chronological order.
func (l *ExecutionLogger) AllLines() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.lines))
	copy(out, l.lines)
	return out
}

// LogConsole is a UI component rendering the streaming execution log panel.
type LogConsole struct {
	container *fyne.Container
	borderBg  *canvas.Rectangle
	entry     *widget.Entry
	vscroll   *container.Scroll
}

// NewLogConsole creates a terminal-style log box with theme-aware borders.
func NewLogConsole(logger *ExecutionLogger) *LogConsole {
	entry := widget.NewMultiLineEntry()
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	entry.Wrapping = fyne.TextWrapBreak
	entry.Disable() // Read-only
	entry.SetMinRowsVisible(6)

	// Pre-fill existing logs
	lines := logger.AllLines()
	if len(lines) > 0 {
		entry.SetText(strings.Join(lines, "\n") + "\n")
	}

	vscroll := container.NewVScroll(entry)

	logger.AddListener(func(newLine string) {
		if entry.Text == "" {
			entry.SetText(newLine + "\n")
		} else {
			entry.SetText(entry.Text + newLine + "\n")
		}
		entry.CursorRow = strings.Count(entry.Text, "\n")
		vscroll.ScrollToBottom()
	})

	logger.AddClearListener(func() {
		entry.SetText("")
	})

	p := GetPalette()
	borderBg := canvas.NewRectangle(p.WellBg)
	borderBg.CornerRadius = 8
	borderBg.StrokeColor = p.Border
	borderBg.StrokeWidth = 1

	boxed := container.NewStack(borderBg, container.NewPadded(vscroll))

	return &LogConsole{
		container: boxed,
		borderBg:  borderBg,
		entry:     entry,
		vscroll:   vscroll,
	}
}

// UpdateTheme refreshes the console border and background.
func (lc *LogConsole) UpdateTheme() {
	p := GetPalette()
	lc.borderBg.FillColor = p.WellBg
	lc.borderBg.StrokeColor = p.Border
	lc.borderBg.Refresh()
}

// Container returns the Fyne CanvasObject for layout placement.
func (lc *LogConsole) Container() fyne.CanvasObject {
	return lc.container
}
