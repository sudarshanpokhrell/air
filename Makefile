SERVER_DIR := server
BIN_DIR    := $(SERVER_DIR)/bin
DB_URL     ?= postgres://ota:ota@localhost:5432/ota?sslmode=disable

.DEFAULT_GOAL := help
.PHONY: help up down logs psql db-reset run build test vet fmt tidy check \
        migrate-up migrate-down migrate-create clean

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ---------- Docker / Postgres ----------

up: ## Start Postgres (waits until healthy)
	docker compose up -d --wait

down: ## Stop Postgres (data is kept)
	docker compose down

logs: ## Follow Postgres logs
	docker compose logs -f postgres

psql: ## Open a psql shell to the local database
	psql "$(DB_URL)"

db-reset: ## Delete ALL local data and start fresh Postgres
	docker compose down -v
	docker compose up -d --wait

# ---------- Migrations ----------

migrate-up: ## Apply all pending migrations
	cd $(SERVER_DIR) && go run ./cmd/migrate up

migrate-down: ## Roll back the last migration
	cd $(SERVER_DIR) && go run ./cmd/migrate down

migrate-create: ## Create a new migration: make migrate-create name=add_users
	@test -n "$(name)" || (echo "usage: make migrate-create name=<migration_name>" && exit 1)
	migrate create -ext sql -dir $(SERVER_DIR)/migrations -seq $(name)

# ---------- Go server ----------

run: ## Run the API server
	cd $(SERVER_DIR) && go run ./cmd/api

build: ## Build api and migrate binaries into server/bin
	cd $(SERVER_DIR) && go build -o bin/api ./cmd/api
	cd $(SERVER_DIR) && go build -o bin/migrate ./cmd/migrate

test: ## Run tests
	cd $(SERVER_DIR) && go test ./...

vet: ## Run go vet
	cd $(SERVER_DIR) && go vet ./...

fmt: ## Format Go code
	cd $(SERVER_DIR) && gofmt -w .

tidy: ## Tidy go.mod / go.sum
	cd $(SERVER_DIR) && go mod tidy

check: fmt vet test ## Format, vet and test

clean: ## Remove build output
	rm -rf $(BIN_DIR)
