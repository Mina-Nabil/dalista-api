.PHONY: help run dev build test swagger migrate-up migrate-down migrate-status tools tidy

# Load .env for local targets (migrate-*, run). CI passes vars explicitly.
ifneq (,$(wildcard ./.env))
include .env
export
endif

MIGRATION_DIR := ./app/repository/migrations

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

tools: ## Install dev tools (goose, swag, air)
	go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
	go install github.com/swaggo/swag/cmd/swag@v1.16.6
	go install github.com/air-verse/air@latest

tidy: ## Sync go.mod / go.sum
	go mod tidy

swagger: ## Regenerate OpenAPI docs into ./docs
	swag init --parseDependency --parseInternal -g main.go -o ./docs

build: ## Build the server binary
	CGO_ENABLED=0 go build -o ./tmp/server ./main.go

run: ## Run the server (expects a reachable Postgres)
	go run ./main.go

dev: ## Run with live reload (air)
	air

test: ## Run the test suite (serial, real Postgres)
	go test ./test/... -count=1 -p 1 -v

migrate-up: ## Apply all pending migrations
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATION_DIR) up

migrate-down: ## Roll back the most recent migration
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATION_DIR) down

migrate-status: ## Show migration status
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATION_DIR) status
