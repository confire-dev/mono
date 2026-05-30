package optimizer

import (
	"regexp"
	"strings"
)

const (
	bashMaxLines = 200
	bashHeadLines = 20
	bashTailLines = 50
	bashMaxBytes = 40_000
)

var (
	testPassRe = regexp.MustCompile(`(?i)^\s*(✓|✔|PASS|passing|\d+ passing|ok\s+\S|·\s+✓)`)
	testFailRe = regexp.MustCompile(`(?i)^\s*(✗|✘|FAIL|failing|not ok|×|\d+ failing|Error:|AssertionError)`)
	progressRe = regexp.MustCompile(`[─╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r`)
	testRunnerRe = regexp.MustCompile(`(?i)\b(jest|vitest|mocha|pytest|cargo\s+test|go\s+test|npm\s+test|yarn\s+test)\b`)
	summaryCntRe = regexp.MustCompile(`(?i)\d+\s+(passing|failing|skipped|pending)`)
)

type BashOptimizer struct{}

func (b *BashOptimizer) Matches(toolName string) bool {
	return toolName == "bash"
}

func (b *BashOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	text, ok := data.(string)
	if !ok {
		return data
	}
	result := optimizeBashOutput(text, "")
	if result == text {
		return data
	}
	return result
}

// OptimizeBashWithCmd allows passing the original command for better heuristics.
func OptimizeBashWithCmd(output, cmd string) string {
	return optimizeBashOutput(output, cmd)
}

func optimizeBashOutput(output, cmd string) string {
	if len(output) < 500 {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 20 && len(output) < bashMaxBytes {
		return output
	}

	result := output

	if isTestOutput(output, cmd) {
		result = filterTestOutput(lines)
	} else if len(lines) > bashMaxLines {
		dropped := len(lines) - bashHeadLines - bashTailLines
		if dropped > 0 {
			head := lines[:bashHeadLines]
			tail := lines[len(lines)-bashTailLines:]
			marker := []string{"", "[confire: " + itoa(dropped) + " lines omitted]", ""}
			combined := append(head, marker...)
			combined = append(combined, tail...)
			result = strings.Join(combined, "\n")
		}
	}

	// Strip progress-bar lines
	var kept []string
	for _, line := range strings.Split(result, "\n") {
		if !progressRe.MatchString(line) {
			kept = append(kept, line)
		}
	}
	result = strings.Join(kept, "\n")

	// Hard byte cap
	if len(result) > bashMaxBytes {
		result = result[:bashMaxBytes] + "\n[confire: output truncated at " + itoa(bashMaxBytes) + " bytes]"
	}

	if len(result) >= len(output) {
		return output
	}
	return result
}

func isTestOutput(text, cmd string) bool {
	return testPassRe.MatchString(text) || testFailRe.MatchString(text) ||
		testRunnerRe.MatchString(cmd)
}

func filterTestOutput(lines []string) string {
	var out []string
	passCount := 0
	inFailBlock := false

	for _, line := range lines {
		switch {
		case testFailRe.MatchString(line):
			inFailBlock = true
			out = append(out, line)
		case summaryCntRe.MatchString(line):
			inFailBlock = false
			out = append(out, line)
		case testPassRe.MatchString(line) && !inFailBlock:
			passCount++
		default:
			out = append(out, line)
		}
	}

	if passCount > 0 {
		header := "[confire: " + itoa(passCount) + " passing tests omitted]"
		out = append([]string{header}, out...)
	}
	return strings.Join(out, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
