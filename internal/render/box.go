package render

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

// Writer is the minimum interface a renderer exposes: stream chunks of harness
// stdout through Write, then call Close once the harness has exited.
type Writer interface {
	io.WriteCloser
}

// Raw is a pass-through renderer: bytes go straight to the underlying writer.
type Raw struct{ W io.Writer }

func (r Raw) Write(p []byte) (int, error) { return r.W.Write(p) }
func (r Raw) Close() error                 { return nil }

// Box wraps streamed harness output in a Unicode box with the harness name as
// a header label. It assembles lines before drawing them — bytes can arrive
// in arbitrary chunks, so we buffer until we see a newline or Close.
type Box struct {
	W     io.Writer
	Label string
	Width int // usable width (including borders). Clamped to a safe minimum.

	headerDone bool
	buf        []byte
	closed     bool
}

const (
	boxTopLeft     = "╭"
	boxTopRight    = "╮"
	boxBottomLeft  = "╰"
	boxBottomRight = "╯"
	boxHorizontal  = "─"
	boxVertical    = "│"
	minBoxWidth    = 20
	defaultWidth   = 80
)

func (b *Box) ensureWidth() int {
	w := b.Width
	if w <= 0 {
		w = defaultWidth
	}
	if w < minBoxWidth {
		w = minBoxWidth
	}
	return w
}

func (b *Box) writeHeader() error {
	w := b.ensureWidth()
	// inner width is w minus the two corner runes ("╭" + "╮")
	inner := w - 2
	label := b.Label
	// pattern: ╭─ <label> <fill>─╮
	// "─ " (2) + label + " " (1) + fill + "─" (1) == inner
	prefix := boxHorizontal + " "
	suffixDash := boxHorizontal
	used := utf8.RuneCountInString(prefix) + utf8.RuneCountInString(label) + 1 + utf8.RuneCountInString(suffixDash)
	if used > inner {
		// Label too long; truncate label to fit.
		excess := used - inner
		lr := []rune(label)
		if excess >= len(lr) {
			label = ""
		} else {
			label = string(lr[:len(lr)-excess])
		}
		used = utf8.RuneCountInString(prefix) + utf8.RuneCountInString(label) + 1 + utf8.RuneCountInString(suffixDash)
	}
	fill := strings.Repeat(boxHorizontal, inner-used)
	line := boxTopLeft + prefix + label + " " + fill + suffixDash + boxTopRight + "\n"
	_, err := io.WriteString(b.W, line)
	return err
}

func (b *Box) writeFooter() error {
	w := b.ensureWidth()
	inner := w - 2
	line := boxBottomLeft + strings.Repeat(boxHorizontal, inner) + boxBottomRight + "\n"
	_, err := io.WriteString(b.W, line)
	return err
}

func (b *Box) drawLine(line string) error {
	w := b.ensureWidth()
	inner := w - 2 // inside the vertical bars
	// pad layout: "│ " + content + padding + " │"
	contentWidth := inner - 2 // subtract left/right spaces
	content := truncateOrPad(line, contentWidth)
	out := boxVertical + " " + content + " " + boxVertical + "\n"
	_, err := io.WriteString(b.W, out)
	return err
}

func (b *Box) Write(p []byte) (int, error) {
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	if !b.headerDone {
		if err := b.writeHeader(); err != nil {
			return 0, err
		}
		b.headerDone = true
	}
	b.buf = append(b.buf, p...)
	for {
		idx := bytes.IndexByte(b.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(b.buf[:idx])
		b.buf = b.buf[idx+1:]
		if err := b.drawLine(line); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (b *Box) Close() error {
	if b.closed {
		return nil
	}
	b.closed = true
	if !b.headerDone {
		if err := b.writeHeader(); err != nil {
			return err
		}
		b.headerDone = true
	}
	if len(b.buf) > 0 {
		if err := b.drawLine(string(b.buf)); err != nil {
			return err
		}
		b.buf = nil
	}
	return b.writeFooter()
}

// truncateOrPad returns s trimmed to width or right-padded with spaces.
// Width is counted in runes (not bytes) — good enough for ASCII + common
// latin; wide-char support is out of scope for M2.
func truncateOrPad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	n := utf8.RuneCountInString(s)
	if n == width {
		return s
	}
	if n > width {
		// drop the tail
		r := []rune(s)
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-n)
}

