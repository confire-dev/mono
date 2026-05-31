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

	// Claude Code sends {"type":"text","file":{"filePath":"...","content":"..."}}
	if m, ok := data.(map[string]interface{}); ok {
		if fileObj, ok := m["file"].(map[string]interface{}); ok {
			if content, ok := fileObj["content"].(string); ok {
				optimized := r.optimizeText(content)
				if optimized == content {
					return data
				}
				newFile := make(map[string]interface{}, len(fileObj))
				for k, v := range fileObj {
					newFile[k] = v
				}
				newFile["content"] = optimized
				result := make(map[string]interface{}, len(m))
				for k, v := range m {
					result[k] = v
				}
				result["file"] = newFile
				return result
			}
		}
	}

	// Fallback: plain string (tests, direct calls)
	text, ok := data.(string)
	if !ok {
		return data
	}
	return r.optimizeText(text)
}

func (r *ReadOptimizer) optimizeText(text string) string {
	if len(text) <= readMaxBytes {
		return text
	}

	lines := strings.Split(text, "\n")
	if len(lines) > readMaxLines {
		head := lines[:readMaxLines]
		dropped := len(lines) - readMaxLines
		return strings.Join(head, "\n") +
			fmt.Sprintf("\n[confire: %d lines truncated — use offset/limit params to read more]", dropped)
	}

	truncated := text[:readMaxBytes]
	dropped := len(text) - readMaxBytes
	return truncated + fmt.Sprintf("\n[confire: %d bytes truncated]", dropped)
}
