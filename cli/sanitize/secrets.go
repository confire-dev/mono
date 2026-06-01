// Package sanitize provides secret detection/redaction and prompt-injection
// detection/sanitization for tool outputs before they enter Claude's context.
package sanitize

import (
	"fmt"
	"regexp"
	"strings"
)

// SecretType identifies the kind of secret detected.
type SecretType string

const (
	SecretAWSAccessKey  SecretType = "aws_access_key"
	SecretGitHubToken   SecretType = "github_token"
	SecretOpenAIKey     SecretType = "openai_key"
	SecretAnthropicKey  SecretType = "anthropic_key"
	SecretStripeKey     SecretType = "stripe_key"
	SecretJWT           SecretType = "jwt"
	SecretBearerToken   SecretType = "bearer_token"
	SecretPrivateKey    SecretType = "private_key"
	SecretDatabaseURL   SecretType = "database_url"
	SecretGenericEnvVal SecretType = "env_secret_value"
)

// SecretMatch is one detected secret occurrence.
type SecretMatch struct {
	Type  SecretType
	Value string // the matched text (for internal use only — never log this)
	Start int
	End   int
}

type secretPattern struct {
	kind    SecretType
	pattern *regexp.Regexp
}

var secretPatterns = []secretPattern{
	// AWS access key IDs
	{SecretAWSAccessKey, regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	// GitHub tokens (classic and fine-grained)
	{SecretGitHubToken, regexp.MustCompile(`\bgh[pours]_[A-Za-z0-9_]{36,}\b`)},
	// OpenAI API keys
	{SecretOpenAIKey, regexp.MustCompile(`\bsk-[A-Za-z0-9]{32,}\b`)},
	// Anthropic API keys
	{SecretAnthropicKey, regexp.MustCompile(`\bsk-ant-[A-Za-z0-9\-_]{20,}\b`)},
	// Stripe secret keys (live)
	{SecretStripeKey, regexp.MustCompile(`\bsk_live_[A-Za-z0-9]{24,}\b`)},
	// JWTs (three base64url segments separated by dots)
	{SecretJWT, regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`)},
	// Bearer tokens in HTTP headers/output
	{SecretBearerToken, regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]{20,}`)},
	// PEM private key blocks
	{SecretPrivateKey, regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
	// Database URLs with embedded credentials
	{SecretDatabaseURL, regexp.MustCompile(`(?i)\b(?:postgres|postgresql|mysql|mongodb|redis)://[^:@\s]+:[^@\s]+@[^\s]+`)},
	// .env-style KEY=<long-looking-value> (heuristic: value ≥ 20 chars, alphanumeric+special)
	{SecretGenericEnvVal, regexp.MustCompile(`(?m)^[A-Z][A-Z0-9_]{4,}=([A-Za-z0-9/+\-_]{20,}|"[A-Za-z0-9/+\-_]{20,}")`)},
}

// Scan returns all detected secret matches in text.
func Scan(text string) []SecretMatch {
	var matches []SecretMatch
	seen := make(map[string]bool)
	for _, p := range secretPatterns {
		locs := p.pattern.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			val := text[loc[0]:loc[1]]
			key := p.kind.String() + ":" + val
			if seen[key] {
				continue
			}
			seen[key] = true
			matches = append(matches, SecretMatch{
				Type:  p.kind,
				Value: val,
				Start: loc[0],
				End:   loc[1],
			})
		}
	}
	return matches
}

// Redact replaces detected secrets in text with [REDACTED_SECRET:<type>] placeholders.
// Returns the sanitized text, count of redactions, and the distinct types found.
func Redact(text string) (redacted string, count int, types []string) {
	matches := Scan(text)
	if len(matches) == 0 {
		return text, 0, nil
	}

	// Apply replacements right-to-left so offsets stay valid.
	runes := []byte(text)
	typeSet := make(map[string]bool)
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		placeholder := "[REDACTED_SECRET:" + string(m.Type) + "]"
		runes = append(runes[:m.Start], append([]byte(placeholder), runes[m.End:]...)...)
		typeSet[string(m.Type)] = true
		count++
	}

	for t := range typeSet {
		types = append(types, t)
	}
	return string(runes), count, types
}

// RedactString is a convenience wrapper that always returns a string.
func RedactString(text string) string {
	out, _, _ := Redact(text)
	return out
}

func (s SecretType) String() string { return string(s) }

// OutputToString coerces a tool output value to a string for scanning.
// Returns "", false if the value cannot be meaningfully scanned (e.g. already structured).
func OutputToString(output any) (string, bool) {
	switch v := output.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		if output == nil {
			return "", false
		}
		// For maps/slices, do a shallow string representation only if it's not too deep.
		s := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(
			strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " "), "\t", " "), "  ", " "))
		if len(s) > 0 {
			return s, true
		}
		return "", false
	}
}
