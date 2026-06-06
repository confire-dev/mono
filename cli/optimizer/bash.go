package optimizer

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	// Emergency caps — last-resort safety net only.
	// Structural cleanup (test/build/repeated-line/ANSI) runs first;
	// these fire only for extremely large outputs.
	bashEmergencyBytes = 512 * 1024                              // 512 KB
	bashEmergencyHead  = 200 * 1024                              // 200 KB from the top
	bashEmergencyTail  = bashEmergencyBytes - bashEmergencyHead  // 312 KB from the bottom
)

var (
	testPassRe     = regexp.MustCompile(`(?i)^\s*(✓|✔|PASS|passing|\d+ passing|ok\s+\S|·\s+✓)`)
	testFailRe     = regexp.MustCompile(`(?i)^\s*(✗|✘|FAIL|failing|not ok|×|\d+ failing|Error:|AssertionError)`)
	progressRe     = regexp.MustCompile(`[─╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r`)
	ansiRe         = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	testRunnerRe   = regexp.MustCompile(`(?i)\b(jest|vitest|mocha|pytest|cargo\s+test|go\s+test|npm\s+test|yarn\s+test)\b`)
	summaryCntRe   = regexp.MustCompile(`(?i)\d+\s+(passing|failing|skipped|pending)`)
	buildErrorRe   = regexp.MustCompile(`(?i)(error TS\d+|error\[E\d+\]|SyntaxError:|Cannot find|Module not found|ld: error|make\[1\].*Error)`)
	buildSuccessRe = regexp.MustCompile(`(?i)(Successfully compiled|Compiled \d+ files|webpack.*built|cargo.*Finished|Build succeeded)`)
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

	// 1. Strip ANSI escape codes.
	result := ansiRe.ReplaceAllString(output, "")

	lines := strings.Split(result, "\n")

	// 2. Structural cleanup by output type.
	if isTestOutput(result, cmd) {
		lines = strings.Split(filterTestOutput(lines), "\n")
	} else if isBuildOutput(result) {
		lines = strings.Split(filterBuildOutput(lines), "\n")
	}

	// 3. Strip progress-bar / spinner lines.
	filtered := lines[:0]
	for _, line := range lines {
		if !progressRe.MatchString(line) {
			filtered = append(filtered, line)
		}
	}
	lines = filtered

	// 4. Collapse consecutive identical lines (≥3 repeats).
	lines = collapseRepeatedLines(lines)

	result = strings.Join(lines, "\n")

	// 5. Emergency byte cap — fires only for very large output.
	if len(result) > bashEmergencyBytes {
		dropped := len(result) - bashEmergencyHead - bashEmergencyTail
		if dropped > 0 {
			result = result[:bashEmergencyHead] +
				"\n[confire: " + strconv.Itoa(dropped) + " bytes omitted — output exceeded 512KB]\n" +
				result[len(result)-bashEmergencyTail:]
		}
	}

	if len(result) >= len(output) {
		return output
	}
	return result
}

// collapseRepeatedLines collapses runs of identical consecutive lines.
// A run of 3 or more identical lines is replaced by the first line plus a marker.
func collapseRepeatedLines(lines []string) []string {
	const minRepeat = 3
	if len(lines) < minRepeat {
		return lines
	}
	out := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		j := i + 1
		for j < len(lines) && lines[j] == lines[i] {
			j++
		}
		count := j - i
		if count >= minRepeat {
			out = append(out, lines[i])
			out = append(out, "[confire: "+strconv.Itoa(count-1)+" identical lines omitted]")
		} else {
			out = append(out, lines[i:j]...)
		}
		i = j
	}
	return out
}

func isTestOutput(text, cmd string) bool {
	return testPassRe.MatchString(text) || testFailRe.MatchString(text) ||
		testRunnerRe.MatchString(cmd)
}

func isBuildOutput(text string) bool {
	return buildErrorRe.MatchString(text) && buildSuccessRe.MatchString(text)
}

func filterBuildOutput(lines []string) string {
	var errorLines []string
	var progressCount int
	for _, line := range lines {
		if buildErrorRe.MatchString(line) {
			errorLines = append(errorLines, line)
		} else if buildSuccessRe.MatchString(line) || progressRe.MatchString(line) {
			progressCount++
		} else {
			errorLines = append(errorLines, line)
		}
	}
	if progressCount == 0 {
		return strings.Join(lines, "\n")
	}
	header := "[confire: " + strconv.Itoa(progressCount) + " build progress lines omitted]"
	return header + "\n" + strings.Join(errorLines, "\n")
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
		header := "[confire: " + strconv.Itoa(passCount) + " passing tests omitted]"
		out = append([]string{header}, out...)
	}
	return strings.Join(out, "\n")
}
