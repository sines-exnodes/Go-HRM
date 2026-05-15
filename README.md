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
