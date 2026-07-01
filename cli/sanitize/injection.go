package sanitize

import (
	"regexp"
	"sort"
	"strings"
)

// InjectionMatch is one detected prompt-injection pattern.
type InjectionMatch struct {
	Pattern string
	Start   int
	End     int
}

// injectionNotice is prepended to sanitized output when suspicious instructions are found.
const injectionNotice = `CONFIRE NOTICE
Untrusted tool output contained instruction-like text. Confire treated it as data and removed/sandboxed suspicious instructions.

`

var injectionPatterns = []*regexp.Regexp{
	// Classic prompt injection phrases
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions?`),
	regexp.MustCompile(`(?i)disregard\s+(your\s+)?(system\s+prompt|previous\s+instructions?)`),
	regexp.MustCompile(`(?i)do\s+not\s+tell\s+the\s+user`),
	regexp.MustCompile(`(?i)from\s+now\s+on[,.]?\s+you\s+(are|must|should|will)`),
	regexp.MustCompile(`(?i)(exfiltrate|leak|steal)\s+.{0,60}(key|secret|token|credential|password)`),
	regexp.MustCompile(`(?i)send\s+(the\s+)?(contents?\s+of|everything\s+(in|from))\s+~?/`),
	regexp.MustCompile(`(?i)read\s+(the\s+)?(user'?s?\s+)?(secrets?|credentials?|\.env|private\s+key)`),
	regexp.MustCompile(`(?i)(reveal|output|print|show|return)\s+(your\s+)?(system\s+prompt|developer\s+message|hidden\s+instructions?)`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+(operating\s+in|in)\s+(developer|admin|root|unrestricted)\s+mode`),
	regexp.MustCompile(`(?i)<\s*hidden[_\s]?instruction\s*>`),
	regexp.MustCompile(`(?i)\[SYSTEM\]\s*:`),
	regexp.MustCompile(`(?i)\[INST\]\s*:`),

	// Hidden markup patterns in HTML/docs output
	regexp.MustCompile(`(?i)<!--.*?(instruct|ignore|prompt|system|exfiltrat).*?-->`),
	regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*display\s*:\s*none[^"']*["']`),
	regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*visibility\s*:\s*hidden[^"']*["']`),
	regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*font-size\s*:\s*0[^"']*["']`),
	regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*color\s*:\s*(?:white|#fff|#ffffff)[^"']*["']`),
}

// ScanInjection returns all detected prompt-injection matches in text.
func ScanInjection(text string) []InjectionMatch {
	var matches []InjectionMatch
	seen := make(map[string]bool)
	for _, re := range injectionPatterns {
		locs := re.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			val := text[loc[0]:loc[1]]
			if seen[val] {
				continue
			}
			seen[val] = true
			matches = append(matches, InjectionMatch{
				Pattern: re.String(),
				Start:   loc[0],
				End:     loc[1],
			})
		}
	}
	return matches
}

// Sanitize removes detected prompt-injection patterns from text.
// If anything was removed, it prepends the injectionNotice.
// Returns the sanitized text, whether anything was found, and the notice string used.
func Sanitize(text string) (sanitized string, found bool, notice string) {
	matches := mergeOverlapping(ScanInjection(text))
	if len(matches) == 0 {
		return text, false, ""
	}

	// Replace each match with a neutral placeholder, back to front to preserve
	// offsets. Safe because mergeOverlapping guarantees matches are sorted by
	// Start and non-overlapping: every later replacement happens at a position
	// >= the End of the match being replaced now, so result[:m.Start] and
	// result[m.End:] still point at the right content in the original text.
	result := text
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		result = result[:m.Start] + "[CONFIRE: suspicious instruction removed]" + result[m.End:]
	}

	// Collapse multiple adjacent placeholders.
	for strings.Contains(result, "[CONFIRE: suspicious instruction removed][CONFIRE: suspicious instruction removed]") {
		result = strings.ReplaceAll(result,
			"[CONFIRE: suspicious instruction removed][CONFIRE: suspicious instruction removed]",
			"[CONFIRE: suspicious instruction removed]")
	}

	return injectionNotice + result, true, injectionNotice
}

// mergeOverlapping sorts matches by Start position and merges any that
// overlap (or are contained within one another) into their union, so callers
// can safely apply replacements back-to-front without stale offsets.
// Multiple patterns can legitimately match overlapping spans — e.g. a
// standalone "ignore previous instructions" pattern and a broader
// "<!-- ...suspicious keyword... -->" comment pattern both matching inside
// the same HTML comment.
func mergeOverlapping(matches []InjectionMatch) []InjectionMatch {
	if len(matches) == 0 {
		return matches
	}
	sorted := append([]InjectionMatch(nil), matches...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		return sorted[i].End > sorted[j].End
	})

	merged := []InjectionMatch{sorted[0]}
	for _, m := range sorted[1:] {
		last := &merged[len(merged)-1]
		if m.Start < last.End {
			if m.End > last.End {
				last.End = m.End
			}
			continue
		}
		merged = append(merged, m)
	}
	return merged
}
