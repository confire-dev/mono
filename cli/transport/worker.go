package transport

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/confire-dev/confire/intercept"
)

const defaultWorkerURL = "https://api.confire.dev"

type optimizeRequest struct {
	Event  intercept.InterceptEvent `json:"event"`
	APIKey string                   `json:"apiKey"`
}

type optimizeResponse struct {
	Result  intercept.InterceptResult `json:"result"`
	Warning string                    `json:"_warning,omitempty"`
}

// WorkerTransport ships InterceptEvents to the Cloudflare Worker over HTTPS/HTTP2.
// Every request is HMAC-signed (see signer.go) and carries a device ID.
// TLS is enforced — MinVersion TLS 1.2, InsecureSkipVerify is never set.
type WorkerTransport struct {
	baseURL  string
	apiKey   string
	deviceID string
	client   *http.Client
	errOut   io.Writer // injectable for tests; defaults to os.Stderr
}

func NewWorker(baseURL, apiKey, deviceID string) *WorkerTransport {
	if baseURL == "" {
		baseURL = defaultWorkerURL
	}
	return &WorkerTransport{
		baseURL:  baseURL,
		apiKey:   apiKey,
		deviceID: deviceID,
		errOut:   os.Stderr,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				// Warm, multiplexed HTTP/2 — the daemon is long-lived.
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          4,
				MaxIdleConnsPerHost:   4,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 8 * time.Second,
			},
		},
	}
}

func (t *WorkerTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	body, err := json.Marshal(optimizeRequest{Event: event, APIKey: t.apiKey})
	if err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, err
	}

	req, err := http.NewRequest(http.MethodPost, t.baseURL+"/optimize", bytes.NewReader(body))
	if err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	// HMAC signing + device binding — see signer.go for the scheme.
	for k, v := range SignedHeaders(t.apiKey, t.deviceID, body) {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, fmt.Errorf("worker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough},
			fmt.Errorf("worker: HTTP %d", resp.StatusCode)
	}

	var out optimizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, err
	}

	// Surface upgrade nudge to stderr — never blocks the developer.
	if out.Warning != "" {
		fmt.Fprintf(t.errOut, "[confire] %s\n", out.Warning)
	}

	return out.Result, nil
}

func (t *WorkerTransport) Mode() OptimizerMode { return OptimizerModeRemote }
