package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/twistingmercury/mnemonic-integration-tests/helpers"
)

func TestEnrichment_PatternBecomesEnriched(t *testing.T) {
	client := helpers.NewTestClient(t)
	pattern := helpers.CreateEnrichablePattern(t, client)

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	enriched := helpers.PollForEnrichment(t, client, pattern.ID, 60*time.Second)

	if enriched.EnrichmentStatus != "enriched" {
		t.Fatalf("expected enrichment_status 'enriched', got %q", enriched.EnrichmentStatus)
	}
	if enriched.EnrichedAt == "" {
		t.Fatal("expected enriched_at to be set after enrichment")
	}
}

func TestEnrichment_ChunksCreatedAfterEnrichment(t *testing.T) {
	client := helpers.NewTestClient(t)
	pattern := helpers.CreateEnrichablePattern(t, client)

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	helpers.PollForEnrichment(t, client, pattern.ID, 60*time.Second)

	resp, err := client.Get("/v1/api/patterns/" + pattern.ID + "/chunks")
	if err != nil {
		t.Fatalf("GET /v1/api/patterns/%s/chunks: %v", pattern.ID, err)
	}

	helpers.AssertStatusCode(t, resp, http.StatusOK)
	chunks := helpers.ParseJSON[helpers.ChunkListResponse](t, resp)

	if chunks.Count == 0 {
		t.Fatal("expected at least 1 chunk after enrichment, got 0")
	}
	if len(chunks.Chunks) == 0 {
		t.Fatal("expected chunks array to be non-empty")
	}
}
