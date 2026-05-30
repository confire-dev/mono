package optimizer

import (
	"fmt"
	"strings"
)

// ReadOptimizer caps large file reads before they bloat context.
// Infrastructure optimizer — runs locally as part of the free tier.
// Full session deduplication (repeated reads) runs in the Worker.
const (
	readMaxLines = 500
	readMaxBytes = 80_000
)

type ReadOptimizer struct{}

func (r *ReadOptimizer) Matches(toolName string) bool {
	return toolName == "read"
}

func (r *ReadOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	text, ok := data.(string)
	if !ok {
		return data
	}

	if len(text) <= readMaxBytes {
		return data
	}

	lines := strings.Split(text, "\n")
	if len(lines) > readMaxLines {
		head := lines[:readMaxLines]
		dropped := len(lines) - readMaxLines
		return strings.Join(head, "\n") +
			fmt.Sprintf("\n[confire: %d lines truncated — use offset/limit params to read more]", dropped)
	}

	// Over byte limit but not line limit
	truncated := text[:readMaxBytes]
	dropped := len(text) - readMaxBytes
	return truncated + fmt.Sprintf("\n[confire: %d bytes truncated]", dropped)
}
