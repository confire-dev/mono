package transport

import "github.com/confire-dev/confire/intercept"

// PassthroughTransport always returns ResultPassthrough — used in tests and fallback paths.
type PassthroughTransport struct{}

func NewPassthrough() *PassthroughTransport {
	return &PassthroughTransport{}
}

func (t *PassthroughTransport) Send(_ intercept.InterceptEvent) (intercept.InterceptResult, error) {
	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
}
