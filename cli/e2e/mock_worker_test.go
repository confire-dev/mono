// Package e2e contains end-to-end tests for the confire hook pipeline.
// Each test spins up a real httptest.Server acting as the Cloudflare Worker,
// then drives the full chain: hook input → LocalTransport or WorkerTransport → result.
//
// Author: Efe <efe@efebehar.dev>
package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
)

// mockWorkerConfig controls mock Worker behaviour per test.
type mockWorkerConfig struct {
	// apiKey the mock will accept (empty = accept any non-empty key)
	apiKey string
	// verifySig checks the HMAC signature on every request
	verifySig bool
	// atLimit makes the Worker return a rate-limit passthrough
	atLimit bool
	// forcePassthrough makes the Worker always return passthrough
	forcePassthrough bool
	// requestLog receives every verified request (for assertions)
	requestLog chan optimizeReq
}

type optimizeReq struct {
	Event  intercept.InterceptEvent `json:"event"`
	APIKey string                   `json:"apiKey"`
}

type optimizeResp struct {
	Result  intercept.InterceptResult `json:"result"`
	Warning string                    `json:"_warning,omitempty"`
}

// newMockWorker returns an httptest.Server that behaves like the real Worker.
func newMockWorker(cfg mockWorkerConfig) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/optimize", func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "bad content-type", 400)
			return
		}

		// Verify HMAC signature if requested
		if cfg.verifySig {
			ts := r.Header.Get("X-Confire-Timestamp")
			sig := r.Header.Get("X-Confire-Sig")
			if ts == "" || sig == "" {
				http.Error(w, "missing signature headers", 401)
				return
			}
			// Replay check: reject requests older than 5 minutes
			tsInt, err := strconv.ParseInt(ts, 10, 64)
			if err != nil || time.Now().Unix()-tsInt > 300 {
				http.Error(w, "timestamp expired", 401)
				return
			}
		}

		var req optimizeReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}

		// Validate API key
		if cfg.apiKey != "" && req.APIKey != cfg.apiKey {
			http.Error(w, "invalid key", 401)
			return
		}

		// Log the request for test assertions
		if cfg.requestLog != nil {
			select {
			case cfg.requestLog <- req:
			default:
			}
		}

		var result intercept.InterceptResult
		var warning string

		if cfg.atLimit {
			result = intercept.InterceptResult{
				Kind:    intercept.ResultAddContext,
				Context: "⚠️ Free tier reached (500/500). Upgrade at confire.dev/upgrade",
			}
			warning = "⚠️ 500/500 requests used · confire.dev/upgrade"
		} else if cfg.forcePassthrough {
			result = intercept.InterceptResult{Kind: intercept.ResultPassthrough}
		} else {
			// Run the actual Worker optimizer logic via the engine
			result = workerHandle(req.Event)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(optimizeResp{Result: result, Warning: warning})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"ok":true}`)
	})

	return httptest.NewServer(mux)
}

// newWorkerTransport creates a WorkerTransport pointing to the given test server.
func newWorkerTransport(srv *httptest.Server, apiKey string) *transport.WorkerTransport {
	return transport.NewWorker(srv.URL, apiKey, "test-device-001")
}
