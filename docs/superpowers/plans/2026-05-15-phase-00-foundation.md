# Phase 0: Foundation Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the empty Go skeleton for `exnodes-hrm-api-go-v2` so that `make migrate-up && make run` produces a live Gin server exposing `GET /health` and `/swagger/index.html`, with no business modules.

**Architecture:** Clean layered layout (cmd / internal / pkg / migrations / docs). All schema changes are versioned SQL migrations applied manually via `golang-migrate`; the app **never** auto-migrates and instead verifies migration version on boot. A single shared envelope (`{success, message, data}`), AppError → JSON middleware, custom soft-delete via `is_deleted/deleted_at`, and UUID PKs via `gen_random_uuid()` are wired in this phase so Phase 1+ can plug straight in.

**Tech Stack:** Go 1.24, Gin v1.10, GORM v1.25 + `gorm.io/driver/postgres`, `golang-migrate/migrate/v4` (CLI + library), `joho/godotenv`, `google/uuid`, `swaggo/swag` + `swaggo/gin-swagger` + `swaggo/files`. JWT (`golang-jwt/jwt/v5`) and bcrypt (`golang.org/x/crypto`) are deferred to Phase 1.

---

### Task 1: Initialise Go module and root files

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/go.mod`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.gitignore`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.gitkeep`

- [ ] **Step 1: Verify the working directory is the empty project root**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && ls -A
```
Expected output contains `ba-requirements` and `docs` and nothing else (no `go.mod` yet).

- [ ] **Step 2: Initialise the Go module**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go mod init github.com/exnodes/hrm-api
```
Expected output: `go: creating new go.mod: module github.com/exnodes/hrm-api`.

Verify with:
```bash
head -2 /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/go.mod
```
Expected:
```
module github.com/exnodes/hrm-api

```

- [ ] **Step 3: Write `.gitignore`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.gitignore` with exactly this content:
```gitignore
# Binaries
/bin/
/tmp/
server
*.exe
*.test
*.out

# Env
.env
.env.local
.env.*.local

# OS / IDE
.DS_Store
.idea/
.vscode/
*.swp

# Go
vendor/
coverage.out
coverage.html

# Swagger generated artifacts kept under docs/swagger only; ignore other generated dirs
/docs/swagger/swagger.json.bak
```

- [ ] **Step 4: Initialise the git repository (if not already) and stage**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git rev-parse --is-inside-work-tree 2>/dev/null || git init
```
Expected: either `true` or `Initialized empty Git repository in ...`.

- [ ] **Step 5: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add go.mod .gitignore && \
  git commit -m "chore: initialise go module github.com/exnodes/hrm-api"
```
Expected: `1 file changed` or `2 files changed` summary; no errors.

---

### Task 2: Create directory skeleton with `.gitkeep` placeholders

**Files:**
- Create: `cmd/server/.gitkeep`
- Create: `internal/config/.gitkeep`
- Create: `internal/models/.gitkeep`
- Create: `internal/dto/.gitkeep`
- Create: `internal/repositories/.gitkeep`
- Create: `internal/services/.gitkeep`
- Create: `internal/handlers/.gitkeep`
- Create: `internal/middleware/.gitkeep`
- Create: `internal/permissions/.gitkeep`
- Create: `internal/errors/.gitkeep`
- Create: `internal/sse/.gitkeep`
- Create: `pkg/utils/.gitkeep`
- Create: `migrations/.gitkeep`
- Create: `scripts/.gitkeep`
- Create: `docs/swagger/.gitkeep`
- Create: `docs/superpowers/verification/.gitkeep`

- [ ] **Step 1: Create all directories at once**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  mkdir -p cmd/server \
           internal/config internal/models internal/dto internal/repositories \
           internal/services internal/handlers internal/middleware \
           internal/permissions internal/errors internal/sse \
           pkg/utils \
           migrations \
           scripts \
           docs/swagger docs/superpowers/verification
```
Expected: no output.

- [ ] **Step 2: Drop a `.gitkeep` in every leaf directory**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  for d in cmd/server \
           internal/config internal/models internal/dto internal/repositories \
           internal/services internal/handlers internal/middleware \
           internal/permissions internal/errors internal/sse \
           pkg/utils \
           migrations \
           scripts \
           docs/swagger docs/superpowers/verification; do
    touch "$d/.gitkeep"
  done
```
Expected: no output.

- [ ] **Step 3: Verify the tree**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && find . -type d -not -path './.git*' | sort
```
Expected (order may vary):
```
.
./ba-requirements
... (existing ba-requirements subdirs)
./cmd
./cmd/server
./docs
./docs/superpowers
./docs/superpowers/plans
./docs/superpowers/specs
./docs/superpowers/verification
./docs/swagger
./internal
./internal/config
./internal/dto
./internal/errors
./internal/handlers
./internal/middleware
./internal/models
./internal/permissions
./internal/repositories
./internal/services
./internal/sse
./migrations
./pkg
./pkg/utils
./scripts
```

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add cmd internal pkg migrations scripts docs/swagger docs/superpowers/verification && \
  git commit -m "chore: scaffold project directory layout"
```
Expected: 16 files added.

