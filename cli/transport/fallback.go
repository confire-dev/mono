package transport

import "github.com/confire-dev/confire/intercept"

// FallbackTransport tries the primary Transport and falls back to the secondary
// on any error. Ensures we never return raw output due to a network blip.
type FallbackTransport struct {
	primary   Transport
	secondary Transport
}

func NewFallback(primary, secondary Transport) *FallbackTransport {
	return &FallbackTransport{primary: primary, secondary: secondary}
}

func (t *FallbackTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	result, err := t.primary.Send(event)
	if err == nil {
		return result, nil
	}
	// Primary failed (network down, Worker unreachable, etc.) → local fallback.
	// Never propagate the error upward — never return raw output.
	return t.secondary.Send(event)
}

func (t *FallbackTransport) Mode() OptimizerMode { return t.primary.Mode() }
