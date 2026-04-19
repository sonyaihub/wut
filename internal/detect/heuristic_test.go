package detect

import "testing"

func TestClassifyContractionsRoute(t *testing.T) {
	// "whats going on" → 3 tokens, "whats" is an interrogative (apostrophe-dropped)
	// and len>=6 is still false → only 1 signal, still PassThrough.
	// "whats going on here now bro" → still 1 signal (interrogative) + len>=6 = 2.
	if got := Classify("whats going on here now bro"); got != Route {
		t.Fatalf("expected Route for contraction-interrogative long line, got %v", got)
	}
	// "hows the weather today" → stopword "the" + interrogative "hows" = 2.
	if got := Classify("hows the weather today"); got != Route {
		t.Fatalf("expected Route for 'hows the weather today', got %v", got)
	}
}

func TestClassifyExtraSets(t *testing.T) {
	// Without config — "yo bruh sup" has 0 signals → PassThrough.
	if got := Classify("yo bruh sup"); got != PassThrough {
		t.Fatalf("unconfigured slang should passthrough, got %v", got)
	}
	// With extras registered, "yo" is an interrogative and "bruh" is a stopword → 2 signals.
	opts := Options{
		ExtraStopwords:      []string{"bruh"},
		ExtraInterrogatives: []string{"yo"},
	}
	if got := Classify("yo bruh sup", opts); got != Route {
		t.Fatalf("configured slang should route, got %v", got)
	}
}

func TestParsePassthroughTokens(t *testing.T) {
	opts := Options{Passthrough: []string{"howto", "make"}}
	// Would normally route (interrogative "make" + stopwords), but config
	// says pass through.
	if got := Parse("make all the things for me please", opts); got.Class != PassThrough {
		t.Fatalf("expected passthrough for user-configured token, got %+v", got)
	}
	// Different first token still classifies normally.
	if got := Parse("how do I rebase onto main", opts); got.Class != Route {
		t.Fatalf("unrelated token should route, got %+v", got)
	}
}

func TestParsePrefixes(t *testing.T) {
	cases := []struct {
		name       string
		line       string
		wantClass  Classification
		wantForced ForcedMode
		wantLine   string
	}{
		{"double-q forces headless", "??how do I rebase onto main", Route, ForceHeadless, "how do I rebase onto main"},
		{"q-bang forces interactive", "?!fix this regex for me", Route, ForceInteractive, "fix this regex for me"},
		{"single q routes no force", "? short", Route, ForceNone, "short"},
		{"backslash routes no force", "\\ foo bar", Route, ForceNone, "foo bar"},
		{"bang passes through", "!how do I", PassThrough, ForceNone, "!how do I"},
		{"plain NL falls through to classifier", "how do I rebase onto main", Route, ForceNone, "how do I rebase onto main"},
		{"plain typo falls through to classifier", "gti status", PassThrough, ForceNone, "gti status"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.line)
			if got.Class != tc.wantClass || got.Forced != tc.wantForced || got.Line != tc.wantLine {
				t.Fatalf("Parse(%q) = %+v; want {Class:%v Forced:%q Line:%q}", tc.line, got, tc.wantClass, tc.wantForced, tc.wantLine)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		line string
		want Classification
	}{
		// Real commands — must pass through.
		{"ls flags", "ls -la", PassThrough},
		{"git status", "git status", PassThrough},
		{"cd tilde", "cd ~/tmp", PassThrough},
		{"relative script", "./script.sh", PassThrough},
		{"python -V", "python3 -V", PassThrough},

		// Typos — must pass through.
		{"single typo", "gti", PassThrough},
		{"single sl", "sl", PassThrough},
		{"two-word typo", "pythno script.py", PassThrough},
		{"git typo", "gti statsu", PassThrough},

		// Natural language — must route.
		{"rebase q", "how do I rebase onto main", Route},
		{"reset vs revert", "what is the difference between git reset and git revert", Route},
		{"regex explain", "explain what this regex does in plain english", Route},

		// Escape hatch.
		{"qmark prefix", "? one", Route},
		{"backslash prefix", "\\ foo bar", Route},

		// Passthrough prefix beats heuristic.
		{"bang prefix", "!how do I rebase onto main", PassThrough},

		// Shell metachars anywhere → pass through.
		{"pipe in prose", "how do i grep | sort", PassThrough},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.line)
			if got != tc.want {
				t.Fatalf("Classify(%q) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
}
