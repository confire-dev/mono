package transport

import "github.com/confire-dev/confire/intercept"

// LocalFirstTransport runs the local optimizer first.
// Only delegates to cloud when local passes through — i.e. MCP tools and
// anything the local optimizer can't handle. This ensures Bash/Read/WebFetch
// are always optimized locally (zero latency, works offline, works on free plan),
// while MCP/platform tools still reach the cloud optimizer for paid users.
type LocalFirstTransport struct {
	local Transport
	cloud Transport
}

func NewLocalFirst(local, cloud Transport) *LocalFirstTransport {
	return &LocalFirstTransport{local: local, cloud: cloud}
}

func (t *LocalFirstTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	result, err := t.local.Send(event)
	if err == nil && result.Kind != intercept.ResultPassthrough {
		return result, nil
	}
	// Only escalate to cloud for MCP tools — built-in tools (Bash, Read, WebFetch)
	// are fully covered locally; calling cloud just produces an entitlement message.
	if event.Tool == nil || !event.Tool.IsMCP {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}
	return t.cloud.Send(event)
}

func (t *LocalFirstTransport) Mode() OptimizerMode { return OptimizerModeRemote }
