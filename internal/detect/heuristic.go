package detect

import (
	"strings"
	"unicode"
)

// Classification is the result of running Classify on an input line.
type Classification int

const (
	// PassThrough means the shell should produce its usual "command not found"
	// for the line; we do not believe it was natural language.
	PassThrough Classification = iota
	// Route means the line looks like natural language and should be handed to
	// the configured harness.
	Route
)

// ForcedMode captures a per-prompt mode override parsed from the line's
// leading prefix (see spec §13 1a).
type ForcedMode string

const (
	ForceNone        ForcedMode = ""
	ForceInteractive ForcedMode = "interactive"
	ForceHeadless    ForcedMode = "headless"
)

// Result is the richer output of Parse — it reports the classification, any
// forced mode implied by a prefix, and the line stripped of its prefix so
// callers have a clean prompt to forward to the harness.
type Result struct {
	Class  Classification
	Forced ForcedMode
	Line   string
}

// Options tweaks Parse. Currently just user-defined passthrough tokens.
type Options struct {
	// Passthrough first tokens — never route even if NL-shaped.
	// Matched case-sensitively against the first whitespace-delimited token.
	Passthrough []string
}

// Parse runs prefix handling, user-configured passthrough, then Classify.
// Prefix precedence, checked longest-first:
//
//	??    → Route, ForceHeadless, strip "??"
//	?!    → Route, ForceInteractive, strip "?!"
//	?     → Route, strip "?"
//	\     → Route, strip "\"
//	!     → PassThrough, no stripping
func Parse(line string, opts ...Options) Result {
	line = strings.TrimSpace(line)
	if line == "" {
		return Result{Class: PassThrough, Line: line}
	}
	switch {
	case strings.HasPrefix(line, "??"):
		return Result{Class: Route, Forced: ForceHeadless, Line: strings.TrimSpace(line[2:])}
	case strings.HasPrefix(line, "?!"):
		return Result{Class: Route, Forced: ForceInteractive, Line: strings.TrimSpace(line[2:])}
	case line[0] == '?':
		return Result{Class: Route, Line: strings.TrimSpace(line[1:])}
	case line[0] == '\\':
		return Result{Class: Route, Line: strings.TrimSpace(line[1:])}
	case line[0] == '!':
		return Result{Class: PassThrough, Line: line}
	}

	// User-configured passthrough first-tokens (spec §7 behavior.passthrough).
	if len(opts) > 0 && len(opts[0].Passthrough) > 0 {
		first := firstToken(line)
		for _, tok := range opts[0].Passthrough {
			if tok == first {
				return Result{Class: PassThrough, Line: line}
			}
		}
	}

	return Result{Class: Classify(line), Line: line}
}

func firstToken(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] == ' ' || line[i] == '\t' {
			return line[:i]
		}
	}
	return line
}

var stopwords = map[string]struct{}{
	"the": {}, "a": {}, "is": {}, "how": {}, "what": {}, "why": {},
	"can": {}, "i": {}, "my": {}, "to": {}, "do": {}, "does": {},
	"should": {},
}

var interrogatives = map[string]struct{}{
	"how": {}, "what": {}, "why": {}, "explain": {}, "write": {},
	"make": {}, "fix": {}, "help": {},
}

// Classify returns Route when line looks like natural language the user meant
// for their AI harness, or PassThrough otherwise. See spec §6.
func Classify(line string) Classification {
	line = strings.TrimSpace(line)
	if line == "" {
		return PassThrough
	}

	// Explicit escape hatches take precedence over everything else.
	switch line[0] {
	case '!':
		return PassThrough
	case '?', '\\':
		return Route
	}

	tokens := strings.Fields(line)

	// Hard gates.
	if len(tokens) < 3 {
		return PassThrough
	}
	if strings.ContainsAny(tokens[0], "/.-~$") {
		return PassThrough
	}
	if containsShellMetachar(line) {
		return PassThrough
	}

	// Soft signals — need at least two.
	signals := 0
	lower := strings.ToLower(line)
	lowerTokens := tokenizeWords(lower)
	if hasStopword(lowerTokens) {
		signals++
	}
	if strings.ContainsRune(line, '?') {
		signals++
	}
	if strings.ContainsAny(line, "',") {
		signals++
	}
	if len(tokens) >= 6 {
		signals++
	}
	if _, ok := interrogatives[strings.ToLower(tokens[0])]; ok {
		signals++
	}

	if signals >= 2 {
		return Route
	}
	return PassThrough
}

// containsShellMetachar flags the unquoted shell metacharacters listed in
// spec §6. We do a straight-quote scan — good enough for v1.
func containsShellMetachar(line string) bool {
	inSingle := false
	inDouble := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch c {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
			continue
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
			continue
		}
		if inSingle || inDouble {
			continue
		}
		switch c {
		case '|', '>', '<', '&', ';', '`':
			return true
		case '$':
			if i+1 < len(line) && line[i+1] == '(' {
				return true
			}
		}
	}
	return false
}

// tokenizeWords splits on non-letter/digit runs so that "i" in "do i" and
// "difference" are both recognizable as whole words.
func tokenizeWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func hasStopword(tokens []string) bool {
	for _, t := range tokens {
		if _, ok := stopwords[t]; ok {
			return true
		}
	}
	return false
}
