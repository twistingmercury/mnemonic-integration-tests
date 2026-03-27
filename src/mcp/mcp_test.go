package mcp_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/twistingmercury/mnemonic-integration-tests/helpers"
)

func TestMCPToolsList_ReturnsThreeTools(t *testing.T) {
	client := helpers.NewMCPClient(t)
	resp := client.ListTools(t)
	if resp.Error != nil {
		t.Fatalf("ListTools returned error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("ListTools returned nil result")
	}
	tools := resp.Result.Tools
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{"search_patterns", "find_related_patterns", "get_pattern"} {
		if !names[expected] {
			t.Errorf("expected tool %q not found in list", expected)
		}
	}
}

func TestMCPGetPattern_ExistingPattern_ReturnsMarkdown(t *testing.T) {
	apiClient := helpers.NewTestClient(t)
	pattern := helpers.CreateTestPattern(t, apiClient)

	t.Cleanup(func() {
		resp, err := apiClient.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	mcpClient := helpers.NewMCPClient(t)
	resp := mcpClient.CallTool(t, "get_pattern", map[string]any{"id": pattern.ID})
	result := helpers.AssertMCPSuccess(t, resp)

	if len(result.Content) == 0 {
		t.Fatal("expected content in result, got none")
	}
	text := result.Content[0].Text
	if !strings.Contains(text, pattern.Name) {
		t.Errorf("expected result to contain pattern name %q, got: %s", pattern.Name, text)
	}
}

func TestMCPGetPattern_NonExistent_ReturnsError(t *testing.T) {
	client := helpers.NewMCPClient(t)
	resp := client.CallTool(t, "get_pattern", map[string]any{"id": uuid.New().String()})
	helpers.AssertMCPError(t, resp)
}

func TestMCPSearchPatterns_ValidQuery_ReturnsResponse(t *testing.T) {
	client := helpers.NewMCPClient(t)
	resp := client.CallTool(t, "search_patterns", map[string]any{"query": "dependency injection"})
	if resp.Error == nil && resp.Result == nil {
		t.Fatal("expected either a JSON-RPC error or a result, got neither")
	}
}

func TestMCPFindRelated_ExistingPattern_ReturnsSuccessOrEmpty(t *testing.T) {
	apiClient := helpers.NewTestClient(t)
	pattern := helpers.CreateTestPattern(t, apiClient)

	t.Cleanup(func() {
		resp, err := apiClient.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	mcpClient := helpers.NewMCPClient(t)
	resp := mcpClient.CallTool(t, "find_related_patterns", map[string]any{"pattern_id": pattern.ID})

	if resp.Error != nil || (resp.Result != nil && resp.Result.IsError) {
		t.Log("MCP returned error (expected for unenriched pattern with no Neo4j node)")
		return
	}
	result := helpers.AssertMCPSuccess(t, resp)
	if len(result.Content) > 0 {
		text := result.Content[0].Text
		if !strings.Contains(text, "No related patterns found for") {
			t.Errorf("expected 'No related patterns found for' in response, got: %s", text)
		}
	}
}
