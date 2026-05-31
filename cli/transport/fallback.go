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
	if err != nil {
		// Network / transport failure → local fallback.
		return t.secondary.Send(event)
	}
	// Worker responded but couldn't optimize (quota exhausted, entitlement denied,
	// passthrough) → try local so the user still gets something useful.
	if result.Kind == intercept.ResultPassthrough || result.Kind == intercept.ResultAddContext {
		if local, lerr := t.secondary.Send(event); lerr == nil && local.Kind != intercept.ResultPassthrough {
			return local, nil
		}
	}
	return result, nil
}

func (t *FallbackTransport) Mode() OptimizerMode { return t.primary.Mode() }
