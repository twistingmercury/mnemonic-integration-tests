.DEFAULT_GOAL := help

.PHONY: help build test start stop analyze

COMPOSE_FILE := docker-compose.yaml

build: ## Build test runner image and run all integration tests
	@printf "\n=== Starting integration test infrastructure ===\n"
	docker compose -f $(COMPOSE_FILE) up -d postgres neo4j rabbitmq
	@printf "\n=== Starting application services ===\n"
	docker compose -f $(COMPOSE_FILE) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from mnemonic_tests \
		mnemonic_api mnemonic_mcp mnemonic_enricher mnemonic_tests
	@$(MAKE) stop

test: build ## Alias for build

start: ## Start the full stack without running tests (for local dev)
	docker compose -f $(COMPOSE_FILE) up -d postgres neo4j rabbitmq
	docker compose -f $(COMPOSE_FILE) up -d mnemonic_api mnemonic_mcp mnemonic_enricher

stop: ## Tear down all containers and volumes
	@printf "Stopping integration test infrastructure..."
	@docker compose -f $(COMPOSE_FILE) down -v --remove-orphans > /dev/null 2>&1 || true
	@docker rmi mnemonic_integration_tests:latest -f > /dev/null 2>&1 || true
	@printf "done\n"

analyze: ## Lint all tracked Markdown files
	markdownlint -c .markdownlint.json $$(git ls-files '*.md')

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
