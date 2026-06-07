package transport

import (
	"github.com/confire-dev/confire/intercept"
)

// LocalTransport runs the local security intercept engine without contacting
// the Cloudflare Worker. Used for offline / fallback paths and in tests.
//
// The optional hintServer parameter is currently unused but kept for API
// compatibility with call sites that pass a server name hint.
type LocalTransport struct {
	engine *intercept.Engine
}

// NewLocal creates a LocalTransport that runs the built-in security handlers.
// hintServer is accepted for forward-compatibility but currently unused.
func NewLocal(hintServer string) *LocalTransport {
	handlers := []intercept.Handler{
		&intercept.MCPRiskHandler{},
		&intercept.MCPSanitizeHandler{},
	}
	return &LocalTransport{
		engine: intercept.NewEngine(handlers...),
	}
}

func (t *LocalTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	result, err := t.engine.Handle(event)
	if err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}
	return result, nil
}