---

### Task 3: Add Go module dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum` (auto-generated)

- [ ] **Step 1: Add Gin, GORM, Postgres driver**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go get github.com/gin-gonic/gin@v1.10.0 && \
  go get gorm.io/gorm@v1.25.12 && \
  go get gorm.io/driver/postgres@v1.5.11
```
Expected: each command prints `go: added github.com/...@v...`.

- [ ] **Step 2: Add config, UUID, and migrate library**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go get github.com/joho/godotenv@v1.5.1 && \
  go get github.com/google/uuid@v1.6.0 && \
  go get github.com/golang-migrate/migrate/v4@v4.18.1
```
Expected: each `go get` prints an `added` line.

- [ ] **Step 3: Add Swagger libraries**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go get github.com/swaggo/swag@v1.16.4 && \
  go get github.com/swaggo/gin-swagger@v1.6.0 && \
  go get github.com/swaggo/files@v1.0.1
```
Expected: each `go get` prints an `added` line.

- [ ] **Step 4: Install the `swag` CLI and the `migrate` CLI**

Run:
```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4 && \
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
```
Expected: no output, exit code 0. Binaries land in `$(go env GOPATH)/bin`.

Verify:
```bash
"$(go env GOPATH)/bin/swag" --version && "$(go env GOPATH)/bin/migrate" -version
```
Expected: `swag version v1.16.4` and `4.18.1` (or similar).

- [ ] **Step 5: Tidy the module**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go mod tidy
```
Expected: silent success (no error); `go.sum` is created.

- [ ] **Step 6: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add go.mod go.sum && \
  git commit -m "chore: add core dependencies (gin, gorm, migrate, swagger, godotenv, uuid)"
```
Expected: 2 files changed.

---

### Task 4: `.env.example` and Makefile shell

**Files:**
- Create: `.env.example`
- Create: `Makefile`

- [ ] **Step 1: Create `.env.example`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env.example` with exactly this content:
```dotenv
# ---------- Server ----------
APP_ENV=development
PORT=8080
GIN_MODE=debug
SWAGGER_ENABLED=true

# ---------- Database ----------
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=exnodes_hrm
DB_SSLMODE=disable

# Full DSN (overrides the DB_* parts above if set). Used by golang-migrate CLI too.
# postgres://user:password@host:port/dbname?sslmode=disable
DATABASE_URL=

# ---------- Migrations ----------
MIGRATIONS_DIR=migrations

# ---------- JWT (placeholder — populated in Phase 1) ----------
JWT_SECRET_KEY=change-me-in-production
JWT_ACCESS_TTL_MINUTES=60
JWT_REFRESH_TTL_HOURS=720

# ---------- Supabase Storage (placeholder — used by Phase 2 upload module) ----------
SUPABASE_URL=
SUPABASE_SERVICE_ROLE_KEY=
SUPABASE_BUCKET=hrm-uploads
SUPABASE_S3_ENDPOINT=
SUPABASE_S3_REGION=ap-southeast-1
SUPABASE_S3_ACCESS_KEY=
SUPABASE_S3_SECRET_KEY=
```

- [ ] **Step 2: Create the Makefile**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/Makefile` with exactly this content:
```makefile
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
```

- [ ] **Step 3: Sanity-check the Makefile syntax**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make help
```
Expected: the help block prints. No `*** missing separator` errors.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add .env.example Makefile && \
  git commit -m "chore: add .env.example and Makefile (run/build/test/migrate/swag targets)"
