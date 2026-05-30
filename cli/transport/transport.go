// Package transport defines how InterceptEvents reach an optimizer.
//
// Two optimizer modes — names describe WHERE the optimizer runs, not pricing:
//
//   OptimizerModeLocal  — runs in-process in the CLI daemon.
//                          Tools: Bash, Read, WebFetch, Generic.
//                          No network, zero latency, works offline.
//
//   OptimizerModeRemote — ships to the Cloudflare Worker.
//                          Tools: Figma, GitHub PR, Linear, Jira, Slack, etc.
//                          Requires a valid API key.
//                          FallbackTransport degrades to Local if unreachable.
//
// Entitlement (who can access Remote) is resolved at the daemon level by
// checking the keychain for a valid API key. The engine is mode-agnostic.
package transport

import "github.com/confire-dev/confire/intercept"

// OptimizerMode describes which optimizer layer handles a request.
// Use Local/Remote — not Free/Paid — to keep business logic out of the engine.
type OptimizerMode string

const (
	OptimizerModeLocal  OptimizerMode = "local"
	OptimizerModeRemote OptimizerMode = "remote"
)

// Transport sends an InterceptEvent and returns an InterceptResult.
// Implementations must never panic — on any error return ResultPassthrough.
type Transport interface {
	Send(event intercept.InterceptEvent) (intercept.InterceptResult, error)
	Mode() OptimizerMode
}
