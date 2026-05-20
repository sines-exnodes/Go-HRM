# Project Overview — exnodes-hrm-api

**This memory is a POINTER, not a snapshot.** Authoritative state lives in
[`docs/superpowers/CHECKPOINT.md`](../../docs/superpowers/CHECKPOINT.md).
If this file and CHECKPOINT.md disagree, CHECKPOINT.md wins.

## What this project is

Go rewrite (Go 1.25 + Gin + GORM + PostgreSQL 14+) of the Python Exnodes HRM
backend. Serves a web operations portal and a mobile employee app. REST API
under `/api/v1/`, JWT auth, in-code RBAC permission registry.

The work is **phased and specification-first**. Each phase = a vertical
slice (migration → models → DTOs → repo → service → handler → tests →
live verification → commit).

## Where to start when resuming

Read in this order — STOP after each step is enough to answer your question:

1. `docs/superpowers/CHECKPOINT.md` — current phase, what's verified, what's next, known follow-ups
2. The relevant spec or per-phase plan:
   - Spec: `docs/superpowers/specs/2026-05-15-go-migration-design.md` (the migration design)
   - Plans: `docs/superpowers/plans/2026-05-15-phase-NN-*.md` (always check the `⚠️ REVISION NOTES` block at the top of a plan before executing — task bodies are pre-revision and may be wrong)
3. The previous phase's verification log: `docs/superpowers/verification/phase-NN.md` — proof of what was actually built vs the spec
4. Only NOW read source files for the specific question. Use Serena's `find_symbol` / `get_references` instead of scanning files.

## Tech stack snapshot

- **HTTP**: Gin (`gin-gonic/gin`)
- **ORM**: GORM + Postgres driver
- **Migrations**: `golang-migrate/migrate/v4`, versioned SQL only (NEVER `db.AutoMigrate()` — the server asserts the applied version on boot)
- **Auth**: JWT HS256 (`golang-jwt/jwt/v5`) + bcrypt cost 12
- **Storage**: AWS SDK Go v2 S3 → Supabase S3-compatible (configurable via `STORAGE_*` env)
- **API docs**: `swaggo/swag` → `docs/swagger/` (regenerate via `make swag`, never hand-edit)
- **Tests**: `stretchr/testify` + a real Postgres test DB (no SQLite mock)

## Phase progress (concise — CHECKPOINT.md has full detail)

Phases 0–5 done, live-verified, on `main`. Migration version 8.
Phase 6 (Attendance) is next.

## Local environment

- Postgres: Docker `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`
- Main DB: `exnodes_hrm`. Test DB: `exnodes_hrm_test`.
- Storage: optional at boot (`NewUploadService` does not ping S3). For attachment-upload live verification, recreate the MinIO container per `docs/superpowers/verification/phase-04.md` §3.

## Related memories in this directory

- [`resume_protocol.md`](resume_protocol.md) — exact session boot/resume order
- [`code_map.md`](code_map.md) — where things live in the source tree + key conventions
