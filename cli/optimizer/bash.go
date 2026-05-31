package optimizer

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	bashMaxLines = 200
	bashHeadLines = 20
	bashTailLines = 50
	bashMaxBytes = 40_000
)

var (
	testPassRe     = regexp.MustCompile(`(?i)^\s*(✓|✔|PASS|passing|\d+ passing|ok\s+\S|·\s+✓)`)
	testFailRe     = regexp.MustCompile(`(?i)^\s*(✗|✘|FAIL|failing|not ok|×|\d+ failing|Error:|AssertionError)`)
	progressRe     = regexp.MustCompile(`[─╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r`)
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

	lines := strings.Split(output, "\n")
	if len(lines) < 20 && len(output) < bashMaxBytes {
		return output
	}

	result := output

	if isTestOutput(output, cmd) {
		result = filterTestOutput(lines)
	} else if isBuildOutput(output) {
		result = filterBuildOutput(lines)
	} else if len(lines) > bashMaxLines {
		dropped := len(lines) - bashHeadLines - bashTailLines
		if dropped > 0 {
			head := lines[:bashHeadLines]
			tail := lines[len(lines)-bashTailLines:]
			marker := []string{"", "[confire: " + strconv.Itoa(dropped) + " lines omitted]", ""}
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

	// Byte-budget head+tail for few-lines-but-large outputs
	const headBytes = 20_000
	const tailBytes = 10_000
	if len(result) > bashMaxBytes && len(result) > headBytes+tailBytes {
		dropped := len(result) - headBytes - tailBytes
		result = result[:headBytes] +
			"\n[confire: " + strconv.Itoa(dropped) + " bytes omitted]\n" +
			result[len(result)-tailBytes:]
	} else if len(result) > bashMaxBytes {
		result = result[:bashMaxBytes] + "\n[confire: output truncated at " + strconv.Itoa(bashMaxBytes) + " bytes]"
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

