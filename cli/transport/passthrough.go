package transport

import "github.com/confire-dev/confire/intercept"

// PassthroughTransport always returns ResultPassthrough without running any optimizer.
// Used when no API key is configured — optimization requires an account.
type PassthroughTransport struct{}

func NewPassthrough() *PassthroughTransport {
	return &PassthroughTransport{}
}

func (t *PassthroughTransport) Send(_ intercept.InterceptEvent) (intercept.InterceptResult, error) {
	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
}

func (t *PassthroughTransport) Mode() OptimizerMode { return OptimizerModeLocal }
