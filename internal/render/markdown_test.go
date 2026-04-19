package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkdownHeading(t *testing.T) {
	var buf bytes.Buffer
	m := &Markdown{W: &buf}
	m.Write([]byte("# Hello\nbody\n"))
	m.Close()
	out := buf.String()
	if !strings.Contains(out, ansiBold) {
		t.Fatalf("no bold ANSI: %q", out)
	}
	if !strings.Contains(out, "Hello") {
		t.Fatalf("heading body missing: %q", out)
	}
	if !strings.Contains(out, "body") {
		t.Fatalf("paragraph missing: %q", out)
	}
}

func TestMarkdownInline(t *testing.T) {
	var buf bytes.Buffer
	m := &Markdown{W: &buf}
	m.Write([]byte("this is **strong** and `code` together"))
	m.Close()
	out := buf.String()
	if !strings.Contains(out, ansiBold+"strong"+ansiReset) {
		t.Fatalf("bold not applied: %q", out)
	}
	if !strings.Contains(out, ansiInverse+" code "+ansiReset) {
		t.Fatalf("inline code not applied: %q", out)
	}
}

func TestMarkdownListBullets(t *testing.T) {
	var buf bytes.Buffer
	m := &Markdown{W: &buf}
	m.Write([]byte("- one\n- two\n"))
	m.Close()
	out := buf.String()
	if strings.Count(out, "• ") != 2 {
		t.Fatalf("want 2 bullets, got: %q", out)
	}
}

func TestMarkdownFencedCode(t *testing.T) {
	var buf bytes.Buffer
	m := &Markdown{W: &buf}
	m.Write([]byte("```go\nfmt.Println(\"hi\")\n```\n"))
	m.Close()
	out := buf.String()
	if !strings.Contains(out, ansiDim) {
		t.Fatalf("code not dimmed: %q", out)
	}
	if !strings.Contains(out, "fmt.Println") {
		t.Fatalf("code body missing: %q", out)
	}
}
