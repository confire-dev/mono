// Package transport defines how InterceptEvents reach the daemon and remote backends.
//
// DaemonClient — used by the hook process to send events to the local daemon via Unix socket.
// PassthroughTransport — used in tests and fallback paths.
// WorkerTransport — ships security events and telemetry to the Cloudflare Worker.
package transport

import "github.com/confire-dev/confire/intercept"

// Transport sends an InterceptEvent and returns an InterceptResult.
// Implementations must never panic — on any error return ResultPassthrough.
type Transport interface {
	Send(event intercept.InterceptEvent) (intercept.InterceptResult, error)
}
