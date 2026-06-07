package transport

import (
	"encoding/json"
	"net"
	"time"

	"github.com/confire-dev/confire/intercept"
)

// DaemonTransport is used by the hook shim to talk to the long-lived daemon
// over a unix socket. If the daemon isn't running (connect fails), it falls
// back to the provided secondary Transport so the developer is never blocked.
type DaemonTransport struct {
	socketPath string
	fallback   Transport
}

func NewDaemonClient(socketPath string, fallback Transport) *DaemonTransport {
	return &DaemonTransport{socketPath: socketPath, fallback: fallback}
}

func (t *DaemonTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	// Fast connect attempt — if the daemon isn't running we find out in 50ms.
	conn, err := net.DialTimeout("unix", t.socketPath, 50*time.Millisecond)
	if err != nil {
		return t.fallback.Send(event)
	}
	defer conn.Close()

	// Give the daemon up to 8 seconds total (generous for cloud round-trips).
	conn.SetDeadline(time.Now().Add(8 * time.Second))

	if err := json.NewEncoder(conn).Encode(event); err != nil {
		return t.fallback.Send(event)
	}

	var result intercept.InterceptResult
	if err := json.NewDecoder(conn).Decode(&result); err != nil {
		return t.fallback.Send(event)
	}
	return result, nil
}

