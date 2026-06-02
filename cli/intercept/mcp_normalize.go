package intercept

import (
	"encoding/json"
	"strings"
)

// MCPNormalizeHandler runs PostToolUse normalization passes on MCP tool output:
// 1. Null/empty pruner — removes null values, empty arrays, empty objects
// 2. Array truncation — large arrays are head+tail sliced with metadata injected
// 3. String field budget — very long string fields are truncated
type MCPNormalizeHandler struct {
	// MaxArrayItems is the max array length before truncation (default 20).
	MaxArrayItems int
	// MaxStringBytes is the max string value length before truncation (default 8192).
	MaxStringBytes int
}

func NewMCPNormalizeHandler() *MCPNormalizeHandler {
	return &MCPNormalizeHandler{
		MaxArrayItems:  20,
		MaxStringBytes: 8192,
	}
}

func (h *MCPNormalizeHandler) ID() string     { return "mcp.normalize" }
func (h *MCPNormalizeHandler) Phases() []Phase { return []Phase{PhaseToolPost} }

func (h *MCPNormalizeHandler) Matches(e InterceptEvent) bool {
	return e.Tool != nil && e.Tool.IsMCP && e.Tool.Output != nil
}

func (h *MCPNormalizeHandler) Run(e InterceptEvent) (InterceptResult, error) {
	before := jsonSize(e.Tool.Output)

	truncated := false
	pruned := pruneEmpty(e.Tool.Output)
	out := h.normalizeValue(pruned, &truncated)

	after := jsonSize(out)
	if after >= before && !truncated {
		return InterceptResult{Kind: ResultPassthrough}, nil
	}

	return InterceptResult{
		Kind:       ResultSanitize,
		ToolOutput: out,
		Stats: &Stats{
			BeforeBytes: before,
			AfterBytes:  after,
			Optimizer:   "mcp.normalize",
		},
	}, nil
}

// pruneEmpty removes null values and empty arrays/objects recursively.
func pruneEmpty(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			if val == nil {
				continue
			}
			pruned := pruneEmpty(val)
			// Skip empty collections after pruning.
			switch p := pruned.(type) {
			case map[string]any:
				if len(p) == 0 {
					continue
				}
			case []any:
				if len(p) == 0 {
					continue
				}
			}
			out[k] = pruned
		}
		return out
	case []any:
		out := make([]any, 0, len(x))
		for _, val := range x {
			if val == nil {
				continue
			}
			out = append(out, pruneEmpty(val))
		}
		return out
	default:
		return v
	}
}

func (h *MCPNormalizeHandler) normalizeValue(v any, truncated *bool) any {
	switch x := v.(type) {
	case string:
		return h.normalizeString(x, truncated)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = h.normalizeValue(val, truncated)
		}
		return out
	case []any:
		return h.normalizeArray(x, truncated)
	default:
		return v
	}
}

func (h *MCPNormalizeHandler) normalizeArray(arr []any, truncated *bool) any {
	max := h.MaxArrayItems
	if max <= 0 {
		max = 20
	}
	if len(arr) <= max {
		out := make([]any, len(arr))
		for i, v := range arr {
			out[i] = h.normalizeValue(v, truncated)
		}
		return out
	}

	// Keep first 10 + last 3, inject truncation marker.
	head := 10
	tail := 3
	if head+tail >= max {
		head = max - 3
		tail = 3
	}
	omitted := len(arr) - head - tail
	*truncated = true

	out := make([]any, 0, head+1+tail)
	for _, v := range arr[:head] {
		out = append(out, h.normalizeValue(v, truncated))
	}
	out = append(out, map[string]any{
		"_confire_truncated": true,
		"_omitted_count":     omitted,
		"_total":             len(arr),
	})
	for _, v := range arr[len(arr)-tail:] {
		out = append(out, h.normalizeValue(v, truncated))
	}
	return out
}

func (h *MCPNormalizeHandler) normalizeString(s string, truncated *bool) string {
	max := h.MaxStringBytes
	if max <= 0 {
		max = 8192
	}
	if len(s) <= max {
		return s
	}
	*truncated = true
	// Keep first 80% + ellipsis + last 5%.
	head := int(float64(max) * 0.80)
	tail := int(float64(max) * 0.05)
	return s[:head] + "\n…[confire: truncated " + itoa(len(s)-head-tail) + " bytes]…\n" + s[len(s)-tail:]
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	b, _ := json.Marshal(n)
	return strings.TrimSpace(string(b))
}
