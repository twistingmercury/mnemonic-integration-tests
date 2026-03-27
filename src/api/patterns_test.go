package api_test

import (
	"net/http"
	"testing"

	"github.com/twistingmercury/mnemonic-integration-tests/helpers"
)

func TestCreatePattern_Returns202(t *testing.T) {
	client := helpers.NewTestClient(t)
	pattern := helpers.CreateTestPattern(t, client)

	if pattern.ID == "" {
		t.Fatal("expected pattern ID to be non-empty")
	}
	if pattern.EnrichmentStatus != "pending" {
		t.Fatalf("expected enrichment_status 'pending', got %q", pattern.EnrichmentStatus)
	}

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + pattern.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})
}

func TestGetPattern_ReturnsCreatedPattern(t *testing.T) {
	client := helpers.NewTestClient(t)
	created := helpers.CreateTestPattern(t, client)

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + created.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	resp, err := client.Get("/v1/api/patterns/" + created.ID)
	if err != nil {
		t.Fatalf("GET /v1/api/patterns/%s: %v", created.ID, err)
	}

	helpers.AssertStatusCode(t, resp, http.StatusOK)
	got := helpers.ParseJSON[helpers.Pattern](t, resp)

	if got.ID != created.ID {
		t.Errorf("ID mismatch: want %s, got %s", created.ID, got.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Name mismatch: want %q, got %q", created.Name, got.Name)
	}
}

func TestUpdatePattern_Returns200(t *testing.T) {
	client := helpers.NewTestClient(t)
	created := helpers.CreateTestPattern(t, client)

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + created.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	update := helpers.PatternUpdate{
		Name:       created.Name,
		Content:    "# Updated\n\nUpdated content.",
		EntityType: "go-pattern",
		Language:   "go",
		Domain:     "backend",
	}
	resp, err := client.Put("/v1/api/patterns/"+created.ID, update)
	if err != nil {
		t.Fatalf("PUT /v1/api/patterns/%s: %v", created.ID, err)
	}
	defer resp.Body.Close()

	// API returns 204 No Content on successful update.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body := helpers.ReadBody(t, resp)
		t.Fatalf("expected 204 or 200, got %d\nBody: %s", resp.StatusCode, string(body))
	}

	// Verify by fetching the updated pattern.
	getResp, err := client.Get("/v1/api/patterns/" + created.ID)
	if err != nil {
		t.Fatalf("GET after PUT: %v", err)
	}
	helpers.AssertStatusCode(t, getResp, http.StatusOK)
	updated := helpers.ParseJSON[helpers.Pattern](t, getResp)

	if updated.Content != "# Updated\n\nUpdated content." {
		t.Errorf("content not updated: got %q", updated.Content)
	}
}

func TestListPatterns_ReturnsResults(t *testing.T) {
	client := helpers.NewTestClient(t)
	created := helpers.CreateTestPattern(t, client)

	t.Cleanup(func() {
		resp, err := client.Delete("/v1/api/patterns/" + created.ID)
		if err != nil {
			t.Logf("cleanup DELETE: %v", err)
			return
		}
		resp.Body.Close()
	})

	resp, err := client.Get("/v1/api/patterns")
	if err != nil {
		t.Fatalf("GET /v1/api/patterns: %v", err)
	}

	helpers.AssertStatusCode(t, resp, http.StatusOK)
	list := helpers.ParseJSON[helpers.PatternList](t, resp)

	if list.Data == nil {
		t.Fatal("expected data field to be present")
	}
	if list.Pagination.Limit <= 0 {
		t.Fatalf("expected pagination.limit > 0, got %d", list.Pagination.Limit)
	}
}

func TestSearchPatterns_ReturnsResponse(t *testing.T) {
	client := helpers.NewTestClient(t)

	resp, err := client.Get("/v1/api/patterns/search?q=dependency+injection")
	if err != nil {
		t.Fatalf("GET /v1/api/patterns/search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		t.Skip("search unavailable: embedding service not configured")
	}

	helpers.AssertStatusCode(t, resp, http.StatusOK)
	sr := helpers.ParseJSON[helpers.PatternSearchResponse](t, resp)

	if sr.Metadata.Query == "" {
		t.Fatal("expected metadata.query to be non-empty")
	}
}

func TestDeletePattern_Returns204(t *testing.T) {
	client := helpers.NewTestClient(t)
	created := helpers.CreateTestPattern(t, client)

	resp, err := client.Delete("/v1/api/patterns/" + created.ID)
	if err != nil {
		t.Fatalf("DELETE /v1/api/patterns/%s: %v", created.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body := helpers.ReadBody(t, resp)
		t.Fatalf("expected 204 or 200, got %d\nBody: %s", resp.StatusCode, string(body))
	}

	getResp, err := client.Get("/v1/api/patterns/" + created.ID)
	if err != nil {
		t.Fatalf("GET after DELETE: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}
