package testmcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// Client is a minimal MCP client for use in tests.
// It speaks MCP Streamable HTTP transport (JSON-RPC 2.0 over HTTP POST).
type Client struct {
	baseURL    string
	httpClient *http.Client
	nextID     atomic.Int64
}

// NewClient creates a client pointed at the given MCP server URL.
func NewClient(serverURL string) *Client {
	return &Client{
		baseURL:    serverURL + "/mcp",
		httpClient: &http.Client{},
	}
}

// Initialize sends the MCP initialize handshake.
func (c *Client) Initialize() error {
	_, err := c.call("initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "testclient", "version": "0.0.1"},
	})
	if err != nil {
		return err
	}
	// Send notifications/initialized (fire-and-forget, ignore error)
	_ = c.notify("notifications/initialized", map[string]any{})
	return nil
}

// ListTools returns the tools advertised by the server.
func (c *Client) ListTools() ([]map[string]any, error) {
	result, err := c.call("tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	m, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	raw, _ := json.Marshal(m["tools"])
	var tools []map[string]any
	json.Unmarshal(raw, &tools)
	return tools, nil
}

// CallTool invokes a tool by name with the given arguments.
// Returns the raw result map from the server.
func (c *Client) CallTool(name string, args map[string]any) (map[string]any, error) {
	result, err := c.call("tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return nil, err
	}
	m, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	return m, nil
}

// CallToolText calls a tool and returns the concatenated text content.
func (c *Client) CallToolText(name string, args map[string]any) (string, error) {
	result, err := c.CallTool(name, args)
	if err != nil {
		return "", err
	}
	return extractText(result), nil
}

// ── internal ──────────────────────────────────────────────────────────────────

func (c *Client) call(method string, params any) (any, error) {
	id := c.nextID.Add(1)
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		req.Params = b
	}

	body, _ := json.Marshal(req)
	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(raw, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

func (c *Client) notify(method string, params any) error {
	req := map[string]any{"jsonrpc": "2.0", "method": method}
	if params != nil {
		req["params"] = params
	}
	body, _ := json.Marshal(req)
	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// extractText pulls the text content from a tools/call result envelope.
func extractText(result map[string]any) string {
	raw, _ := json.Marshal(result["content"])
	var items []map[string]any
	json.Unmarshal(raw, &items)
	var out string
	for _, item := range items {
		if t, ok := item["text"].(string); ok {
			out += t
		}
	}
	return out
}
