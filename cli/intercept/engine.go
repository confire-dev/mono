package intercept

// Handler processes an InterceptEvent for one or more phases.
type Handler interface {
	ID() string
	Phases() []Phase
	Matches(e InterceptEvent) bool
	Run(e InterceptEvent) (InterceptResult, error)
}

// Engine routes InterceptEvents to registered Handlers.
type Engine struct {
	handlers []Handler
}

func NewEngine(handlers ...Handler) *Engine {
	return &Engine{handlers: handlers}
}

func (e *Engine) Handle(event InterceptEvent) (InterceptResult, error) {
	// Pre-compact is disabled in v1 — ships dark.
	if event.Phase == PhasePreCompact {
		return InterceptResult{Kind: ResultPassthrough}, nil
	}

	for _, h := range e.handlers {
		if !matchesPhase(h, event.Phase) {
			continue
		}
		if !h.Matches(event) {
			continue
		}
		result, err := h.Run(event)
		if err != nil {
			// Handler failed → pass through; never block the developer.
			return InterceptResult{Kind: ResultPassthrough}, nil
		}
		if result.Kind != ResultPassthrough {
			return result, nil
		}
	}
	return InterceptResult{Kind: ResultPassthrough}, nil
}

func matchesPhase(h Handler, phase Phase) bool {
	for _, p := range h.Phases() {
		if p == phase {
			return true
		}
	}
	return false
}
