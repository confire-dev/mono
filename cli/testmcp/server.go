// Package testmcp provides an in-process MCP server for use in tests.
//
// The server speaks the MCP Streamable HTTP transport (JSON-RPC 2.0 over HTTP POST).
// It is backed by a configurable set of Tool definitions whose handlers return
// whatever the test needs — clean output, secret-laden output, injected output, etc.
//
// Usage:
//
//	srv := testmcp.New(testmcp.WithTool(testmcp.EchoTool), testmcp.WithTool(testmcp.SecretTool))
//	defer srv.Close()
//
//	// The server URL is srv.URL — pass it to your MCP client or intercept pipeline.
//	// Tool calls are recorded and can be inspected via srv.Calls().
package testmcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// ── JSON-RPC types ────────────────────────────────────────────────────────────

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ── MCP types ─────────────────────────────────────────────────────────────────

// ToolDef describes an MCP tool registered with the server.
type ToolDef struct {
	// Name is the tool name (e.g. "get_data").
	Name string
	// Description is the human-readable description included in tools/list.
	Description string
	// InputSchema is the JSON Schema for the tool's input parameters (any → marshalled).
	InputSchema any
	// Handler is called when the tool is invoked. It receives the raw params JSON
	// and returns the tool result content and an optional error.
	Handler func(params map[string]any) (any, error)
}

// Call records one tools/call invocation.
type Call struct {
	ToolName string
	Params   map[string]any
	Result   any
	Err      error
}

// ── Server ─────────────────────────────────────────────────────────────────────

// Server is a test MCP server backed by httptest.Server.
type Server struct {
	*httptest.Server
	tools map[string]ToolDef

	mu    sync.Mutex
	calls []Call
}

// Option configures a Server.
type Option func(*Server)

// WithTool registers a tool with the server.
func WithTool(t ToolDef) Option {
	return func(s *Server) {
		s.tools[t.Name] = t
	}
}

// New creates and starts a new MCP test server.
func New(opts ...Option) *Server {
	s := &Server{tools: make(map[string]ToolDef)}
	for _, o := range opts {
		o(s)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	s.Server = httptest.NewServer(mux)
	return s
}

// Calls returns all recorded tool invocations in order.
func (s *Server) Calls() []Call {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Call, len(s.calls))
	copy(out, s.calls)
	return out
}

// ResetCalls clears the recorded call log.
func (s *Server) ResetCalls() {
	s.mu.Lock()
	s.calls = nil
	s.mu.Unlock()
}

// handleMCP is the single HTTP endpoint for MCP Streamable HTTP transport.
// All JSON-RPC messages arrive as POST /mcp with Content-Type: application/json.
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, nil, -32700, "parse error")
		return
	}
	if req.JSONRPC != "2.0" {
		writeError(w, req.ID, -32600, "invalid JSON-RPC version")
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req)
	case "notifications/initialized":
		// Client notification — no response needed.
		w.WriteHeader(http.StatusAccepted)
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolsCall(w, req)
	case "ping":
		writeResult(w, req.ID, map[string]any{})
	default:
		writeError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(w http.ResponseWriter, req jsonRPCRequest) {
	writeResult(w, req.ID, map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "testmcp",
			"version": "0.0.1",
		},
	})
}

func (s *Server) handleToolsList(w http.ResponseWriter, req jsonRPCRequest) {
	tools := make([]any, 0, len(s.tools))
	for _, t := range s.tools {
		schema := t.InputSchema
		if schema == nil {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		tools = append(tools, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": schema,
		})
	}
	writeResult(w, req.ID, map[string]any{"tools": tools})
}

func (s *Server) handleToolsCall(w http.ResponseWriter, req jsonRPCRequest) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		writeError(w, req.ID, -32602, "invalid params")
		return
	}

	tool, ok := s.tools[p.Name]
	if !ok {
		writeError(w, req.ID, -32602, fmt.Sprintf("unknown tool: %s", p.Name))
		return
	}

	result, err := tool.Handler(p.Arguments)

	s.mu.Lock()
	s.calls = append(s.calls, Call{ToolName: p.Name, Params: p.Arguments, Result: result, Err: err})
	s.mu.Unlock()

	if err != nil {
		writeResult(w, req.ID, map[string]any{
			"content": []any{map[string]any{"type": "text", "text": err.Error()}},
			"isError": true,
		})
		return
	}

	// Wrap plain values in the MCP content envelope.
	var content any
	switch v := result.(type) {
	case string:
		content = []any{map[string]any{"type": "text", "text": v}}
	case []any:
		content = v
	default:
		b, _ := json.Marshal(v)
		content = []any{map[string]any{"type": "text", "text": string(b)}}
	}
	writeResult(w, req.ID, map[string]any{"content": content})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeResult(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func writeError(w http.ResponseWriter, id any, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: &jsonRPCError{Code: code, Message: msg}})
}
