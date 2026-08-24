# mnemonic-integration-tests

> **Maturity Level**: Emerging — black-box smoke tests for the complete
> Mnemonic stack.
> **Version**: v0.0.2

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
- `make help` — list targets

## How it works

The suite exercises public HTTP interfaces only. Docker Compose starts
PostgreSQL, Neo4j, RabbitMQ, the Mnemonic API, MCP server, and enrichment
worker. The `mnemonic_tests` container waits for the services to become healthy
and runs `go test -v -count=1 ./...`.

The tests cover:

- service health and version endpoints;
- pattern CRUD, filtering, semantic search, and enrichment through the REST API;
- MCP tool discovery, pattern retrieval, search, and related-pattern lookup; and
- the full create → enrich → MCP-search workflow.

The application services use internal Docker DNS names such as `mnemonic_api`
and `mnemonic_mcp`. They are not exposed to the host.

## Key Considerations

- **Prerequisites:** Docker Compose and access to the required service images.
  The integration project builds only the Go test-runner image; the API, MCP,
  enrichment, PostgreSQL, and Neo4j images are pulled or must already be
  available locally.
- **OpenAI API key:** Set `MNEMONIC_OPENAI_API_KEY` to enable enrichment and
  semantic-search coverage. Without a valid key, enrichment-dependent checks
  cannot succeed.
- **Isolation:** The stack publishes no host ports. Compose scopes its
  containers, `mnemonic-network`, and database volumes under the
  `mnemonic-integration-tests` project name, so it can run alongside other
  Mnemonic Compose stacks.
- **Cleanup:** `make stop` removes this project's containers and volumes. It
  also runs `docker system prune -f`, which can remove unused Docker resources
  from other projects; it does not stop running containers.

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

### Testing and Debugging

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

## Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/). Version
is determined from Git tags:

```bash
git describe --tags --always
```

The current release tag is `v0.0.2`.

---

Part of the Mnemonic multi-repository workspace. See [mnemonic-docs](https://github.com/twistingmercury/mnemonic-docs) for
architecture and design documentation.
