package optimizer

import (
	"regexp"
	"strings"
)

// WebFetchOptimizer strips HTML/CSS/JS boilerplate from web page responses.
// Infrastructure optimizer — runs locally as part of the free tier.
const webfetchMaxBytes = 30_000

var (
	wfHeadRe    = regexp.MustCompile(`(?is)<head\b[^>]*>.*?</head>`)
	wfScriptRe  = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	wfStyleRe   = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	wfSVGRe     = regexp.MustCompile(`(?is)<svg\b[^>]*>.*?</svg>`)
	wfFigureRe  = regexp.MustCompile(`(?is)<(?:figure|picture)\b[^>]*>.*?</(?:figure|picture)>`)
	wfImgRe     = regexp.MustCompile(`(?i)<img\b[^>]*>`)
	// RE2 has no backreferences — use alternation in the closing tag instead.
	// Slightly over-matches across mismatched pairs, fine for noise stripping.
	wfNavRe     = regexp.MustCompile(`(?is)<(?:nav|header|footer|aside|menu)\b[^>]*>.*?</(?:nav|header|footer|aside|menu)>`)
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

	// Claude Code sends {"bytes":N,"code":200,"codeText":"OK","result":"...markdown...","durationMs":N,"url":"..."}
	// result is already processed markdown — just cap length if needed.
	if m, ok := data.(map[string]interface{}); ok {
		if result, ok := m["result"].(string); ok {
			optimized := w.optimizeMarkdown(result)
			if optimized == result {
				return data
			}
			newMap := make(map[string]interface{}, len(m))
			for k, v := range m {
				newMap[k] = v
			}
			newMap["result"] = optimized
			return newMap
		}
	}

	// Fallback: raw HTML string (tests, direct calls, older Claude Code versions)
	text, ok := data.(string)
	if !ok {
		return data
	}
	return w.optimizeHTML(text)
}

func (w *WebFetchOptimizer) optimizeMarkdown(text string) string {
	if len(text) <= webfetchMaxBytes {
		return text
	}
	return text[:webfetchMaxBytes] + "\n[confire: truncated]"
}

func (w *WebFetchOptimizer) optimizeHTML(text string) string {
	if len(text) < 2_000 {
		return text
	}
	if !strings.Contains(text, "<") || !strings.Contains(text, ">") {
		return text
	}

	result := text
	result = wfHeadRe.ReplaceAllString(result, "")
	result = wfScriptRe.ReplaceAllString(result, "")
	result = wfStyleRe.ReplaceAllString(result, "")
	result = wfSVGRe.ReplaceAllString(result, "[svg]")
	result = wfFigureRe.ReplaceAllString(result, "")
	result = wfImgRe.ReplaceAllString(result, "[image]")
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
		return text
	}
	return result
}
