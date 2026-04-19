package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestBoxWritesHeaderFooterAndLines(t *testing.T) {
	var buf bytes.Buffer
	b := &Box{W: &buf, Label: "claude", Width: 30}
	if _, err := b.Write([]byte("hello\nworld\n")); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "╭") {
		t.Fatalf("missing top-left corner: %q", out)
	}
	if !strings.Contains(out, "claude") {
		t.Fatalf("header missing label: %q", out)
	}
	if !strings.Contains(out, "│ hello") || !strings.Contains(out, "│ world") {
		t.Fatalf("missing body lines: %q", out)
	}
	if !strings.Contains(out, "╰") {
		t.Fatalf("missing bottom-left corner: %q", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 4 { // header + 2 body + footer
		t.Fatalf("want 4 lines, got %d: %q", len(lines), lines)
	}
}

func TestBoxBuffersPartialLines(t *testing.T) {
	var buf bytes.Buffer
	b := &Box{W: &buf, Label: "x", Width: 40}
	// write a partial line without trailing newline — should not draw yet
	b.Write([]byte("hel"))
	beforeNL := buf.String()
	if strings.Contains(beforeNL, "hel") {
		t.Fatalf("partial line drawn prematurely: %q", beforeNL)
	}
	b.Write([]byte("lo\n"))
	after := buf.String()
	if !strings.Contains(after, "│ hello") {
		t.Fatalf("expected joined line drawn: %q", after)
	}
}

func TestBoxFlushesTrailingOnClose(t *testing.T) {
	var buf bytes.Buffer
	b := &Box{W: &buf, Label: "x", Width: 40}
	b.Write([]byte("no trailing newline"))
	b.Close()
	if !strings.Contains(buf.String(), "│ no trailing newline") {
		t.Fatalf("trailing line not flushed on close: %q", buf.String())
	}
}
