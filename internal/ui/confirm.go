package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Confirm prints a Y/n prompt (default Yes) and returns true if the user
// accepts. Reads from /dev/tty so piped stdout doesn't break it.
// When no tty is available, returns (false, ErrNotTTY) so callers can decide
// whether to fail open or closed.
func Confirm(prompt string) (bool, error) {
	tty, err := openTTY()
	if err != nil {
		return false, err
	}
	defer tty.Close()
	fd := int(tty.Fd())
	if !term.IsTerminal(fd) {
		return false, ErrNotTTY
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return false, err
	}
	defer term.Restore(fd, oldState)

	fmt.Fprintf(os.Stderr, "%s [Y/n] ", prompt)
	defer fmt.Fprintln(os.Stderr)

	buf := make([]byte, 1)
	n, err := tty.Read(buf)
	if err != nil || n == 0 {
		return false, err
	}
	ch := strings.ToLower(string(buf[:n]))
	switch ch {
	case "\r", "\n", "y":
		return true, nil
	case "\x03", "\x1b": // Ctrl-C, Esc
		return false, ErrCancelled
	default:
		return false, nil
	}
}
