.PHONY: up down run migrate-up migrate-down migration create-admin reset-password

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

# Usage: make create-admin email=you@example.com name="Your Name"
create-admin:
	cd server && go run ./cmd/admin create-admin --email "$(email)" --name "$(name)"

# Usage: make reset-password email=you@example.com
reset-password:
	cd server && go run ./cmd/admin reset-password --email "$(email)"
