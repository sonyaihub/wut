package ui

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ErrNotTTY is returned when a Select is invoked without a terminal.
var ErrNotTTY = errors.New("ui: stdin is not a tty")

// ErrCancelled is returned when the user presses Ctrl-C or Escape.
var ErrCancelled = errors.New("ui: user cancelled")

// Option is one row in a Select prompt.
type Option struct {
	Label string
	Hint  string // shown in dim text after the label
}

// Select shows a labelled list of options, lets the user move with ↑/↓, and
// returns the chosen index.
//
// The prompt draws to stderr (so it doesn't contaminate stdout captures) and
// reads from /dev/tty directly when available — that keeps it working even
// when stdout is piped.
func Select(title string, opts []Option, initial int) (int, error) {
	if len(opts) == 0 {
		return 0, errors.New("ui.Select: no options")
	}

	tty, err := openTTY()
	if err != nil {
		return 0, err
	}
	defer tty.Close()

	fd := int(tty.Fd())
	if !term.IsTerminal(fd) {
		return 0, ErrNotTTY
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer term.Restore(fd, oldState)

	w := os.Stderr
	cursor := initial
	if cursor < 0 || cursor >= len(opts) {
		cursor = 0
	}

	hideCursor(w)
	defer showCursor(w)

	// Every complete render leaves the cursor one line below the last option,
	// so redraws start by moving up (len(opts)+1) lines to the title line,
	// then overwrite each line in turn. Clear-to-EOL on every line keeps
	// width-shrinking renders from leaving gunk behind.
	moveUp := len(opts) + 1

	redraw := func() {
		fmt.Fprintf(w, "\x1b[%dA\r", moveUp)
		fmt.Fprintf(w, "\x1b[2K%s\r\n", title)
		for i, o := range opts {
			marker := "  "
			label := o.Label
			if i == cursor {
				marker = "> "
				label = "\x1b[1m" + o.Label + "\x1b[0m"
			}
			suffix := ""
			if o.Hint != "" {
				suffix = dim(" — " + o.Hint)
			}
			fmt.Fprintf(w, "\x1b[2K%s%s%s\r\n", marker, label, suffix)
		}
	}

	// Reserve the block by emitting blank lines (\r\n, explicit CR because
	// terminal is in raw mode), then let redraw walk back up and fill them.
	for i := 0; i < moveUp; i++ {
		fmt.Fprint(w, "\r\n")
	}
	redraw()

	draw := redraw

	buf := make([]byte, 4)
	for {
		n, err := tty.Read(buf)
		if err != nil {
			return 0, err
		}
		seq := string(buf[:n])
		switch {
		case seq == "\x03", seq == "\x1b": // Ctrl-C, bare ESC
			return 0, ErrCancelled
		case seq == "\r", seq == "\n":
			return cursor, nil
		case seq == "\x1b[A", seq == "\x1b[D", seq == "k":
			if cursor > 0 {
				cursor--
				draw()
			}
		case seq == "\x1b[B", seq == "\x1b[C", seq == "j":
			if cursor < len(opts)-1 {
				cursor++
				draw()
			}
		}
	}
}

// openTTY opens /dev/tty so Select works even when stdout is piped.
func openTTY() (*os.File, error) {
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		return f, nil
	}
	// Fallback — only usable when the process does have a real terminal.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, nil
	}
	return nil, ErrNotTTY
}

func dim(s string) string   { return "\x1b[2m" + s + "\x1b[0m" }
func hideCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25l") }
func showCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25h") }
