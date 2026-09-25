.PHONY: up down run migrate-up migrate-down migration

up:
	docker compose up -d --wait

down:
	docker compose down

run:
	cd server && go run ./cmd/api

migrate-up:
	cd server && go run ./cmd/migrate up

migrate-down:
	cd server && go run ./cmd/migrate down

# Requires the golang-migrate CLI:
#   brew install golang-migrate
# Usage: make migration name=init
migration:
	migrate create -ext sql -dir server/migrations -seq $(name)
