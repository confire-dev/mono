package optimizer

import (
	"regexp"
	"strings"
)

// WebFetchOptimizer strips HTML/CSS/JS boilerplate from web page responses.
// Infrastructure optimizer — runs locally as part of the free tier.
const webfetchMaxBytes = 30_000

var (
	wfScriptRe  = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	wfStyleRe   = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	wfSVGRe     = regexp.MustCompile(`(?is)<svg\b[^>]*>.*?</svg>`)
	// RE2 has no backreferences — use alternation in the closing tag instead.
	// Slightly over-matches across mismatched pairs, fine for noise stripping.
	wfNavRe = regexp.MustCompile(`(?is)<(?:nav|header|footer|aside|menu)\b[^>]*>.*?</(?:nav|header|footer|aside|menu)>`)
	wfCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)
	wfAttrRe    = regexp.MustCompile(`(?i)\s+(class|style|id|data-[a-z-]+|aria-[a-z-]+|on\w+)="[^"]*"`)
	wfTagRe     = regexp.MustCompile(`<[^>]+>`)
	wfMultiSPRe = regexp.MustCompile(`[ \t]{2,}`)
	wfMultiNLRe = regexp.MustCompile(`\n{3,}`)
)

type WebFetchOptimizer struct{}

func (w *WebFetchOptimizer) Matches(toolName string) bool {
	return toolName == "webfetch"
}

func (w *WebFetchOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	text, ok := data.(string)
	if !ok {
		return data
	}

	if len(text) < 2_000 {
		return data
	}
	if !strings.Contains(text, "<") || !strings.Contains(text, ">") {
		return data
	}

	result := text
	result = wfScriptRe.ReplaceAllString(result, "")
	result = wfStyleRe.ReplaceAllString(result, "")
	result = wfSVGRe.ReplaceAllString(result, "[svg]")
	result = wfNavRe.ReplaceAllString(result, "")
	result = wfCommentRe.ReplaceAllString(result, "")
	result = wfAttrRe.ReplaceAllString(result, "")
	result = wfTagRe.ReplaceAllString(result, " ")
	result = wfMultiSPRe.ReplaceAllString(result, " ")
	result = wfMultiNLRe.ReplaceAllString(result, "\n\n")
	result = strings.TrimSpace(result)

	if len(result) > webfetchMaxBytes {
		result = result[:webfetchMaxBytes] + "\n[confire: truncated]"
	}

	if len(result) >= int(float64(len(text))*0.7) {
		return data
	}
	return result
}
