#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://mnemonic_api:8080}"
MCP_URL="${MCP_URL:-http://mnemonic_mcp:8081}"
ENRICHER_URL="${ENRICHER_URL:-http://mnemonic_enricher:8080}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_INTERVAL="${RETRY_INTERVAL:-2}"

echo "=== Integration Test Runner ==="
echo "API URL:      ${API_URL}"
echo "MCP URL:      ${MCP_URL}"
echo "Enricher URL: ${ENRICHER_URL}"
echo ""

wait_for_service() {
    local name="$1"
    local url="$2"

    echo "Waiting for ${name} to be ready..."
    for i in $(seq 1 "${MAX_RETRIES}"); do
        if curl -sf "${url}/health" > /dev/null 2>&1; then
            echo "${name} is ready (attempt ${i}/${MAX_RETRIES})"
            return 0
        fi
        if [ "$i" -eq "${MAX_RETRIES}" ]; then
            echo "ERROR: ${name} failed to become ready after ${MAX_RETRIES} attempts"
            return 1
        fi
        echo "Waiting for ${name}... (${i}/${MAX_RETRIES})"
        sleep "${RETRY_INTERVAL}"
    done
}

wait_for_service "API" "${API_URL}"
wait_for_service "MCP" "${MCP_URL}"
wait_for_service "Enricher" "${ENRICHER_URL}"

echo ""
echo "All services ready. Running integration tests..."
echo ""

cd /e2e
stdbuf -oL go test -v -count=1 ./...

echo ""
echo "=== Integration Tests Complete ==="
