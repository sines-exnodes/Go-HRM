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

.PHONY: help run build test tidy fmt vet swag migrate-new migrate-up migrate-down migrate-version migrate-force

help:
	@echo "Targets:"
	@echo "  run               Run the API server (go run ./cmd/server)"
	@echo "  build             Build server binary to ./bin/server"
	@echo "  test              Run all tests"
	@echo "  tidy              go mod tidy"
	@echo "  fmt               gofmt -s -w ."
	@echo "  vet               go vet ./..."
	@echo "  swag              Regenerate Swagger docs into docs/swagger"
	@echo "  migrate-new name=NAME    Create empty up/down migration pair"
	@echo "  migrate-up        Apply all pending migrations"
	@echo "  migrate-down      Rollback one migration step"
	@echo "  migrate-version   Print current applied migration version"
	@echo "  migrate-force version=N  Force version (use only to fix dirty state)"

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

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
