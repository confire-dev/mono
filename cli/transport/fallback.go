package transport

import "github.com/confire-dev/confire/intercept"

// FallbackTransport tries the primary Transport and falls back to the secondary
// on any error. Ensures we never return passthrough due to a network blip.
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
		return t.secondary.Send(event)
	}
	return result, nil
}
