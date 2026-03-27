// Package integration_tests provides end-to-end integration smoke tests for
// the full Mnemonic stack. Tests in this package require a live stack (API,
// MCP, enricher, postgres, neo4j, rabbitmq) and treat all services as black
// boxes, exercising only their public HTTP interfaces.
//
// Run with:
//
//	go test -v -count=1 ./...
package integration_tests
