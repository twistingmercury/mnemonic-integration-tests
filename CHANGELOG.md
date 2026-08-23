# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Documentation

- Updated project documentation to reflect the current release version.

## [0.0.2] - 2026-08-23

### Added

- Black-box Docker Compose integration coverage for service health, REST API,
  MCP tools, enrichment, and the end-to-end create-to-search pipeline.
- A scheduled GitHub Actions workflow for the integration suite.

### Updated

- Use development service image tags published from the `develop` branch.
- Run integration tests for pushes and merged pull requests to `develop`.
- Improve CI diagnostics and ensure Compose resources are removed after every
  workflow run.

### Fixed

- Correct MCP health endpoint ports used by the test runner.
- Align pattern-update assertions with the API's `204 No Content` response.
- Increase Neo4j's CI startup allowance to reduce intermittent readiness
  failures.

## [0.0.1] - 2026-03-26

### Initial Release

- Initial release of the Mnemonic black-box integration test project.
- Docker Compose orchestration for the Mnemonic service stack.
- Shared Go helpers for REST API and MCP integration tests.
