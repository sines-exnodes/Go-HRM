# Phase: Holiday Management — End-to-End Verification Log

**Date:** 2026-06-16  
**Branch:** main  
**Migration:** 000023 (holidays + holiday_templates tables)

---

## Environment

- DB: `postgres://postgres:devpassword@localhost:5432/exnodes_hrm` (Docker container `exnodes-hrm-postgres`, postgres:14-alpine)
- Server: `go run ./cmd/server/` with explicit `DB_HOST=localhost DB_PORT=5432 ...` env vars
- Admin: `admin@local.dev` / `admin123`

---

## Steps Completed

### 1. Format + Vet
```
gofmt -s -w .   → no changes
go vet ./...    → 0 errors
```

### 2. Tests
```
go test ./internal/services/... -timeout 300s
```
- `TestHolidayService_*` — ALL PASS (Create, Update, Delete, List, GetYears, ListTemplates, Import, recalc)
- Pre-existing failures (unrelated): `TestLeaveApprove_*` in `leave_approve_test.go` — column `line_manager_id` does not exist (DB has `manager_id`). These predate this work.

### 3. Migration
```
migrate -path migrations -database "postgres://...?sslmode=disable" up
→ no change (000023 already applied)
migrate version → 23
```

### 4. Swagger Regeneration
```
swag init -g "cmd/server/main.go" -o "docs/swagger" --parseDependency --parseInternal
→ success, docs/swagger/ updated
```

### 5. Server Startup
```
go run ./cmd/server/  (DB_HOST=localhost env override)
→ "exnodes-hrm-api listening on :8080 (env=development, swagger=true)"
```

---

## Endpoint Verification (all 7 endpoints)

### Auth
```
POST /api/v1/auth/login  →  200 OK, token received
```

### POST /api/v1/holidays (Create)
```json
// Request: {"year":2025,"name":"Tết Nguyên Đán","from_date":"2025-01-29T00:00:00Z","to_date":"2025-02-02T00:00:00Z"}
// Response 201: {"success":true,"message":"Holiday has been created","data":{"id":"...","year":2025,"name":"Tết Nguyên Đán","total_days":5,...}}
```
Duplicate name → 409 `{"success":false,"message":"a holiday named \"Tết Nguyên Đán\" already exists in 2025","code":"conflict"}`

### GET /api/v1/holidays?year=2025 (List)
```json
// Response 200: {"success":true,"data":{"items":[...],"total":1,"page":1,"page_size":20,"total_pages":1}}
```

### GET /api/v1/holidays/years (GetYears)
```json
// Response 200: {"success":true,"data":[2025,2026]}
```
(2026 injected as current year)

### GET /api/v1/holidays/templates?year=2025 (ListTemplates)
```json
// Response 200: {"success":true,"data":[...17 Vietnamese public holidays...]}
```

### POST /api/v1/holidays/import (Import)
```json
// Request: {"year":2025,"template_ids":["<uuid1>","<uuid2>"]}
// Response 200: {"success":true,"message":"2 holiday(s) imported for 2025","data":{"imported":2,"skipped":0}}
```

### PATCH /api/v1/holidays/:id (Update)
```json
// Request: {"name":"Test Holiday UpdateDelete RENAMED"}
// Response 200: {"success":true,"message":"Holiday has been updated","data":{"name":"Test Holiday UpdateDelete RENAMED",...}}
```
Not-found ID → 404 `{"success":false,"message":"holiday not found","code":"not_found"}`

### DELETE /api/v1/holidays/:id (Delete)
```json
// Response 200: {"success":true,"message":"Holiday has been deleted"}
```
Already-deleted ID → 404 `{"success":false,"message":"holiday not found","code":"not_found"}`

---

## Bug Found and Fixed

**Bug:** Typed nil `*AppError` in `error` interface — Go's interface nil semantics.

**Root cause:** In `holiday_handler.go`, `Update` and `Delete` used `aerr` (declared as `error` interface by `parseIDParam`). The `:=` reassignment from service methods returning `*apperrors.AppError` re-used the same `error`-typed variable. On success, a nil `*AppError` gets boxed into a non-nil `error` interface, so `aerr != nil` evaluated to `true`. The handler called `c.Error(aerr)` with a typed-nil, error middleware's `errors.As` extracted a nil `*AppError`, and `ae.HTTP` / `ae.Message` caused a nil pointer dereference panic (HTTP 500).

**Fix:** Renamed `parseIDParam` result from `aerr` to `err` in both `Update` and `Delete` handlers (matching the convention used in all other handlers: `employee_handler.go`, `user_handler.go`, `leave_handler.go`, `invite_handler.go`). The service call `:=` then freshly declares `aerr` as `*apperrors.AppError`, making the nil check correct.

**File:** `internal/handlers/holiday_handler.go` lines 93-97 (Update) and 125-129 (Delete).

**Verified:** After fix — `go build ./internal/handlers/...` clean, `go vet ./...` clean, PATCH 200 OK, DELETE 200 OK, error paths (not-found) correctly return 404.

---

## Result

**DONE** — All 7 endpoints verified end-to-end. Bug found during verification and fixed before close. No panics. No regressions.

### Note on test coverage gap
The service-level integration tests (`TestHolidayService_Update_*`, `TestHolidayService_Delete_*`) pass because they bypass the handler layer. The handler-level typed-nil bug was only catchable via end-to-end curl — confirming the necessity of the curl verification step.

### Note on two-session verification
This module was verified across two agent sessions (the first ran out of context). Create/List/GetYears/ListTemplates/Import were verified in session 1 (row names: `Tết Nguyên Đán`/2025). Update/Delete were verified in session 2 after applying the typed-nil fix (row name: `Test Holiday UpdateDelete`/2026). All 7 endpoints are covered; the data differs between sessions because the DB was live between them.