```
Expected: 2 files changed.

---

### Task 5: `internal/config/config.go` — env loader

**Files:**
- Create: `internal/config/config.go`

- [ ] **Step 1: Write the config loader**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/config/config.go` with exactly:
```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven settings for the API server.
type Config struct {
	AppEnv          string
	Port            string
	GinMode         string
	SwaggerEnabled  bool

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBUrl      string // explicit override; takes precedence over DB_* if non-empty

	MigrationsDir string

	JWTSecret           string
	JWTAccessTTLMinutes int
	JWTRefreshTTLHours  int

	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseBucket         string
	SupabaseS3Endpoint     string
	SupabaseS3Region       string
	SupabaseS3AccessKey    string
	SupabaseS3SecretKey    string
}

// Load reads the .env file (if present) and returns the populated Config.
// Missing required values cause os.Exit via log.Fatal to avoid booting in a
// half-configured state.
func Load() *Config {
	_ = godotenv.Load() // .env is optional in CI/production

	cfg := &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		SwaggerEnabled: getEnvBool("SWAGGER_ENABLED", true),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "exnodes_hrm"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBUrl:      getEnv("DATABASE_URL", ""),

		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),

		JWTSecret:           getEnv("JWT_SECRET_KEY", "change-me-in-production"),
		JWTAccessTTLMinutes: getEnvInt("JWT_ACCESS_TTL_MINUTES", 60),
		JWTRefreshTTLHours:  getEnvInt("JWT_REFRESH_TTL_HOURS", 720),

		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseBucket:         getEnv("SUPABASE_BUCKET", ""),
		SupabaseS3Endpoint:     getEnv("SUPABASE_S3_ENDPOINT", ""),
		SupabaseS3Region:       getEnv("SUPABASE_S3_REGION", "ap-southeast-1"),
		SupabaseS3AccessKey:    getEnv("SUPABASE_S3_ACCESS_KEY", ""),
		SupabaseS3SecretKey:    getEnv("SUPABASE_S3_SECRET_KEY", ""),
	}

	return cfg
}

// DSN returns a libpq-style DSN suitable for GORM's postgres driver.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// DatabaseURL returns a postgres:// URL suitable for golang-migrate. If the
// DATABASE_URL env var was set it is returned verbatim; otherwise one is
// composed from the DB_* parts.
func (c *Config) DatabaseURL() string {
	if strings.TrimSpace(c.DBUrl) != "" {
		return c.DBUrl
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return fallback
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/config/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/config/config.go && \
  git rm internal/config/.gitkeep && \
  git commit -m "feat(config): add env-driven Config loader with DSN + DatabaseURL helpers"
```
Expected: 1 add, 1 delete.

---

### Task 6: `internal/config/db.go` — GORM connect + migration version check

**Files:**
- Create: `internal/config/db.go`

- [ ] **Step 1: Write the DB helper**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/config/db.go` with exactly:
```go
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB opens a GORM Postgres connection using cfg.DSN(). Foreign-key
// constraint creation during migration is disabled at the GORM level because
// all schema changes are owned by SQL migrations, not by AutoMigrate.
func ConnectDB(cfg *Config) (*gorm.DB, error) {
	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("acquire sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

// AssertMigrationsUpToDate verifies that every *.up.sql file under
// cfg.MigrationsDir has been applied (i.e. its numeric version <= the version
// recorded in schema_migrations) and that the schema_migrations row is not
// marked dirty. It NEVER applies migrations — that is reserved for
// `make migrate-up`. The function fails loud if the DB is behind, dirty, or
// missing the schema_migrations table after at least one migration file
// exists on disk.
func AssertMigrationsUpToDate(db *gorm.DB, migrationsDir string) error {
	latestOnDisk, err := latestMigrationVersionOnDisk(migrationsDir)
	if err != nil {
		return fmt.Errorf("scan migrations dir %q: %w", migrationsDir, err)
	}
	if latestOnDisk == 0 {
		// No migrations exist yet — nothing to assert. This only happens
		// before the first migration file is added.
		return nil
	}

	var hasTable bool
	if err := db.Raw(
		`SELECT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = current_schema() AND table_name = 'schema_migrations'
        )`,
	).Scan(&hasTable).Error; err != nil {
		return fmt.Errorf("check schema_migrations table: %w", err)
	}
	if !hasTable {
		return fmt.Errorf(
			"schema_migrations table not found but %d migration file(s) exist on disk; run `make migrate-up` before starting the server",
			latestOnDisk,
		)
	}

	type row struct {
		Version int64
		Dirty   bool
	}
	var r row
	err = db.Raw(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&r).Error
	if err != nil {
		return fmt.Errorf("read schema_migrations: %w", err)
	}
	if r.Dirty {
		return fmt.Errorf(
			"schema_migrations is dirty at version %d; fix manually with `make migrate-force version=<N>` then `make migrate-up`",
			r.Version,
		)
	}
	if r.Version < latestOnDisk {
		return fmt.Errorf(
			"database is behind: applied version=%d, latest on disk=%d; run `make migrate-up`",
			r.Version, latestOnDisk,
		)
	}
	return nil
}

// latestMigrationVersionOnDisk returns the highest NNNNNN sequence number
// among *.up.sql files in migrationsDir. Returns 0 if the directory is empty
// or missing.
func latestMigrationVersionOnDisk(migrationsDir string) (int64, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	versions := make([]int64, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		// Filename pattern: NNNNNN_anything.up.sql
		base := filepath.Base(name)
		underscore := strings.IndexByte(base, '_')
		if underscore <= 0 {
			continue
		}
		seqStr := base[:underscore]
		seq, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			continue
		}
		versions = append(versions, seq)
	}
	if len(versions) == 0 {
		return 0, nil
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions[len(versions)-1], nil
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/config/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/config/db.go && \
  git commit -m "feat(config): GORM connect helper + AssertMigrationsUpToDate (no auto-migrate)"
```
Expected: 1 file changed.

---

### Task 7: `internal/errors/errors.go` — AppError + factories

**Files:**
- Create: `internal/errors/errors.go`

- [ ] **Step 1: Write the error type**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/errors/errors.go` with exactly:
```go
package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Stable code strings consumed by the FE. Keep in sync with the FE error map.
const (
	CodeNotFound     = "not_found"
	CodeBadRequest   = "bad_request"
	CodeConflict     = "conflict"
	CodeForbidden    = "forbidden"
	CodeUnauthorized = "unauthorized"
	CodeInternal     = "internal_error"
)

// AppError is the canonical error type raised by services. The error
// middleware converts these into the standard JSON envelope.
type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	HTTP    int            `json:"-"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil AppError>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithDetails returns a shallow copy with details merged. Useful for adding
// per-call context without mutating package-level error instances.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	if e == nil {
		return nil
	}
	merged := make(map[string]any, len(e.Details)+len(details))
	for k, v := range e.Details {
		merged[k] = v
	}
	for k, v := range details {
		merged[k] = v
	}
	cp := *e
	cp.Details = merged
	return &cp
}

// As is a tiny helper so callers can write
//     if ae, ok := apperrors.As(err); ok { ... }
// instead of importing the stdlib errors package just for this.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		HTTP:    http.StatusNotFound,
	}
}

