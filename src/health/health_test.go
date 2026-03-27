package health_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/twistingmercury/mnemonic-integration-tests/helpers"
)

func TestAPIHealth_Returns200(t *testing.T) {
	client := helpers.NewTestClient(t)

	resp, err := client.Get("/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	helpers.AssertStatusCode(t, resp, http.StatusOK)

	body := helpers.ParseJSON[map[string]any](t, resp)
	status, ok := body["status"].(string)
	if !ok {
		t.Fatalf("expected 'status' field to be a string, got %T", body["status"])
	}
	if status != "OK" && status != "NotSet" {
		t.Fatalf("expected status 'OK' or 'NotSet', got %q", status)
	}
}

func TestAPIVersion_Returns200(t *testing.T) {
	client := helpers.NewTestClient(t)

	resp, err := client.Get("/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	defer resp.Body.Close()

	helpers.AssertStatusCode(t, resp, http.StatusOK)

	ver := helpers.ParseJSON[helpers.VersionResponse](t, resp)
	if ver.Version == "" {
		t.Fatal("expected version field to be non-empty")
	}
}

func TestMCPHealth_Returns200(t *testing.T) {
	mcpURL := helpers.MCPOpsBaseURL()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get(mcpURL + "/health")
	if err != nil {
		t.Fatalf("GET %s/health: %v", mcpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMCPVersion_Returns200(t *testing.T) {
	mcpURL := helpers.MCPOpsBaseURL()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get(mcpURL + "/version")
	if err != nil {
		t.Fatalf("GET %s/version: %v", mcpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := helpers.ReadBody(t, resp)
	if !strings.Contains(string(body), "version") {
		t.Fatalf("expected version field in response body: %s", string(body))
	}
}

func TestEnricherHealth_Returns200(t *testing.T) {
	enricherURL := helpers.EnricherBaseURL()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get(enricherURL + "/health")
	if err != nil {
		t.Fatalf("GET %s/health: %v", enricherURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestEnricherVersion_Returns200(t *testing.T) {
	enricherURL := helpers.EnricherBaseURL()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get(enricherURL + "/version")
	if err != nil {
		t.Fatalf("GET %s/version: %v", enricherURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := helpers.ReadBody(t, resp)
	if !strings.Contains(string(body), "version") {
		t.Fatalf("expected version field in response body: %s", string(body))
	}
}
