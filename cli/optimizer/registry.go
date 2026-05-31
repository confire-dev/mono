package optimizer

import "strings"

type Optimizer interface {
	Matches(toolName string) bool
	Optimize(data interface{}) interface{}
}

type Registry struct {
	optimizers []Optimizer
	fallback   Optimizer
}

// NewRegistry returns the LOCAL optimizer registry (free tier).
// Covers infrastructure tools: Bash, Read, WebFetch, WebSearch + Generic fallback.
// Platform-specific MCP optimizers (Figma, GitHub, Slack, etc.) run in the
// Cloudflare Worker only via WorkerTransport — they are the paid tier.
func NewRegistry(_ string) *Registry {
	return &Registry{
		// Order matters: first match wins. Bash before generic so bash output
		// isn't parsed as JSON.
		optimizers: []Optimizer{
			&BashOptimizer{},
			&ReadOptimizer{},
			&WebFetchOptimizer{},
			&WebSearchOptimizer{},
		},
		fallback: &GenericOptimizer{},
	}
}

// NewWorkerRegistry returns the FULL optimizer registry including Figma and all
// platform-specific MCP optimizers. Used by the mock Worker in E2E tests and
// any future self-hosted Worker mode.
func NewWorkerRegistry(hint string) *Registry {
	h := strings.ToLower(hint)
	figma := &FigmaOptimizer{matchAll: strings.Contains(h, "figma")}
	return &Registry{
		optimizers: []Optimizer{
			figma,
			&BashOptimizer{},
			&ReadOptimizer{},
			&WebFetchOptimizer{},
			&WebSearchOptimizer{},
		},
		fallback: &GenericOptimizer{},
	}
}

func (r *Registry) Resolve(toolName string) Optimizer {
	lower := strings.ToLower(toolName)
	for _, opt := range r.optimizers {
		if opt.Matches(lower) {
			return opt
		}
	}
	return r.fallback
}
