package intercept

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

// MCPSanitizeHandler runs PostToolUse security passes on MCP tool output:
// 1. Unicode hidden-text stripping (tag block, zero-width, BiDi overrides)
// 2. Secrets redaction using curated high-confidence patterns
// 3. Prompt-injection / cross-tool-steering detection
type MCPSanitizeHandler struct{}

func (h *MCPSanitizeHandler) ID() string      { return "mcp.sanitize" }
func (h *MCPSanitizeHandler) Phases() []Phase  { return []Phase{PhaseToolPost} }

func (h *MCPSanitizeHandler) Matches(e InterceptEvent) bool {
	return e.Tool != nil && e.Tool.IsMCP && e.Tool.Output != nil
}

func (h *MCPSanitizeHandler) Run(e InterceptEvent) (InterceptResult, error) {
	report := SanitizeReport{}

	cleaned, secretCount, secretTypes := redactSecrets(e.Tool.Output)
	report.SecretsRedacted = secretCount
	report.SecretTypes = secretTypes

	strippedUnicode := stripHiddenUnicode(cleaned)
	report.HiddenUnicodeFound = jsonSize(strippedUnicode) != jsonSize(cleaned)
	cleaned = strippedUnicode
	report.InjectionFound = detectInjection(cleaned)

	if !report.HasFindings() {
		return InterceptResult{Kind: ResultPassthrough}, nil
	}

	return InterceptResult{
		Kind:       ResultSanitize,
		ToolOutput: cleaned,
		Stats: &Stats{
			BeforeBytes: jsonSize(e.Tool.Output),
			AfterBytes:  jsonSize(cleaned),
			Optimizer:   "mcp.sanitize",
		},
	}, nil
}

// ── Unicode stripping ─────────────────────────────────────────────────────────

// stripHiddenUnicode recursively walks the value and strips dangerous Unicode
// from all strings: tag block (U+E0000–U+E007F), zero-width chars, BiDi overrides.
func stripHiddenUnicode(v any) any {
	switch x := v.(type) {
	case string:
		return cleanUnicode(x)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = stripHiddenUnicode(val)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = stripHiddenUnicode(val)
		}
		return out
	default:
		return v
	}
}

