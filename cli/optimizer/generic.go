package optimizer

import (
	"encoding/json"
	"strings"
)

// GenericOptimizer applies universal noise stripping to any MCP tool response.
// Rules ported from leanmcp/optimizers/generic.js:
//   1. Strip always-strip fields (avatars, internal IDs, URL chains)
//   2. Drop null / empty values
//   3. Drop *_url fields except keep-list
//   4. Try to optimize text content inside content[0].text if it's JSON
type GenericOptimizer struct{}

func (g *GenericOptimizer) Matches(_ string) bool { return true }

func (g *GenericOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	// Try to optimize text content inside MCP content wrapper first
	if m, ok := data.(map[string]interface{}); ok {
		if optimized := g.tryOptimizeContent(m); optimized != nil {
			return optimized
		}
	}

	return stripNulls(stripURLFields(data))
}

// tryOptimizeContent checks for content[0].text that is JSON and strips it.
// Returns nil if nothing to optimize (caller falls back to full-object strip).
func (g *GenericOptimizer) tryOptimizeContent(m map[string]interface{}) interface{} {
	content, ok := m["content"].([]interface{})
	if !ok || len(content) == 0 {
		return nil
	}
	first, ok := content[0].(map[string]interface{})
	if !ok {
		return nil
	}
	text, ok := first["text"].(string)
	if !ok || text == "" {
		return nil
	}

	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return nil
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}

	cleaned := stripNulls(stripURLFields(parsed))
	out, err := json.Marshal(cleaned)
	if err != nil || len(out) >= len(text) {
		return nil
	}

	newFirst := make(map[string]interface{}, len(first))
	for k, v := range first {
		newFirst[k] = v
	}
	newFirst["text"] = string(out)

	newContent := make([]interface{}, len(content))
	copy(newContent, content)
	newContent[0] = newFirst

	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	result["content"] = newContent
	return result
}
