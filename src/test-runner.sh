#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://mnemonic_api:8080}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_INTERVAL="${RETRY_INTERVAL:-2}"

echo "=== Integration Test Runner ==="
echo "API URL: ${API_URL}"
echo ""

# MCP and enricher health is gated by docker-compose depends_on/service_healthy.
# We only need to wait for the API here — the other services use their built-in
# --health CLI flag for Docker health checks (no curl available in those containers).

echo "Waiting for API to be ready..."
for i in $(seq 1 "${MAX_RETRIES}"); do
    if curl -sf "${API_URL}/health" > /dev/null 2>&1; then
        echo "API is ready (attempt ${i}/${MAX_RETRIES})"
        break
    fi
    if [ "$i" -eq "${MAX_RETRIES}" ]; then
        echo "ERROR: API failed to become ready after ${MAX_RETRIES} attempts"
        exit 1
    fi
    echo "Waiting for API... (${i}/${MAX_RETRIES})"
    sleep "${RETRY_INTERVAL}"
done

echo ""
echo "All services ready. Running integration tests..."
echo ""

cd /e2e
stdbuf -oL go test -v -count=1 ./...

echo ""
echo "=== Integration Tests Complete ==="
