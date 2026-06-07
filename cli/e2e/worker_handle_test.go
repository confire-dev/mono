package e2e_test

import (
	"github.com/confire-dev/confire/intercept"
)

// workerHandle simulates what the real Cloudflare Worker does.
// Since optimization has been removed, all events return passthrough.
// The worker now focuses solely on firewall/security functions.
func workerHandle(event intercept.InterceptEvent) intercept.InterceptResult {
	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
}
