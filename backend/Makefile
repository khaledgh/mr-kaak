# Developer task runner. On Windows, run targets via `make <target>` (Git Bash
# / WSL) or copy the underlying command. All commands assume CWD = backend/.

BINARY_DIR := bin

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: build
build: ## Build api, worker, migrate binaries into bin/
	go build -o $(BINARY_DIR)/api ./cmd/api
	go build -o $(BINARY_DIR)/worker ./cmd/worker
	go build -o $(BINARY_DIR)/migrate ./cmd/migrate

.PHONY: run
run: ## Run the API server (loads .env)
	go run ./cmd/api

.PHONY: worker
worker: ## Run the background worker
	go run ./cmd/worker

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format all Go files
	gofmt -w .

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	go run ./cmd/migrate up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	go run ./cmd/migrate down 1

.PHONY: migrate-version
migrate-version: ## Print current migration version
	go run ./cmd/migrate version

.PHONY: infra-up
infra-up: ## Start MySQL, Redis, Meilisearch via docker compose
	docker compose up -d mysql redis meilisearch

.PHONY: infra-down
infra-down: ## Stop the infra containers
	docker compose down

.PHONY: check
check: fmt vet test ## Format, vet, and test
