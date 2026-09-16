# mnemonic-integration-tests

> **Maturity Level**: Emerging — black-box smoke tests for the complete
> Mnemonic stack.
> **Version**: v0.0.3

---

## Table of Contents

- [Usage](#usage)
- [How it works](#how-it-works)
- [Key Considerations](#key-considerations)
- [Continuous Integration](#continuous-integration)
- [Development Considerations](#development-considerations)
- [Versioning](#versioning)

## Usage

Run the complete suite in a self-contained Docker Compose stack:

```bash
# Required for enrichment and semantic-search coverage.
export MNEMONIC_OPENAI_API_KEY="sk-..."
make build
```

For local investigation, start the application stack and run the test runner separately:

```bash
make start
docker compose -f docker-compose.yaml up --build mnemonic_tests
make stop
```

Available targets:

- `make build` — start the stack, run the suite, then tear it down
- `make test` — alias for `make build`
- `make start` — start infrastructure and application services without tests
- `make stop` — remove this Compose stack and its volumes
- `make analyze` — lint all tracked Markdown files
- `make help` — list targets (default)

## How it works

The suite exercises public HTTP interfaces only. Docker Compose starts
PostgreSQL, Neo4j, RabbitMQ, the Mnemonic API, MCP server, and enrichment
worker. The `mnemonic_tests` container starts once the API, MCP, and enrichment
services report healthy, polls `/health` on the API, then runs
`go test -v -count=1 ./...`.

The tests cover:

- service health and version endpoints;
- pattern CRUD, listing with pagination, semantic search, and enrichment
  through the REST API;
- MCP tool discovery, pattern retrieval, search, and related-pattern lookup; and
- the full create → enrich → MCP-search workflow.

The application services use internal Docker DNS names such as `mnemonic_api`
and `mnemonic_mcp`. They are not exposed to the host.

## Key Considerations

- **Prerequisites:** Docker Compose and access to the required service images.
  The integration project builds only the Go test-runner image; the API, MCP,
  enrichment, PostgreSQL, Neo4j, and RabbitMQ images are pulled or must already
  be available locally.
- **Go toolchain:** The test runner uses Go 1.27.1 from
  `ghcr.io/twistingmercury/golang-tooling:go1.27.1`. Docker supplies Go for
  Compose runs; a local Go 1.27.1 or newer toolchain is needed only when running
  tests directly outside Docker.
- **OpenAI API key:** Set `MNEMONIC_OPENAI_API_KEY` to enable enrichment and
  semantic-search coverage. Without a valid key, enrichment-dependent checks
  cannot succeed.
- **Isolation:** The stack publishes no host ports. Compose scopes its
  containers, `mnemonic-network`, and database volumes under the
  `mnemonic-integration-tests` project name, so it can run alongside other
  Mnemonic Compose stacks.
- **Cleanup:** `make stop` removes this Compose project's containers (including
  orphans), networks, and volumes. It also force-removes the local
  `mnemonic_integration_tests:latest` test-runner image.

## Continuous Integration

GitHub Actions runs the complete `make build` suite for every push to
`develop`, including merged pull requests. It also runs daily at midnight UTC
and can be started manually.

On failure, the workflow captures Compose service status and logs before
tearing down the stack. CI requires the `MNEMONIC_OPENAI_API_KEY` repository
secret for enrichment and semantic-search coverage.

## Development Considerations

### Quick Start

```bash
docker images | grep mnemonic
export MNEMONIC_OPENAI_API_KEY="sk-..."
make build
```

### Building & running

See [Usage](#usage) for the Docker Compose build and run commands.

### Testing

Run a standing stack, inspect logs or databases, then stop it:

```bash
make start
docker compose -f docker-compose.yaml logs mnemonic_api mnemonic_enricher
docker compose -f docker-compose.yaml exec postgres psql -U mnemonic -d mnemonic
docker compose -f docker-compose.yaml exec neo4j cypher-shell -u neo4j -p mnemonic_dev
make stop
```

To run just one package while the stack is running, start a one-off test
container with the package passed to Go:

```bash
make start
docker compose -f docker-compose.yaml run --rm --entrypoint go mnemonic_tests \
  test -v ./health
make stop
```

Tests and shared HTTP/MCP helpers live under `src/`; add new coverage in the
appropriate `src/{api,health,mcp,pipeline}/` package.

### Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/). Version
is determined from Git tags:

```bash
git describe --tags --always
```

The current release tag is `v0.0.3`.

---

Part of the Mnemonic multi-repository workspace. See
[mnemonic-docs](https://github.com/twistingmercury/mnemonic-docs) for
architecture and design documentation.
