package sanitize

import (
	"strings"
	"testing"
)

func TestRedact_CommonSecrets(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantTag string // expected [REDACTED_SECRET:<tag>] substring
	}{
		{"aws key", "key=AKIAIOSFODNN7EXAMPLE more text", "aws_access_key"},
		{"github classic token", "token=ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ01234567890 done", "github_token"},
		{"openai key", "sk-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJ", "openai_key"},
		{"jwt", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", "jwt"},
		{"bearer", "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9abc123longtokenvalue", "bearer_token"},
		{"private key", "-----BEGIN PRIVATE KEY----- MIIEvQIBADAN", "private_key"},
		{"database url", "postgres://user:secretpassword@localhost:5432/mydb", "database_url"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, count, types := Redact(tc.input)
			if count == 0 {
				t.Errorf("expected at least 1 redaction, got 0. Input: %q", tc.input)
				return
			}
			if !strings.Contains(out, "[REDACTED_SECRET:"+tc.wantTag+"]") {
				t.Errorf("expected [REDACTED_SECRET:%s] in output, got: %q", tc.wantTag, out)
			}
			found := false
			for _, typ := range types {
				if typ == tc.wantTag {
					found = true
				}
			}
			if !found {
				t.Errorf("expected type %q in types list %v", tc.wantTag, types)
			}
		})
	}
}

func TestRedact_SafeContent(t *testing.T) {
	cases := []string{
		"git status",
		"Hello, world!",
		"package.json",
		"function myFunc() { return 42; }",
		"const API_URL = 'https://api.example.com'",
	}
	for _, input := range cases {
		out, count, _ := Redact(input)
		if count > 0 {
			t.Errorf("safe input %q was incorrectly redacted to %q", input, out)
		}
		if out != input {
			t.Errorf("safe input was modified: %q → %q", input, out)
		}
	}
}