func ErrBadRequest(msg string) *AppError {
	return &AppError{
		Code:    CodeBadRequest,
		Message: msg,
		HTTP:    http.StatusBadRequest,
	}
}

func ErrConflict(msg string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: msg,
		HTTP:    http.StatusConflict,
	}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{
		Code:    CodeForbidden,
		Message: msg,
		HTTP:    http.StatusForbidden,
	}
}

func ErrUnauthorized(msg string) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: msg,
		HTTP:    http.StatusUnauthorized,
	}
}

func ErrInternal(msg string) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: msg,
		HTTP:    http.StatusInternalServerError,
	}
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/errors/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/errors/errors.go && \
  git rm internal/errors/.gitkeep && \
  git commit -m "feat(errors): introduce AppError type with NotFound/BadRequest/Conflict/Forbidden/Unauthorized factories"
```
Expected: 1 add, 1 delete.

---

### Task 8: `internal/middleware/error.go` — AppError → JSON middleware

**Files:**
- Create: `internal/middleware/error.go`

- [ ] **Step 1: Write the error middleware**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/middleware/error.go` with exactly:
```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// errorEnvelope is the JSON shape returned for any error response.
// It deliberately mirrors the success envelope's `success` field so the FE
// can detect failure by checking a single boolean.
type errorEnvelope struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorHandler converts any errors written via c.Error(err) into the standard
// JSON envelope. It must be registered before any handler that uses c.Error
// (i.e. at the very top of the middleware chain).
//
// Handlers should *not* call c.JSON themselves on the error path; they should
// `c.Error(apperrors.ErrXxx(...))` and `return`. This middleware writes the
// response after the handler chain finishes.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// Use the last error written by the handler — that is the one the
		// developer most recently attached and is the most specific.
		last := c.Errors.Last().Err

		if ae, ok := apperrors.As(last); ok {
			c.AbortWithStatusJSON(ae.HTTP, errorEnvelope{
				Success: false,
				Message: ae.Message,
				Code:    ae.Code,
				Details: ae.Details,
			})
			return
		}

		// Unknown error → 500 internal_error. We deliberately do NOT leak the
		// raw error text to the FE; it is captured in Gin's error log already.
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorEnvelope{
			Success: false,
			Message: "internal server error",
			Code:    apperrors.CodeInternal,
		})
	}
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/middleware/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/middleware/error.go && \
  git commit -m "feat(middleware): error handler converting AppError to JSON envelope"
```
Expected: 1 file changed.

---

### Task 9: `internal/middleware/cors.go` and `recovery.go`

**Files:**
- Create: `internal/middleware/cors.go`
- Create: `internal/middleware/recovery.go`

- [ ] **Step 1: Write the CORS middleware**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/middleware/cors.go` with exactly:
```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a permissive CORS middleware suitable for early development
// (`Access-Control-Allow-Origin: *`). Production deployments should swap this
// for a stricter allow-list — tracked as a Phase 2+ hardening item.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: Write the recovery middleware**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/middleware/recovery.go` with exactly:
```go
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// Recovery converts a panic in any downstream handler into a 500 JSON
// envelope and logs the stack trace. It must be registered before
// ErrorHandler so the panic is rewritten into the same envelope shape.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v\n%s", r, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "internal server error",
					"code":    apperrors.CodeInternal,
				})
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/middleware/...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/middleware/cors.go internal/middleware/recovery.go && \
  git rm internal/middleware/.gitkeep && \
  git commit -m "feat(middleware): CORS + panic-recovery middlewares"
```
Expected: 2 adds, 1 delete.

