// Package output provides terminal output helpers: colored text, spinners, tables.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// Level represents the severity/type of output.
type Level int

const (
	Debug Level = iota
	Info
	Success
	Warn
	Error
)

// Colors and styles.
const (
	reset   = "\x1b[0m"
	bold    = "\x1b[1m"
	dim     = "\x1b[2m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"
	gray    = "\x1b[90m"
)

var (
	// Stdout is the writer used by Printf and friends.
	Stdout io.Writer = os.Stdout
	// Stderr is the writer used for error output.
	Stderr io.Writer = os.Stderr
	// ColorEnabled controls whether ANSI codes are emitted.
	ColorEnabled = term.IsTerminal(int(os.Stdout.Fd()))
)

// levelStyles maps levels to their ANSI prefix.
var levelStyles = map[Level]struct{ color, label string }{
	Debug:   {gray, "    "},
	Info:    {cyan, "    "},
	Success: {green, "    "},
	Warn:    {yellow, "Warning"},
	Error:   {red, "Error"},
}

// Printf prints a formatted message with level-appropriate styling.
func Printf(lvl Level, format string, args ...any) {
	Fprintf(Stdout, lvl, format, args...)
}

// Fprintf writes a formatted message to w with level-appropriate styling.
func Fprintf(w io.Writer, lvl Level, format string, args ...any) {
	style, ok := levelStyles[lvl]
	if !ok {
		style = levelStyles[Info]
	}

	msg := fmt.Sprintf(format, args...)

	if ColorEnabled {
		if style.label != "    " {
			fmt.Fprintf(w, "%s%12s%s %s%s%s\n", bold+style.color, style.label, reset, style.color, msg, reset)
		} else {
			fmt.Fprintf(w, "%s%s%s\n", style.color, msg, reset)
		}
	} else {
		if style.label != "    " {
			fmt.Fprintf(w, "%12s %s\n", style.label, msg)
		} else {
			fmt.Fprintln(w, msg)
		}
	}
}

// Errorf prints an error and exits. For use in cmdimpl layer.
func Errorf(format string, args ...any) {
	Fprintf(Stderr, Error, format, args...)
}

// Warnf prints a warning.
func Warnf(format string, args ...any) {
	Fprintf(Stderr, Warn, format, args...)
}

// Infof prints an info message.
func Infof(format string, args ...any) {
	Fprintf(Stdout, Info, format, args...)
}

// Successf prints a success message.
func Successf(format string, args ...any) {
	Fprintf(Stdout, Success, format, args...)
}

// Debugf prints a debug message.
func Debugf(format string, args ...any) {
	Fprintf(Stdout, Debug, format, args...)
}

// Spinner is a terminal spinner for long-running operations.
type Spinner struct {
	mu      sync.Mutex
	writer  io.Writer
	message string
	frames  []string
	stopCh  chan struct{}
	doneCh  chan struct{}
	active  bool
}

// SpinnerOption configures a Spinner.
type SpinnerOption func(*Spinner)

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string, opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		writer:  Stdout,
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		return
	}
	s.active = true
	// Recreate channels only if they were closed by a previous Stop
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		frameIdx := 0

		for {
			select {
			case <-s.stopCh:
				s.clear()
				close(s.doneCh)
				return
			case <-ticker.C:
				s.render(s.frames[frameIdx])
				frameIdx = (frameIdx + 1) % len(s.frames)
			}
		}
	}()
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	ch := s.stopCh
	s.mu.Unlock()
	close(ch)
	<-s.doneCh
}

// Message updates the spinner message.
func (s *Spinner) Message(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = msg
}

func (s *Spinner) render(frame string) {
	if !ColorEnabled {
		return
	}
	s.mu.Lock()
	msg := s.message
	s.mu.Unlock()
	fmt.Fprintf(s.writer, "\r%s %s%s %s", cyan+frame, reset, msg, reset)
}

func (s *Spinner) clear() {
	if !ColorEnabled {
		return
	}
	// Clear the line
	fmt.Fprint(s.writer, "\r\x1b[K")
}

