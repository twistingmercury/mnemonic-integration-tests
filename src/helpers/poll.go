package helpers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// PollForEnrichment polls GET /v1/api/patterns/:id until enrichment_status is
// "enriched" or the timeout elapses. Returns the final pattern.
func PollForEnrichment(t *testing.T, client *TestClient, id string, timeout time.Duration) Pattern {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get("/v1/api/patterns/" + id)
		if err != nil {
			t.Fatalf("PollForEnrichment GET /v1/api/patterns/%s: %v", id, err)
		}
		body := ReadBody(t, resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("PollForEnrichment GET /v1/api/patterns/%s: status %d\nbody: %s",
				id, resp.StatusCode, string(body))
		}
		var p Pattern
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("PollForEnrichment: unmarshal: %v\nbody: %s", err, string(body))
		}
		switch p.EnrichmentStatus {
		case "enriched":
			return p
		case "failed":
			t.Fatalf("pattern %s enrichment failed: %s", id, p.EnrichmentError)
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("timed out after %s waiting for pattern %s enrichment", timeout, id)
	return Pattern{}
}

// EnrichablePatternContent returns pattern content with the [//]: pattern
// decorator required for the enricher to chunk and index it.
func EnrichablePatternContent() string {
	return "## Overview\n\nThis pattern demonstrates idiomatic Go error wrapping using fmt.Errorf and errors.As.\n\n[//]: pattern\n\n## Core Approach\n\nUse fmt.Errorf with the %w verb to wrap errors and preserve the original cause.\nCall errors.As or errors.Is to unwrap and inspect errors without coupling to\nspecific error string representations.\n\n## Example\n\n```go\nif err := doWork(); err != nil {\n    return fmt.Errorf(\"doWork failed: %w\", err)\n}\n```\n\nCallers can recover the original error type with errors.As."
}

// CreateEnrichablePattern creates a pattern with enrichable content and returns it.
// Callers own cleanup.
func CreateEnrichablePattern(t *testing.T, client *TestClient) Pattern {
	t.Helper()
	body := PatternCreate{
		Name:        GenerateUniqueName("integ-enrich"),
		Description: "Integration test pattern for Go error handling",
		Content:     EnrichablePatternContent(),
		Tags:        []string{"go", "errors", "integration-test"},
		EntityType:  "go-pattern",
		Language:    "go",
		Domain:      "backend",
	}
	resp, err := client.Post("/v1/api/patterns", body)
	if err != nil {
		t.Fatalf("CreateEnrichablePattern: POST /v1/api/patterns: %v", err)
	}
	AssertStatusCode(t, resp, http.StatusAccepted)
	return ParseJSON[Pattern](t, resp)
}
