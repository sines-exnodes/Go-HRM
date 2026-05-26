# Suggested Commands

All make targets assume cwd = repo root. `.env` is auto-loaded by the Makefile (`include .env; export`).

## One-time tooling install

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
export PATH="$(go env GOPATH)/bin:$PATH"
cp .env.example .env   # set DB_* (or DATABASE_URL)
```

## Daily

| Action | Command |
|---|---|
| Run API server (localhost:8080) | `make run` |
| Build binary → `./bin/server` | `make build` |
| Format | `make fmt` (= `gofmt -s -w .`) |
| Vet | `make vet` (= `go vet ./...`) |
| Run all tests | `make test` (= `go test ./...`) |
| Create integration-test DB (idempotent) | `make test-db-up` |
| Regenerate Swagger | `make swag` |
| Tidy go.mod | `make tidy` |
| Smoke check | `curl -s http://localhost:8080/health \| jq` |
| Swagger UI | open `http://localhost:8080/swagger/index.html` |

## Migrations (golang-migrate, versioned SQL only)

| Action | Command |
|---|---|
| Create new up/down pair | `make migrate-new name=<snake_case_name>` |
| Apply pending | `make migrate-up` |
| Roll back one step | `make migrate-down` |
| Print applied version | `make migrate-version` |
| Force version (ONLY to fix dirty state) | `make migrate-force version=<N>` |

## Dev seed (idempotent; refuses if `APP_ENV=production`)

```bash
make seed-dev    # loads migrations/seeds/dev_fixtures.sql
```

## Docker

| Action | Command |
|---|---|
| Build prod image | `make docker-build` |
| Start prod stack (app + postgres) detached | `make docker-up` |
| Stop stack (keeps `postgres_data` volume) | `make docker-down` |
| Dev stack with Air hot-reload | `make docker-dev` |
| Tail app logs | `make docker-logs` |

## Local infra reference

- Postgres dev: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`. Main DB `exnodes_hrm`, test DB `exnodes_hrm_test`.
- For attachment-upload live verification, recreate the MinIO container per `docs/superpowers/verification/phase-04.md` §3.

## Darwin-specific notes

System shell utils on Darwin (BSD `sed`, `date`, `find`) differ from GNU. Project code does not currently depend on GNU-specific flags — keep it that way. If a script needs portable behaviour, prefer Go over shell.