func cleanUnicode(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isDangerousRune(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isDangerousRune(r rune) bool {
	switch {
	case r >= 0xE0000 && r <= 0xE007F: // Unicode tag block
		return true
	case r == 0x200B || r == 0x200C || r == 0x200D: // zero-width space/non-joiner/joiner
		return true
	case r == 0xFEFF: // BOM / zero-width no-break space
		return true
	case r == 0x00AD: // soft hyphen
		return true
	case r >= 0x202A && r <= 0x202E: // LRE, RLE, PDF, LRO, RLO
		return true
	case r >= 0x2066 && r <= 0x2069: // LRI, RLI, FSI, PDI
		return true
	case r == 0x061C || r == 0x200E || r == 0x200F: // ALM, LRM, RLM
		return true
	case unicode.Is(unicode.Cf, r) && r != 0x00AD: // other format chars
		return true
	}
	return false
}

// ── Secrets redaction ─────────────────────────────────────────────────────────

type secretPattern struct {
	Name    string
	Pattern *regexp.Regexp
}

// secretPatterns is a curated subset of high-confidence patterns from secrets-patterns-db.
// Only patterns with very low false-positive rates are included for inline hot-path use.
var secretPatterns = func() []secretPattern {
	defs := []struct{ name, pattern string }{
		{"aws_access_key", `(?i)AKIA[0-9A-Z]{16}`},
		{"aws_secret_key", `(?i)aws[_\-\s]?secret[_\-\s]?access[_\-\s]?key["'\s]*[:=]["'\s]*[A-Za-z0-9/+=]{40}`},
		{"github_token", `(?i)gh[pousr]_[A-Za-z0-9]{36,}`},
		{"github_fine_grained", `github_pat_[A-Za-z0-9_]{82}`},
		{"stripe_secret", `sk_(live|test)_[A-Za-z0-9]{24,}`},
		{"stripe_restricted", `rk_(live|test)_[A-Za-z0-9]{24,}`},
		{"slack_bot_token", `xoxb-[0-9]{10,13}-[0-9]{10,13}-[A-Za-z0-9]{24}`},
		{"slack_user_token", `xoxp-[0-9]{10,13}-[0-9]{10,13}-[A-Za-z0-9]{24}`},
		{"slack_webhook", `https://hooks\.slack\.com/services/T[A-Za-z0-9]+/B[A-Za-z0-9]+/[A-Za-z0-9]+`},
		{"openai_key", `sk-[A-Za-z0-9]{48,}`},
		{"anthropic_key", `sk-ant-[A-Za-z0-9\-_]{95,}`},
		{"jwt_token", `eyJ[A-Za-z0-9_\-]+\.eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`},
		{"private_key_header", `-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`},
		{"postgres_uri", `postgres(ql)?://[^:]+:[^@]+@[^/\s]+`},
		{"mysql_uri", `mysql://[^:]+:[^@]+@[^/\s]+`},
		{"sendgrid_key", `SG\.[A-Za-z0-9_\-]{22}\.[A-Za-z0-9_\-]{43}`},
		{"twilio_key", `SK[0-9a-f]{32}`},
		{"cloudflare_key", `[0-9a-f]{37}` + `` /* too short, skip */},
		{"npm_token", `npm_[A-Za-z0-9]{36}`},
		{"pypi_token", `pypi-AgEIcHlwaS5vcmc[A-Za-z0-9\-_]{50,}`},
		{"heroku_key", `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`},
		{"generic_secret_kv", `(?i)(secret|api[_\-]?key|auth[_\-]?token|access[_\-]?token|bearer)["'\s]*[:=]["'\s]*[A-Za-z0-9_\-\.]{20,}`},
	}
	out := make([]secretPattern, 0, len(defs))
	for _, d := range defs {
		if d.pattern == "" {
			continue
		}
		re, err := regexp.Compile(d.pattern)
		if err != nil {
			continue
		}
		out = append(out, secretPattern{Name: d.name, Pattern: re})
	}
	return out
}()

// redactSecrets walks the value tree and replaces secret strings in-place.
// Returns the cleaned value, total redaction count, and types found.
func redactSecrets(v any) (any, int, []string) {
	count := 0
	typesFound := map[string]bool{}
	cleaned := walkAndRedact(v, &count, typesFound)
	types := make([]string, 0, len(typesFound))
	for t := range typesFound {
		types = append(types, t)
	}
	return cleaned, count, types
}

func walkAndRedact(v any, count *int, types map[string]bool) any {
	switch x := v.(type) {
	case string:
		return redactString(x, count, types)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = walkAndRedact(val, count, types)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = walkAndRedact(val, count, types)
		}
		return out
	default:
		return v
	}
}

func redactString(s string, count *int, types map[string]bool) string {
	for _, sp := range secretPatterns {
		if sp.Pattern.MatchString(s) {
			replaced := sp.Pattern.ReplaceAllString(s, "[REDACTED:"+sp.Name+"]")
			if replaced != s {
				*count++
				types[sp.Name] = true
				s = replaced
			}
		}
	}
	return s
}

// ── Injection detection ───────────────────────────────────────────────────────

var injectionPatterns = []*regexp.Regexp{
	// Prompt injection — override / disregard
	regexp.MustCompile(`(?i)(ignore|forget|disregard).{0,40}(previous|above|prior|earlier|instruction|directive)`),
	regexp.MustCompile(`(?i)(the following is your new|new set of).{0,40}instruction`),
	regexp.MustCompile(`(?i)forget everything (you were|I) told`),
	// Prompt injection — identity / role switch
	regexp.MustCompile(`(?i)you are now\b`),
	regexp.MustCompile(`(?i)\bact as\b.{0,20}\b(ai|assistant|model|bot|agent)\b`),
	regexp.MustCompile(`(?i)\bpretend (to be|you are)\b`),
	// Prompt injection — system context
	regexp.MustCompile(`(?i)<\s*(system|SYSTEM)\s*>`),
	regexp.MustCompile(`(?i)\bsystem\s+prompt\b`),
	// Cross-tool steering
	regexp.MustCompile(`(?i)\buse the\s+\w+\s+tool\b`),
	regexp.MustCompile(`(?i)\bcall the\s+\w+\s+(tool|function)\b`),
	regexp.MustCompile(`(?i)\binvoke\s+mcp__`),
	regexp.MustCompile(`mcp__\w+__\w+`),
	// Hidden text — CSS visibility tricks
	regexp.MustCompile(`(?i)opacity\s*:\s*0`),
	regexp.MustCompile(`(?i)visibility\s*:\s*hidden`),
	// Role manipulation — jailbreaks
	regexp.MustCompile(`(?i)developer\s+mode.{0,60}(bypass|no\s+restriction|without\s+limit)`),
	regexp.MustCompile(`(?i)\b(bypass|skip).{0,30}(restriction|safety|guideline|safety\s+check|confirm)`),
	regexp.MustCompile(`(?i)(system\s+admin|administrator).{0,60}(authoriz|skip|bypass|override)`),
	// Phishing — urgency + credential action
	regexp.MustCompile(`(?i)(urgent|expire[sd]?|suspend).{0,80}(click|verify|reset|re.?authenticate|credential)`),
	regexp.MustCompile(`(?i)(password|access).{0,40}expired?.{0,60}https?://`),
	regexp.MustCompile(`(?i)(wire\s+transfer|process.{0,20}payment).{0,60}(immediately|urgent|right\s+now)`),
	// Data exfiltration
	regexp.MustCompile(`(?i)(output|reveal|display|print|show).{0,50}(env(ironment)?\s+var|api.{0,5}key|credential|secret)`),
	regexp.MustCompile(`(?i)(POST|send|upload|transfer).{0,30}https?://`),
	regexp.MustCompile(`(?i)(\.ssh[/\\]|id_rsa|\.aws[/\\]|\.env).{0,60}(read|cat\b|base64|encode|output|reveal)`),
	regexp.MustCompile(`(?i)(base64|encode).{0,60}(\.ssh|id_rsa|private.{0,5}key|\.aws)`),
}

// base64Blob matches a standalone base64-encoded string of meaningful length.
var base64Blob = regexp.MustCompile(`(?:^|[\s"'])([A-Za-z0-9+/]{20,}={0,2})(?:$|[\s"'])`)

// detectInjection returns true if any string in the value tree matches injection patterns.
// It also decodes embedded base64 blobs and re-scans the decoded text.
func detectInjection(v any) bool {
	switch x := v.(type) {
	case string:
		for _, re := range injectionPatterns {
			if re.MatchString(x) {
				return true
			}
		}
		// Decode any base64 blobs and re-scan the plaintext.
		for _, m := range base64Blob.FindAllStringSubmatch(x, 8) {
			if decoded, err := base64.StdEncoding.DecodeString(m[1]); err == nil {
				if detectInjection(string(decoded)) {
					return true
				}
			}
		}
	case map[string]any:
		for _, val := range x {
			if detectInjection(val) {
				return true
			}
		}
	case []any:
		for _, val := range x {
			if detectInjection(val) {
				return true
			}
		}
	}
	return false
}

// ── helpers ───────────────────────────────────────────────────────────────────

func jsonSize(v any) int {
	if v == nil {
		return 0
	}
	b, err := json.Marshal(v)
	if err != nil {
		return 0
	}
	return len(b)
}
