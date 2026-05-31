package optimizer

import (
	"encoding/json"
	"fmt"
	"strings"
)

const webSearchMaxResults = 10

type webSearchResult struct {
	title   string
	url     string
	snippet string
	age     string
}

type WebSearchOptimizer struct{}

func (w *WebSearchOptimizer) Matches(toolName string) bool {
	return toolName == "websearch" ||
		strings.Contains(toolName, "web_search") ||
		strings.Contains(toolName, "brave_") ||
		strings.Contains(toolName, "exa_") ||
		strings.Contains(toolName, "tavily") ||
		strings.Contains(toolName, "perplexity")
}

func (w *WebSearchOptimizer) Optimize(data interface{}) interface{} {
	defer func() { recover() }()

	// The MCP layer unwraps content[0].text before calling optimizers,
	// but handle raw strings too for CLI direct calls.
	text, ok := data.(string)
	if !ok {
		return data
	}
	result := optimizeWebSearchText(text)
	if result == "" || result == text {
		return data
	}
	return result
}

func optimizeWebSearchText(raw string) string {
	if raw == "" {
		return ""
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return ""
	}

	var results []webSearchResult

	switch v := parsed.(type) {
	case map[string]interface{}:
		results = extractFromObject(v)
	case []interface{}:
		results = parseResultArray(v)
	}

	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	limit := len(results)
	if limit > webSearchMaxResults {
		limit = webSearchMaxResults
	}
	for i, r := range results[:limit] {
		fmt.Fprintf(&sb, "%d. %s\n   %s\n", i+1, r.title, r.url)
		if r.snippet != "" {
			fmt.Fprintf(&sb, "   %s\n", r.snippet)
		}
		if r.age != "" {
			fmt.Fprintf(&sb, "   %s\n", r.age)
		}
	}
	if len(results) > webSearchMaxResults {
		fmt.Fprintf(&sb, "[%d more results omitted]\n", len(results)-webSearchMaxResults)
	}
	return strings.TrimSpace(sb.String())
}

func extractFromObject(m map[string]interface{}) []webSearchResult {
	// Brave Search: {"web": {"results": [...]}}
	if web, ok := m["web"].(map[string]interface{}); ok {
		if arr, ok := web["results"].([]interface{}); ok {
			return parseResultArray(arr)
		}
	}
	// Generic / Exa / Tavily: {"results": [...]}
	if arr, ok := m["results"].([]interface{}); ok {
		return parseResultArray(arr)
	}
	return nil
}

func parseResultArray(arr []interface{}) []webSearchResult {
	var out []webSearchResult
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		title := strField(m, "title")
		url := strField(m, "url")
		if title == "" || url == "" {
			continue
		}
		out = append(out, webSearchResult{
			title:   title,
			url:     url,
			snippet: firstNonBlank(strField(m, "description"), strField(m, "content"), strField(m, "text"), strField(m, "snippet")),
			age:     firstNonBlank(strField(m, "age"), strField(m, "publishedDate"), strField(m, "date")),
		})
	}
	return out
}

func strField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func firstNonBlank(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