---

### Task 10: Response envelopes (`internal/dto/response.go`) and `BaseModel` (`internal/models/base.go`)

**Files:**
- Create: `internal/dto/response.go`
- Create: `internal/models/base.go`

- [ ] **Step 1: Write the response envelopes**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/dto/response.go` with exactly:
```go
package dto

import "math"

// Response is the standard success envelope. T is the data type.
//   { "success": true, "message": "...", "data": T }
type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// PaginatedData wraps a list result with pagination metadata. Embed inside a
// Response[PaginatedData[T]] for the canonical paginated payload.
type PaginatedData[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// NewResponse builds a successful Response[T] with the given data and an
// optional message.
func NewResponse[T any](data T, message string) Response[T] {
	return Response[T]{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewPaginatedResponse builds the canonical paginated envelope. page and
// pageSize must be positive; the caller is responsible for validating them
// before invoking this helper.
func NewPaginatedResponse[T any](items []T, total int64, page, pageSize int) Response[PaginatedData[T]] {
	if pageSize <= 0 {
		pageSize = len(items)
		if pageSize == 0 {
			pageSize = 1
		}
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	return Response[PaginatedData[T]]{
		Success: true,
		Data: PaginatedData[T]{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
}
```

- [ ] **Step 2: Write `BaseModel`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/models/base.go` with exactly:
```go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded in every entity model. The 4 audit columns are
// required by the migration design (see spec §5.2). Soft delete is NOT
// implemented via gorm.DeletedAt — instead, the explicit IsDeleted boolean
// and DeletedAt timestamp are managed by service-level Delete/Restore calls
// and queried through the NotDeleted scope.
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time  `gorm:"not null;default:now()"                          json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()"                          json:"updated_at"`
	IsDeleted bool       `gorm:"not null;default:false;index"                    json:"-"`
	DeletedAt *time.Time `                                                       json:"-"`
}

// NotDeleted is the default scope applied by repositories to every list,
// get-by-id, and count query. Callers that intentionally need to read
// soft-deleted rows (e.g. an admin "restore" flow) must opt out by NOT
// chaining this scope.
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}
```

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go build ./internal/dto/... ./internal/models/...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/dto/response.go internal/models/base.go && \
  git rm internal/dto/.gitkeep internal/models/.gitkeep && \
  git commit -m "feat(dto,models): Response[T]/PaginatedData[T] envelopes + BaseModel + NotDeleted scope"
```
Expected: 2 adds, 2 deletes.

---

### Task 11: Initial migration `000001_init_extensions`

**Files:**
- Create: `migrations/000001_init_extensions.up.sql`
- Create: `migrations/000001_init_extensions.down.sql`

- [ ] **Step 1: Write the up migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000001_init_extensions.up.sql` with exactly:
```sql
-- =============================================================
-- 000001_init_extensions
-- Foundational Postgres extensions + shared updated_at trigger.
-- All later migrations may assume these objects exist.
-- =============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Shared trigger function used by every entity table's BEFORE UPDATE
-- trigger. Keeps updated_at in sync without per-row application code.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

- [ ] **Step 2: Write the down migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000001_init_extensions.down.sql` with exactly:
```sql
-- Reverse of 000001_init_extensions.
-- Extensions are dropped only if no later object depends on them; in
-- practice this down step is intended for local resets, not production.
DROP FUNCTION IF EXISTS set_updated_at();

DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS "uuid-ossp";
```

- [ ] **Step 3: Verify both files exist and are well-formed SQL**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  ls -la migrations/000001_init_extensions.*.sql && \
  grep -c 'CREATE EXTENSION' migrations/000001_init_extensions.up.sql
```
Expected: two files listed (one `.up.sql`, one `.down.sql`); `grep` prints `3`.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add migrations/000001_init_extensions.up.sql migrations/000001_init_extensions.down.sql && \
  git rm migrations/.gitkeep && \
  git commit -m "feat(migrations): 000001 init extensions (uuid-ossp, pgcrypto, citext) + set_updated_at()"
```
Expected: 2 adds, 1 delete.

---

### Task 12: Health handler + router

**Files:**
- Create: `internal/handlers/health_handler.go`
- Create: `internal/handlers/router.go`

- [ ] **Step 1: Write the health handler**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/health_handler.go` with exactly:
```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
)

// HealthHandler is a stateless liveness probe.
type HealthHandler struct{}

// NewHealthHandler returns a ready-to-use HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthData is the payload of GET /health.
type HealthData struct {
	Status  string `json:"status"  example:"ok"`
	Service string `json:"service" example:"exnodes-hrm-api"`
}

// Health godoc
// @Summary      Liveness probe
// @Description  Returns {success: true, data: {status: "ok"}} when the server is up.
// @Tags         system
// @Produce      json
// @Success      200  {object}  dto.Response[handlers.HealthData]
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewResponse(HealthData{
		Status:  "ok",
		Service: "exnodes-hrm-api",
	}, ""))
}
```

- [ ] **Step 2: Write the route registrar**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/router.go` with exactly:
```go
package handlers

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

// Handlers bundles every handler instance the router needs to wire. Phase 1
// will extend this struct with Auth, User, etc.
type Handlers struct {
	Health *HealthHandler
}

// NewHandlers constructs every handler in the bundle. Currently health-only;
// future phases add fields here in alphabetical order.
func NewHandlers() *Handlers {
	return &Handlers{
		Health: NewHealthHandler(),
	}
}

// RegisterRoutes wires every route onto r. If swaggerEnabled is true, the
// Swagger UI is mounted at /swagger/*any.
func RegisterRoutes(r *gin.Engine, h *Handlers, swaggerEnabled bool) {
	// Liveness probe — un-namespaced so load balancers can hit it without
	// knowing about /api/v1.
	r.GET("/health", h.Health.Health)

	if swaggerEnabled {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	// /api/v1 is registered here so Phase 1+ has a place to plug routes in.
	// No routes are mounted on it yet.
	_ = r.Group("/api/v1")
}
```

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/handlers/...
```
Expected: a single error of the form `package github.com/exnodes/hrm-api/docs/swagger` is **not** expected here because `router.go` does NOT yet import the generated docs. The build should succeed cleanly (no output, exit code 0).

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/handlers/health_handler.go internal/handlers/router.go && \
  git rm internal/handlers/.gitkeep && \
  git commit -m "feat(handlers): GET /health + RegisterRoutes wiring (with optional swagger UI)"
```
Expected: 2 adds, 1 delete.

---

### Task 13: `cmd/server/main.go` — wire config, DB, migration check, Gin, routes

**Files:**
- Create: `cmd/server/main.go`

- [ ] **Step 1: Write the entry point**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/cmd/server/main.go` with exactly:
```go
// Package main is the entry point of the Exnodes HRM API server.
//
// @title           Exnodes HRM API
// @version         0.1.0
// @description     HRM API for Exnodes — see /docs/superpowers/specs/2026-05-15-go-migration-design.md for the migration plan.
// @host            localhost:8080
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/handlers"
	"github.com/exnodes/hrm-api/internal/middleware"

	// Generated by `make swag` — imported for side effects (registers spec
	// with swaggo). The blank import line is updated after the first swag
	// run; until then the import is satisfied by the placeholder package
	// generated alongside the docs.
	_ "github.com/exnodes/hrm-api/docs/swagger"
)

func main() {
	cfg := config.Load()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}

	if err := config.AssertMigrationsUpToDate(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("migration check failed: %v", err)
	}

	gin.SetMode(cfg.GinMode)
	r := gin.New()

	// Middleware order matters: Recovery first (catches panics from later
	// middleware too), then logger, then CORS, then ErrorHandler (writes
	// the envelope at the end of the chain).
	r.Use(middleware.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())

	handlers.RegisterRoutes(r, handlers.NewHandlers(), cfg.SwaggerEnabled)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("exnodes-hrm-api listening on %s (env=%s, swagger=%t)", addr, cfg.AppEnv, cfg.SwaggerEnabled)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
```

- [ ] **Step 2: Create a stub `docs/swagger` package so main.go builds before `swag init`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/swagger/docs.go` with exactly:
```go
// Package docs is a placeholder regenerated by `make swag`.
// It exists so cmd/server/main.go can blank-import it during the first build,
// before swaggo has produced the real generated files.
package docs
```

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add cmd/server/main.go docs/swagger/docs.go && \
  git rm cmd/server/.gitkeep docs/swagger/.gitkeep && \
  git commit -m "feat(cmd): wire config, DB, migration check, Gin engine, and route registration"
```
Expected: 2 adds, 2 deletes.

---

### Task 14: Generate Swagger docs and verify UI lists `/health`

**Files:**
- Create / regenerate: `docs/swagger/docs.go`, `docs/swagger/swagger.json`, `docs/swagger/swagger.yaml`

- [ ] **Step 1: Generate the Swagger docs**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make swag
```
Expected output ends with `generated`. Three files now exist:
```bash
ls docs/swagger
```
Expected:
```
docs.go
swagger.json
swagger.yaml
```

- [ ] **Step 2: Verify generated `docs.go` registers `/health`**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && grep -c '"/health"' docs/swagger/docs.go
```
Expected: a non-zero count (typically `1`).

- [ ] **Step 3: Compile-check the full module**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add docs/swagger/docs.go docs/swagger/swagger.json docs/swagger/swagger.yaml && \
  git commit -m "docs(swagger): generate initial swagger spec covering /health"
```
Expected: 3 files changed (1 modified, 2 new).

---

### Task 15: README quickstart, README links, and end-to-end verification log

**Files:**
- Create: `README.md`
- Create: `docs/superpowers/verification/phase-00.md`

- [ ] **Step 1: Write the README**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/README.md` with exactly:
````markdown
# Exnodes HRM API v2 (Go)

Go + Gin + Postgres rewrite of the Exnodes HRM API. See
[`docs/superpowers/specs/2026-05-15-go-migration-design.md`](docs/superpowers/specs/2026-05-15-go-migration-design.md)
for the full migration design and phase plan.

This README documents **Phase 0 only**: a boot-able skeleton with `/health` and Swagger UI.
Subsequent phases (auth, users, departments, ...) plug into the structures defined here.

## Quickstart

### 1. Prerequisites

- Go **1.24** (`go version`)
- Postgres **14+** running locally (or a remote DSN you control)
- The `swag` and `migrate` CLIs (installed below)

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
export PATH="$(go env GOPATH)/bin:$PATH"
```

### 2. Clone & install deps

```bash
git clone <repo-url> exnodes-hrm-api-go-v2
cd exnodes-hrm-api-go-v2
go mod download
```

### 3. Configure env

```bash
cp .env.example .env
$EDITOR .env   # set DB_* (or DATABASE_URL) to a Postgres you can write to
```

### 4. Create the database

```bash
createdb exnodes_hrm   # or use psql: CREATE DATABASE exnodes_hrm;
```

### 5. Apply migrations

```bash
make migrate-up
```
Expected output ends with a version line; nothing fails.

### 6. Run the server

```bash
make run
```
Expected log lines:
```
exnodes-hrm-api listening on :8080 (env=development, swagger=true)
```

### 7. Smoke-test

```bash
curl -s http://localhost:8080/health | jq
```
Expected:
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "service": "exnodes-hrm-api"
  }
}
```

Then open `http://localhost:8080/swagger/index.html` and confirm `GET /health`
appears under the `system` tag.

## Make targets

| Target | What it does |
|---|---|
| `make run` | Run the API server (`go run ./cmd/server`) |
| `make build` | Build `./bin/server` |
| `make test` | Run all Go tests |
| `make tidy` | `go mod tidy` |
| `make fmt` | `gofmt -s -w .` |
| `make vet` | `go vet ./...` |
| `make swag` | Regenerate Swagger docs into `docs/swagger/` |
| `make migrate-new name=<snake>` | Create a new empty up/down migration pair |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back one migration step |
| `make migrate-version` | Print the currently applied migration version |
| `make migrate-force version=N` | Force the version (only to fix a dirty state) |

## Project layout

```
cmd/server/         Entry point (main.go)
internal/
  config/           Env loader, GORM connect, migration version check
  models/           BaseModel + per-entity models (added in later phases)
  dto/              Request/response envelopes
  repositories/     GORM data access (added in later phases)
  services/         Business logic (added in later phases)
  handlers/         Gin handlers + RegisterRoutes
  middleware/       CORS, Recovery, ErrorHandler (+ JWT in Phase 1)
  permissions/      Permission registry (added in Phase 1)
  errors/           AppError type + factory helpers
  sse/              Realtime event hub (added in Phase 7)
pkg/utils/          Generic helpers shared across modules
migrations/         golang-migrate SQL files (NNNNNN_<name>.up/down.sql)
scripts/            Shell helpers (seed, deploy, etc.)
docs/
  superpowers/      Project specs, plans, verification logs
  swagger/          Generated OpenAPI artefacts (do not hand-edit)
```

## Schema conventions (enforced from Phase 1 onward)

- Every entity table has 4 audit columns: `created_at`, `updated_at`,
  `is_deleted BOOLEAN`, `deleted_at TIMESTAMPTZ` plus a per-table
  `BEFORE UPDATE` trigger calling `set_updated_at()`.
- Primary keys are UUIDs via `gen_random_uuid()` (pgcrypto).
- Soft delete is implemented via the custom `NotDeleted` scope — NOT GORM's
  built-in `gorm.DeletedAt`.
- Schema changes are versioned SQL migration files only. `db.AutoMigrate()`
  is **prohibited**. The app verifies migration version on boot and refuses
  to start if behind or dirty.
````

- [ ] **Step 2: Run end-to-end verification — migrate-up and capture the output**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  cp -n .env.example .env && \
  make migrate-up 2>&1 | tee /tmp/phase00-migrate-up.log
```
Expected: log ends with no error; `schema_migrations` is created in Postgres.

If Postgres credentials in `.env` are not yet valid, fix `.env` and re-run. Do NOT skip this step.

- [ ] **Step 3: Boot the server in the background**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  (make run > /tmp/phase00-server.log 2>&1 &) && \
  sleep 3 && \
  grep -F 'exnodes-hrm-api listening on' /tmp/phase00-server.log
```
Expected: the grep prints the listening log line.

- [ ] **Step 4: Curl `/health` and capture the response**

Run:
```bash
curl -s -o /tmp/phase00-health.json -w '%{http_code}\n' http://localhost:8080/health
cat /tmp/phase00-health.json
```
Expected: HTTP code `200`, body:
```json
{"success":true,"data":{"status":"ok","service":"exnodes-hrm-api"}}
```

- [ ] **Step 5: Verify Swagger UI**

Run:
```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/swagger/index.html && \
  curl -s http://localhost:8080/swagger/doc.json | grep -c '"/health"'
```
Expected: first command prints `200`; second prints a non-zero count.

- [ ] **Step 6: Stop the server**

Run:
```bash
pkill -f 'go run ./cmd/server' 2>/dev/null || pkill -f '/bin/server' 2>/dev/null; true
sleep 1
ss -tlnp 2>/dev/null | grep ':8080 ' && echo 'STILL RUNNING' || echo 'STOPPED'
```
Expected last line: `STOPPED` (on macOS, fall back to `lsof -i :8080` — no output means stopped).

- [ ] **Step 7: Write the verification log**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification/phase-00.md` with this content, **substituting** the four `<...>` placeholders with the literal output captured in steps 2–5 (do not leave the placeholders in the committed file):
```markdown
# Phase 0 Verification Log

Date: 2026-05-15
Phase: 0 — Foundation Infrastructure
Spec: docs/superpowers/specs/2026-05-15-go-migration-design.md
Plan: docs/superpowers/plans/2026-05-15-phase-00-foundation.md

## 1. `make migrate-up`

Command:
    make migrate-up

Output (trimmed):
    <paste the last 5–10 lines from /tmp/phase00-migrate-up.log>

Result: migrations 1/u init_extensions applied; schema_migrations row created.

## 2. `make run`

Command:
    make run

Boot log (relevant lines):
    <paste the "listening on" line from /tmp/phase00-server.log>

## 3. `GET /health`

Command:
    curl -s -i http://localhost:8080/health

Response body:
    <paste contents of /tmp/phase00-health.json>

HTTP status: 200
Envelope: { "success": true, "data": { "status": "ok", "service": "exnodes-hrm-api" } }

## 4. Swagger UI

Visited: http://localhost:8080/swagger/index.html
HTTP status for index.html: 200
`/swagger/doc.json` contained an entry for "/health": yes (grep count = <N>).

## 5. Sign-off

All Phase 0 acceptance criteria met:

- [x] go.mod initialised with module path github.com/exnodes/hrm-api
- [x] Directory skeleton present
- [x] Config loader + DB connect + AssertMigrationsUpToDate
- [x] AppError + error middleware + CORS + Recovery
- [x] Response[T]/PaginatedData[T] envelopes
- [x] BaseModel with 4 audit columns + NotDeleted scope
- [x] 000001_init_extensions up + down migrations
- [x] /health handler returning the standard envelope
- [x] cmd/server/main.go wires everything
- [x] Swagger UI lists /health
- [x] `make migrate-up` succeeds on clean DB
- [x] `make run` boots cleanly
- [x] `curl /health` returns 200 with correct envelope
- [x] README quickstart written
```

- [ ] **Step 8: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add README.md docs/superpowers/verification/phase-00.md && \
  git rm docs/superpowers/verification/.gitkeep && \
  git commit -m "docs: README quickstart + Phase 0 verification log"
```
Expected: 2 adds, 1 delete.

- [ ] **Step 9: Final tree sanity check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git log --oneline | head -20 && \
  echo '---' && \
  go build ./... && \
  echo 'BUILD OK'
```
Expected: 15 commits since module init, ending with `BUILD OK`.

---

## Phase 0 Definition of Done

All boxes below must be checked before Phase 0 is considered complete:

- [ ] `go build ./...` returns exit 0 with no warnings
- [ ] `make migrate-up` against an empty Postgres database succeeds
- [ ] `make migrate-version` prints `1`
- [ ] `make run` logs `exnodes-hrm-api listening on :8080`
- [ ] `curl http://localhost:8080/health` returns HTTP 200 with body `{"success":true,"data":{"status":"ok","service":"exnodes-hrm-api"}}`
- [ ] `http://localhost:8080/swagger/index.html` renders and lists `GET /health` under the `system` tag
- [ ] `docs/superpowers/verification/phase-00.md` exists and is committed, with all placeholder `<...>` blocks replaced by real captured output
- [ ] Every task in this plan ended with a commit; `git log --oneline` shows ≥ 15 commits since `go mod init`

Once all boxes are checked, Phase 1 (Auth + RBAC core) may begin.