// Step prints a single progress step (non-animated).
func Step(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled {
		fmt.Fprintf(Stdout, "   %s●%s %s\n", dim, reset, msg)
	} else {
		fmt.Fprintf(Stdout, "   • %s\n", msg)
	}
}

// StepDone marks a step as completed.
func StepDone(message string) {
	fmt.Fprintf(Stdout, "\r\x1b[K")
	Printf(Success, "%s %s", green+"✓"+reset, message)
}

// StepFail marks a step as failed.
func StepFail(message string) {
	fmt.Fprintf(Stdout, "\r\x1b[K")
	Printf(Error, "%s %s", red+"✗"+reset, message)
}

// Header prints a section header.
func Header(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled {
		fmt.Fprintf(Stdout, "  %s%s%s\n", bold+blue, msg, reset)
	} else {
		fmt.Fprintf(Stdout, "  %s\n", msg)
	}
}

// Banner prints a fancy banner.
func Banner(name string) {
	if !ColorEnabled {
		fmt.Println(name)
		return
	}
	width := len(name) + 8
	// Top border
	fmt.Print(cyan)
	for i := 0; i < width; i++ {
		fmt.Print("━")
	}
	fmt.Println(reset)
	// Name
	fmt.Printf("%s  %s%s%s%s  %s\n", cyan, bold+white, name, reset, cyan, reset)
	// Bottom border
	fmt.Print(cyan)
	for i := 0; i < width; i++ {
		fmt.Print("━")
	}
	fmt.Println(reset)
}

// HRule prints a horizontal rule.
func HRule() {
	if ColorEnabled {
		fmt.Fprintf(Stdout, "%s%s%s\n", dim, "──────────────", reset)
	} else {
		fmt.Println("──────────────")
	}
}

// Table renders a formatted table.
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table.
func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

// Row adds a row to the table.
func (t *Table) Row(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Print renders the table to stdout.
func (t *Table) Print() {
	t.Fprint(Stdout)
}

// Fprint renders the table to a writer.
func (t *Table) Fprint(w io.Writer) {
	if len(t.headers) == 0 && len(t.rows) == 0 {
		return
	}

	// Calculate column widths
	all := append([][]string{t.headers}, t.rows...)
	cols := len(t.headers)
	for _, r := range t.rows {
		if len(r) > cols {
			cols = len(r)
		}
	}

	widths := make([]int, cols)
	for _, row := range all {
		for i, cell := range row {
			w := utf8.RuneCountInString(cell)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	// Print header
	if len(t.headers) > 0 {
		for i, h := range t.headers {
			if ColorEnabled {
				fmt.Fprintf(w, "  %s%-*s%s", bold, widths[i]+2, h, reset)
			} else {
				fmt.Fprintf(w, "  %-*s", widths[i]+2, h)
			}
			if i < len(t.headers)-1 {
				fmt.Fprint(w, " ")
			}
		}
		fmt.Fprintln(w)
	}

	// Print rows
	for _, row := range t.rows {
		for i, cell := range row {
			pad := widths[i]
			if i < len(widths) {
				pad = widths[i]
			}
			if i == 0 && ColorEnabled {
				fmt.Fprintf(w, "  %s%-*s%s", dim, pad+2, cell, reset)
			} else {
				fmt.Fprintf(w, "  %-*s", pad+2, cell)
			}
			if i < len(row)-1 {
				fmt.Fprint(w, " ")
			}
		}
		fmt.Fprintln(w)
	}
}

// Style applies an ANSI style code to s when colors are enabled.
func Style(s, code string) string {
	if !ColorEnabled {
		return s
	}
	return code + s + reset
}

func Bold(s string) string   { return Style(s, bold) }
func Dim(s string) string    { return Style(s, dim) }
func Red(s string) string    { return Style(s, red) }
func Green(s string) string  { return Style(s, green) }
func Yellow(s string) string { return Style(s, yellow) }
func Cyan(s string) string   { return Style(s, cyan) }
func Blue(s string) string   { return Style(s, blue) }
func Gray(s string) string   { return Style(s, gray) }
