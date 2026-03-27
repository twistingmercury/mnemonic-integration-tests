package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// MCPBaseURL returns the base URL for the MCP server from environment or default.
func MCPBaseURL() string {
	if url := os.Getenv("MCP_URL"); url != "" {
		return url
	}
	return "http://localhost:8081"
}

// MCPClient wraps an HTTP client for issuing JSON-RPC requests to the MCP endpoint.
type MCPClient struct {
	client  *http.Client
	baseURL string
	nextID  int
}

// NewMCPClient creates a new MCP client pointed at the configured MCP base URL.
func NewMCPClient(t *testing.T) *MCPClient {
	t.Helper()
	return &MCPClient{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: MCPBaseURL(),
		nextID:  1,
	}
}

// Post marshals body as JSON and POSTs it to /mcp, returning the raw response.
func (c *MCPClient) Post(t *testing.T, body any) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("MCPClient: marshal request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/mcp", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("MCPClient: new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatalf("MCPClient: POST /mcp: %v", err)
	}
	return resp
}

// CallTool sends a tools/call JSON-RPC request and returns the decoded response.
func (c *MCPClient) CallTool(t *testing.T, toolName string, arguments any) JSONRPCResponse {
	t.Helper()
	id := c.nextID
	c.nextID++
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params:  ToolCallParams{Name: toolName, Arguments: arguments},
	}
	resp := c.Post(t, req)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("MCPClient.CallTool: read body: %v", err)
	}
	var result JSONRPCResponse
	if err := json.Unmarshal(ExtractSSEData(data), &result); err != nil {
		t.Fatalf("MCPClient.CallTool: unmarshal: %v\nbody: %s", err, data)
	}
	return result
}

// ListTools sends a tools/list JSON-RPC request and returns the decoded response.
func (c *MCPClient) ListTools(t *testing.T) JSONRPCToolsListResp {
	t.Helper()
	id := c.nextID
	c.nextID++
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/list",
		Params:  map[string]any{},
	}
	resp := c.Post(t, req)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("MCPClient.ListTools: read body: %v", err)
	}
	var result JSONRPCToolsListResp
	if err := json.Unmarshal(ExtractSSEData(data), &result); err != nil {
		t.Fatalf("MCPClient.ListTools: unmarshal: %v\nbody: %s", err, data)
	}
	return result
}

// AssertMCPSuccess asserts the response is a non-error MCP tool result and returns it.
func AssertMCPSuccess(t *testing.T, resp JSONRPCResponse) MCPToolResult {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("expected MCP success, got error: code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		t.Fatal("expected MCP result, got nil")
	}
	if resp.Result.IsError {
		text := ""
		if len(resp.Result.Content) > 0 {
			text = resp.Result.Content[0].Text
		}
		t.Fatalf("expected MCP success result, got isError=true: %s", text)
	}
	return *resp.Result
}

// AssertMCPError asserts the response is either a JSON-RPC error or a tool-level error result.
func AssertMCPError(t *testing.T, resp JSONRPCResponse) {
	t.Helper()
	if resp.Error != nil {
		return // JSON-RPC level error — ok
	}
	if resp.Result != nil && resp.Result.IsError {
		return // tool-level error — also ok
	}
	t.Fatal("expected MCP error response, but got success")
}
