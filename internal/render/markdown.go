package render

import (
	"bytes"
	"io"
	"regexp"
	"strings"
)

// Markdown buffers every byte until Close, then applies a lightweight
// markdown-to-ANSI pass before writing. Per spec §9, markdown is only
// applied in buffered mode; streaming invocations should get Raw instead.
type Markdown struct {
	W io.Writer

	buf    bytes.Buffer
	closed bool
}

func (m *Markdown) Write(p []byte) (int, error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.buf.Write(p)
}

func (m *Markdown) Close() error {
	if m.closed {
		return nil
	}
	m.closed = true
	out := renderMarkdown(m.buf.String())
	_, err := io.WriteString(m.W, out)
	return err
}

// ANSI SGR codes we emit.
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiDim       = "\x1b[2m"
	ansiItalic    = "\x1b[3m"
	ansiUnderline = "\x1b[4m"
	ansiInverse   = "\x1b[7m"
)

var (
	// **bold** and __bold__
	reBold = regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	// *italic* and _italic_ (single markers, not inside words)
	reItalic = regexp.MustCompile(`(^|\s)\*([^*\s][^*]*)\*|(^|\s)_([^_\s][^_]*)_`)
	// `inline code`
	reInlineCode = regexp.MustCompile("`([^`]+)`")
)

func renderMarkdown(src string) string {
	var out strings.Builder
	lines := strings.Split(src, "\n")
	inCodeFence := false

	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Fenced code blocks: toggle on ``` or ~~~.
		if strings.HasPrefix(strings.TrimSpace(trimmed), "```") || strings.HasPrefix(strings.TrimSpace(trimmed), "~~~") {
			inCodeFence = !inCodeFence
			out.WriteString(ansiDim)
			out.WriteString(trimmed)
			out.WriteString(ansiReset)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}
		if inCodeFence {
			out.WriteString(ansiDim)
			out.WriteString("  " + trimmed)
			out.WriteString(ansiReset)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}

		// Headings.
		if h := headingPrefix(trimmed); h > 0 {
			content := strings.TrimLeft(trimmed[h:], " ")
			out.WriteString(ansiBold)
			out.WriteString(ansiUnderline)
			out.WriteString(content)
			out.WriteString(ansiReset)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}

		// List bullets (-, *, +).
		processed := line
		if leading, rest, ok := splitListMarker(line); ok {
			processed = leading + "• " + applyInline(rest)
		} else {
			processed = applyInline(line)
		}
		out.WriteString(processed)
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// applyInline applies inline formatters in a safe order (inline code first so
// its contents aren't re-interpreted).
func applyInline(s string) string {
	s = reInlineCode.ReplaceAllString(s, ansiInverse+" $1 "+ansiReset)
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		inner := reBold.FindStringSubmatch(m)
		body := inner[1]
		if body == "" {
			body = inner[2]
		}
		return ansiBold + body + ansiReset
	})
	s = reItalic.ReplaceAllStringFunc(s, func(m string) string {
		inner := reItalic.FindStringSubmatch(m)
		// Two alternations, each with a leading-space capture and content capture.
		if inner[2] != "" {
			return inner[1] + ansiItalic + inner[2] + ansiReset
		}
		return inner[3] + ansiItalic + inner[4] + ansiReset
	})
	return s
}

// headingPrefix returns the length of a leading "# " heading marker, or 0.
func headingPrefix(line string) int {
	n := 0
	for n < len(line) && n < 6 && line[n] == '#' {
		n++
	}
	if n == 0 || n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n + 1
}

// splitListMarker returns (leading-whitespace, content) if line starts with a
// bullet marker, plus ok=true.
func splitListMarker(line string) (string, string, bool) {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i >= len(line) {
		return "", "", false
	}
	c := line[i]
	if c != '-' && c != '*' && c != '+' {
		return "", "", false
	}
	if i+1 >= len(line) || line[i+1] != ' ' {
		return "", "", false
	}
	return line[:i], line[i+2:], true
}
