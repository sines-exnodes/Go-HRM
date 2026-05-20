---
name: project-hrm-go-migration
description: Key architectural facts and known issues for the exnodes-hrm-api-go-v2 Python→Go migration project, reviewed at Phases 0–3.
metadata:
  type: project
---

Python→Go HRM API migration (exnodes-hrm-api-go-v2). Owner: danny.tranhoang@exnodes.vn.

**Why:** Full rewrite, not a refactor. Go 1.24, Gin, GORM, Postgres, golang-migrate, JWT v5, bcrypt, aws-sdk-go-v2 (Supabase S3), swaggo.

**Phase status at review:** Phases 0–3 complete (92 commits), Phase 4 next.

## Known issues found in Phase 0–3 review (2026-05-18)

### Critical
1. **Refresh token skips session-invalidation check** — `AuthService.Refresh` never compares the token's `iat` against `users.email_changed_at` / `users.password_reset_at`. A refresh token issued before a password reset remains valid indefinitely until its own expiry.
   - File: `internal/services/auth_service.go:89–113`

2. **Avatar Content-Type check is client-spoofable** — `readAvatar` in `employee_handler.go` reads `fh.Header.Get("Content-Type")` from the multipart header, which is attacker-controlled. A malicious client can upload an SVG/HTML/script with `Content-Type: image/jpeg` to bypass the check. Correct fix: sniff bytes with `http.DetectContentType`.
   - File: `internal/handlers/employee_handler.go:82–84`

3. **Admin employee update (`Update`) is non-atomic for `is_active`** — `UpdateFields` on `employees` and `ToggleActive` on `users` are two separate DB writes outside a transaction; a crash between them leaves the two tables inconsistent.
   - File: `internal/services/employee_service.go:409–416`

4. **Missing admin user list/get routes** — `GET /users` and `GET /users/:id` are not registered in `main.go` despite `PermUsersRead` existing in the registry and permissions granted to Admin/HR Manager roles. Those roles cannot list or view individual users.
   - File: `cmd/server/main.go:132–136`

5. **`user_roles` hard-deleted instead of soft-deleted** — `ReplaceRoles` / `AssignRoles` run a raw `DELETE FROM user_roles WHERE user_id = ?` which destroys audit trail. The `user_roles` table has `is_deleted` + `deleted_at` columns for a reason.
   - Files: `internal/repositories/user_repo.go:138, 193`

### Important (correctness / convention)
- `DependentRepository` has no interface; `DependentService` and `EmployeeService` depend on the concrete struct pointer — not mockable.
- CORS is `Access-Control-Allow-Origin: *` (acceptable for dev, must harden before production).
- `EmployeeSoftDelete` IS atomic (uses a transaction) — correct.
- `toRead`/`toSummary` leaving `Department`/`Position` nil is a known projection gap (DB FKs are correct in migration 000005), confirmed cosmetic.
- bcrypt DefaultCost (10) is acceptable but below OWASP recommended 12 for new code.
- `employees.user_id` FK has `ON DELETE` unspecified (defaults to NO ACTION) — consistent with keeping orphan detection manual; not a bug but worth noting.
- Seed `password_hash` never logged — clean.

**How to apply:** Reference these findings when reviewing Phase 4+ branches to confirm the critical issues are resolved before merge.
