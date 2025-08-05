.PHONY: all build lint test snapshot release clean goreleaser-download

APP_NAME = $(shell basename $(PWD))
BIN_DIR = bin
DIST_DIR = dist
DB_URL ?= postgres://usr_api:iop@127.0.0.1/test2?sslmode=disable

all: build

swagger:
	@echo "==> Init Swagger..."
	swag init -g ./cmd/main.go -o pkg/docs

build:
	@echo "==> Building..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd

snapshot:
	@echo "==> Building snapshot with GoReleaser..."
	@mkdir -p $(DIST_DIR)
	goreleaser release --snapshot --skip-publish --rm-dist

migrate-up:
	@echo "==> Running database migrations up..."
	goose -dir pkg/db/migrations postgres "$(DB_URL)" up

migrate-down:
	@echo "==> Rolling back database migrations..."
	goose -dir pkg/db/migrations postgres "$(DB_URL)" down

migrate-status:
	@echo "==> Checking migration status..."
	goose -dir pkg/db/migrations postgres "$(DB_URL)" status

migrate-create:
	@echo "==> Creating new migration: $(NAME)"
	goose -dir pkg/db/migrations create $(NAME) sql


sqlc-generate:
	@echo "==> Generating SQLC code..."
	sqlc generate

sqlc-verify:
	@echo "==> Verifying SQLC configuration..."
	sqlc verify


dev-setup: migrate-up sqlc-generate
	@echo "==> Development environment ready"

compose-up:
	@echo "==> Starting PostgreSQL..."
	docker compose up -d postgres

clean:
	@echo "==> Cleaning..."
	rm -rf $(BIN_DIR) $(DIST_DIR)