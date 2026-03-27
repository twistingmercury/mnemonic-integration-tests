# mnemonic-integration-tests

> **Maturity Level**: Emerging — black-box integration smoke tests for the full Mnemonic stack
> **Version**: v0.0.1

---

## Table of Contents

- [Usage](#usage)
- [How it works](#how-it-works)
- [Key Considerations](#key-considerations)
- [Development Considerations](#development-considerations)
- [Versioning](#versioning)

## Usage

Run the full integration test suite against a self-contained Docker Compose stack:

```bash
make build
```

Or run tests without tearing down the stack (for local development):

```bash
make start
# Stack is now running. Run tests manually or rebuild the test image:
docker compose up --build mnemonic_tests
make stop  # when done
```

Available Makefile targets:

- `make build` — Start infrastructure, build test runner, run all tests, tear down stack
- `make test` — Alias for build
- `make start` — Start the full stack without running tests
- `make stop` — Tear down all containers and volumes
- `make help` — Show available targets

### Environment Variables

- `MNEMONIC_OPENAI_API_KEY` (required for enrichment tests) — OpenAI API key for pattern enrichment and embedding search. Without this, enrichment will fail with a dummy key and search-related tests will skip.
- `API_URL` (default: http://mnemonic_api:8080) — Mnemonic API endpoint (used inside Docker network)
- `MCP_URL` (default: http://mnemonic_mcp:8081) — MCP JSON-RPC endpoint for `tools/list`, `get_pattern`, `search_patterns`, `find_related_patterns`
- `MCP_OPS_URL` (default: http://mnemonic_mcp:8080) — MCP health and version operations
- `ENRICHER_URL` (default: http://mnemonic_enricher:8080) — Enricher service health endpoint

## How it works

The test suite is a collection of black-box integration tests that exercise the full Mnemonic stack as an integrated system. All tests use only public HTTP interfaces.

### Stack Architecture

The Docker Compose stack (`docker-compose.yaml`) brings up seven services:

| Service | Image | Role |
|---------|-------|------|
| postgres | ghcr.io/twistingmercury/mnemonic-postgres:v1.0.0-dev | PostgreSQL with Mnemonic schema |
| neo4j | ghcr.io/twistingmercury/mnemonic-neo4j:v1.0.0-dev | Neo4j graph database |
| rabbitmq | rabbitmq:3-management-alpine | Message queue for enrichment jobs |
| mnemonic_api | ghcr.io/twistingmercury/mnemonic-api:latest | REST API for pattern CRUD and search |
| mnemonic_mcp | ghcr.io/twistingmercury/mnemonic:latest | MCP server exposing pattern tools |
| mnemonic_enricher | ghcr.io/twistingmercury/mnemonic-enrichment:latest | Async enrichment worker (RabbitMQ-driven) |
| mnemonic_tests | ./src/Dockerfile | Test runner container |

Services depend on healthy databases and message queue before starting. The test runner waits for all application services to be healthy before executing tests.

### Test Suites

Tests are organized under `src/` in four packages:

**health/** (6 tests)
- Validates health and version endpoints for API, MCP, and enricher services
- Verifies that all services are reachable and responding

**api/** (8 tests)
- Pattern CRUD operations via REST API (create, get, update, list, delete)
- Search functionality (vector search, tag filtering, domain filtering)
- Enrichment pipeline (create pattern → enrich → verify chunks populated)

**mcp/** (5 tests)
- MCP `tools/list` — lists available MCP tools
- MCP `get_pattern` — fetches a pattern by ID
- MCP `search_patterns` — semantic search over enriched patterns
- MCP `find_related_patterns` — graph traversal to find related patterns

**pipeline/** (1 test)
- End-to-end centerpiece test: create pattern via API → wait for enrichment → verify MCP search finds it
- Validates the complete flow from pattern creation through enrichment to searchability

### Test Execution Flow

1. Docker Compose starts infrastructure (postgres, neo4j, rabbitmq)
2. Application services (api, mcp, enricher) start and wait for healthy databases
3. Test runner container starts and waits for all services to be healthy
4. Test runner executes all test packages with `go test ./...`
5. On exit, `make build` tears down the entire stack

Tests run inside the Docker network, so they use internal hostnames (e.g., `http://mnemonic_api:8080`).

## Key Considerations

- **Service Images Required**: All service container images (postgres, neo4j, api, mcp, enricher) must already exist locally or in GHCR before running tests. This project does not build them — it only orchestrates them.

- **OpenAI API Key**: Enrichment and search tests require a valid `MNEMONIC_OPENAI_API_KEY` environment variable. Without a valid key, the enrichment worker will fail to generate embeddings, and tests that depend on semantic search will skip or fail. Set this before running:
  ```bash
  export MNEMONIC_OPENAI_API_KEY="sk-..."
  make build
  ```

- **Black-Box Testing**: Tests only exercise public HTTP interfaces. They do not mock internal services or databases — they validate the entire integrated system end-to-end.

- **Stateful Stack**: The Docker Compose stack maintains state across test runs (volumes persist). Use `make stop` to fully clean up databases and volumes between runs if you need a fresh state.

- **Port Isolation**: Inside the Docker network, services use their internal hostnames and fixed ports. Service ports are not exposed to the host.

## Development Considerations

### Quick Start

1. Ensure all service images exist:
   ```bash
   docker images | grep mnemonic
   ```

2. Set your OpenAI API key:
   ```bash
   export MNEMONIC_OPENAI_API_KEY="sk-..."
   ```

3. Run tests:
   ```bash
   make build
   ```

### Testing

Run all tests:

```bash
make build
```

Run tests against a standing stack (without teardown):

```bash
# Start stack
make start

# Run tests manually
cd src && go test -v ./...

# Or rebuild and run test container
docker compose up --build mnemonic_tests

# Teardown when done
make stop
```

Run a specific test suite:

```bash
make start
cd src && go test -v ./health
make stop
```

Run a single test:

```bash
make start
cd src && go test -v ./pipeline -run TestPipelineCreateEnrichMCPSearch
make stop
```

### Debugging

To inspect the running stack:

```bash
make start

# Check service logs
docker compose logs mnemonic_api
docker compose logs mnemonic_enricher

# Inspect test container
docker compose exec mnemonic_tests sh

# Query databases
docker compose exec postgres psql -U mnemonic -d mnemonic
docker compose exec neo4j cypher-shell -u neo4j -p mnemonic_dev

make stop
```

### Adding Tests

1. Create a test file under `src/{suite}/` (e.g., `src/api/new_test.go`)
2. Use helpers from `src/helpers/` for HTTP clients, assertions, and polling
3. Run with `go test -v ./...` against a standing stack
4. Tests will run automatically when `make build` is invoked

Test helpers available:

- `helpers.NewTestClient(t)` — HTTP client preconfigured with `API_URL`
- `helpers.NewMCPClient(t)` — MCP JSON-RPC client preconfigured with `MCP_URL`
- `helpers.AssertStatusCode(t, resp, code)` — Assert HTTP response status
- `helpers.ParseJSON[T](t, resp)` — Parse JSON response body
- `helpers.PollForEnrichment(t, client, id, timeout)` — Poll until pattern enrichment completes
- `helpers.CreateTestPattern(t, client)` — Create a minimal test pattern via API
- `helpers.CreateEnrichablePattern(t, client)` — Create a pattern with enrichable content

### Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/).

Version is determined from git tags:

```bash
git describe --tags --always
```

Currently at **v0.0.1** (development).

---

**Part of the Mnemonic multi-repo workspace.** See [mnemonic-docs](../mnemonic-docs) for architecture and design documentation.
