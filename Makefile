SHELL := /usr/bin/env bash

# Load .env if present so DATABASE_URL/PORT/etc. are visible to recipes.
ifneq (,$(wildcard .env))
include .env
export
endif

GOPATH_BIN := $(shell go env GOPATH)/bin
MIGRATE    := $(GOPATH_BIN)/migrate
SWAG       := $(GOPATH_BIN)/swag

MIGRATIONS_DIR ?= migrations

# Build DATABASE_URL from DB_* if not provided.
ifeq ($(strip $(DATABASE_URL)),)
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
endif

# Separate database used by integration tests. Override TEST_DB_NAME or pass
# TEST_DATABASE_URL directly to point tests at a different instance.
TEST_DB_NAME      ?= exnodes_hrm_test
TEST_DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(TEST_DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: help run build test test-db-up tidy fmt vet swag \
        migrate-new migrate-up migrate-down migrate-version migrate-force \
        seed-dev \
        docker-build docker-up docker-down docker-dev docker-logs

help:
	@echo "Targets:"
	@echo "  run               Run the API server (go run ./cmd/server)"
	@echo "  build             Build server binary to ./bin/server"
	@echo "  test              Run all tests"
	@echo "  test-db-up        Create the integration test database (idempotent)"
	@echo "  tidy              go mod tidy"
	@echo "  fmt               gofmt -s -w ."
	@echo "  vet               go vet ./..."
	@echo "  swag              Regenerate Swagger docs into docs/swagger"
	@echo "  migrate-new name=NAME    Create empty up/down migration pair"
	@echo "  migrate-up        Apply all pending migrations"
	@echo "  migrate-down      Rollback one migration step"
	@echo "  migrate-version   Print current applied migration version"
	@echo "  migrate-force version=N  Force version (use only to fix dirty state)"
	@echo "  seed-dev          Load development sample data (idempotent; refuses if APP_ENV=production)"
	@echo "  docker-build      Build the prod Docker image"
	@echo "  docker-up         Start prod stack (app + postgres) in background"
	@echo "  docker-down       Stop the stack (keeps the postgres_data volume)"
	@echo "  docker-dev        Start dev stack with Air hot-reload"
	@echo "  docker-logs       Tail app logs from the running stack"

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

# Create the integration-test database if it doesn't exist. Uses the same
# credentials as the main DB (DB_USER/DB_PASSWORD/DB_HOST/DB_PORT) but a
# separate dbname (TEST_DB_NAME, defaults to exnodes_hrm_test).
test-db-up:
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres -tAc \
		"SELECT 1 FROM pg_database WHERE datname='$(TEST_DB_NAME)'" | grep -q 1 || \
	PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres -c \
		"CREATE DATABASE \"$(TEST_DB_NAME)\""
	@echo "Test DB ready: $(TEST_DB_NAME)"
	@echo "Export TEST_DATABASE_URL to run integration tests, e.g.:"
	@echo "  export TEST_DATABASE_URL='$(TEST_DATABASE_URL)'"

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

vet:
	go vet ./...

swag:
	$(SWAG) init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal

migrate-new:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-new name=<snake_name>" && exit 1; fi
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-force:
	@if [ -z "$(version)" ]; then echo "usage: make migrate-force version=<N>" && exit 1; fi
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(version)

# ---------------------------------------------------------------------------
# Dev seed
# ---------------------------------------------------------------------------

# Load dev/demo fixtures: 19 sample employees with skills, leave requests,
# attendance, announcements. Idempotent — re-running is a no-op.
# Refuses to run when APP_ENV=production (safety guard).
seed-dev:
	@if [ "$(APP_ENV)" = "production" ]; then \
		echo "seed-dev: refusing to run (APP_ENV=production)"; \
		exit 1; \
	fi
	@echo "seed-dev: target DB = $(DATABASE_URL)"
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f migrations/seeds/dev_fixtures.sql

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up

docker-logs:
	docker compose logs -f app
