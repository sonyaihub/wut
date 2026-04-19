package render

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Spinner draws a single-line spinner to a writer and erases it on Stop.
// It is safe for concurrent use; Start and Stop may be called at most once.
type Spinner struct {
	W     io.Writer
	Label string
	Style string // "dots" | "line" | "pipe" | "none"

	once    sync.Once
	stopCh  chan struct{}
	doneCh  chan struct{}
	stopped bool
	mu      sync.Mutex
}

var spinnerFrames = map[string][]string{
	"dots": {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"line": {"-", "\\", "|", "/"},
	"pipe": {"|", "/", "-", "\\"},
}

func (s *Spinner) frames() []string {
	if f, ok := spinnerFrames[s.Style]; ok {
		return f
	}
	return spinnerFrames["dots"]
}

// Start begins drawing the spinner in a background goroutine. If Style is
// "none", Start is a no-op — Stop is still safe to call.
func (s *Spinner) Start() {
	if s.Style == "none" {
		return
	}
	s.once.Do(func() {
		s.stopCh = make(chan struct{})
		s.doneCh = make(chan struct{})
		go s.run()
	})
}

func (s *Spinner) run() {
	defer close(s.doneCh)
	frames := s.frames()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	s.draw(frames[i])
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			i = (i + 1) % len(frames)
			s.draw(frames[i])
		}
	}
}

func (s *Spinner) draw(frame string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	// \r returns cursor to column 0; \x1b[K clears to end of line.
	fmt.Fprintf(s.W, "\r\x1b[K%s %s", frame, s.Label)
}

// Stop halts the spinner and clears its line. Safe to call multiple times.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	s.mu.Unlock()

	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
	}
	// Erase whatever the last frame drew.
	fmt.Fprint(s.W, "\r\x1b[K")
}
