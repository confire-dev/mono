package optimizer

import "fmt"

// ReadOptimizer applies emergency-only caps — no structural truncation by default.
// Large reads pass through intact unless they exceed the emergency threshold.
// Full session deduplication (repeated reads) runs in the Worker.
const (
	readEmergencyBytes = 1 * 1024 * 1024 // 1 MB
	readEmergencyHead  = 800 * 1024       // keep first 800 KB
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
	if len(text) <= readEmergencyBytes {
		return text
	}

	// Emergency cap only — preserve as much of the file as possible.
	dropped := len(text) - readEmergencyHead
	return text[:readEmergencyHead] +
		fmt.Sprintf("\n[confire: %d bytes omitted — file exceeds 1MB; use offset/limit to read further]", dropped)
}
