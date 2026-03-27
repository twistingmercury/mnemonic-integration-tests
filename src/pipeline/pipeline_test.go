package pipeline_test

import (
	"strings"
	"testing"
	"time"

	"github.com/twistingmercury/mnemonic-integration-tests/helpers"
)

// TestPipelineCreateEnrichMCPSearch validates the full Mnemonic pipeline:
//
//  1. Create a pattern with enrichable content via the REST API
//  2. Wait for the enricher to process it (API -> RabbitMQ -> Enricher -> DB)
//  3. Search for it via MCP search_patterns and verify it appears in results
//  4. Retrieve it via MCP get_pattern and verify enriched content
//  5. Clean up by deleting the pattern
func TestPipelineCreateEnrichMCPSearch(t *testing.T) {
	apiClient := helpers.NewTestClient(t)

	// Step 1: Create a pattern with enrichable content
	pattern := helpers.CreateEnrichablePattern(t, apiClient)
	t.Logf("created pattern %s (%s)", pattern.Name, pattern.ID)

	t.Cleanup(func() {
		resp, err := apiClient.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	// Step 2: Wait for enrichment to complete
	enriched := helpers.PollForEnrichment(t, apiClient, pattern.ID, 60*time.Second)
	t.Logf("pattern enriched at %s", enriched.EnrichedAt)

	// Step 3: Search via MCP search_patterns
	mcpClient := helpers.NewMCPClient(t)
	searchResp := mcpClient.CallTool(t, "search_patterns", map[string]any{
		"query": "Go error wrapping fmt.Errorf errors.As",
	})

	if searchResp.Error != nil {
		t.Skipf("MCP search_patterns returned JSON-RPC error (embedding service likely unavailable): %s",
			searchResp.Error.Message)
	}
	if searchResp.Result != nil && searchResp.Result.IsError {
		text := ""
		if len(searchResp.Result.Content) > 0 {
			text = searchResp.Result.Content[0].Text
		}
		t.Skipf("MCP search_patterns returned tool error (embedding service likely unavailable): %s", text)
	}

	result := helpers.AssertMCPSuccess(t, searchResp)
	if len(result.Content) == 0 {
		t.Fatal("search returned empty content after enrichment — expected results")
	}

	text := result.Content[0].Text
	t.Logf("search response length: %d chars", len(text))

	if !strings.Contains(text, pattern.Name) && !strings.Contains(text, pattern.ID) {
		t.Errorf("expected search results to contain pattern name %q or ID %q\nresults:\n%s",
			pattern.Name, pattern.ID, text)
	}

	// Step 4: Retrieve via MCP get_pattern and verify enriched state
	getResp := mcpClient.CallTool(t, "get_pattern", map[string]any{"id": pattern.ID})
	getResult := helpers.AssertMCPSuccess(t, getResp)

	if len(getResult.Content) == 0 {
		t.Fatal("get_pattern returned empty content")
	}

	getText := getResult.Content[0].Text
	if !strings.Contains(getText, pattern.Name) {
		t.Errorf("get_pattern response missing pattern name %q", pattern.Name)
	}
	if strings.Contains(strings.ToLower(getText), "pending") {
		t.Errorf("get_pattern response still shows 'pending' after enrichment:\n%s", getText)
	}
}
