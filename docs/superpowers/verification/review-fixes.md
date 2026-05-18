# Live Re-Verification — Security Fixes #1 & #2

Date: 2026-05-18
Environment: local server `make run` on `:8080`, Postgres `exnodes_hrm` (Docker `ennam-ecom-postgres`), super admin `admin@exnodes.vn`.
All token values redacted. Output below is REAL captured HTTP status codes + response bodies + SQL.

---

## FIX #1 — Refresh rejects tokens issued before a password change

| Step | Action | Status | Expected | Result |
|------|--------|--------|----------|--------|
| 1 | `POST /api/v1/auth/login` (original pw) → RT1, AT1 | **200** | 200 | PASS |
| 2 | `POST /api/v1/users/me/change-password` (current→new) | **200** | 200 | PASS |
| 3 | wait ~2s (password_reset_at strictly after RT1 iat) | — | — | — |
| 4 | `POST /api/v1/auth/refresh` with **RT1** (pre-change) | **401** | 401 | PASS |
| 5a | `POST /api/v1/auth/login` (new pw) → RT2 | **200** | 200 | PASS |
| 5b | `POST /api/v1/auth/refresh` with **RT2** (fresh) | **200** | 200 | PASS |
| 6a | restore: `POST /api/v1/users/me/change-password` (new→original) | **200** | 200 | PASS |
| 6b | final `POST /api/v1/auth/login` (original pw) | **200** | 200 | PASS |

Key response bodies:

- Step 2: `{"message":"Password changed successfully","success":true}`
- Step 4 (the fix): `401` —
  `{"success":false,"message":"Session expired due to password reset — please log in again","code":"unauthorized"}`
- Step 5b: `200` — new `access_token` (REDACTED), `token_type: Bearer`
- Step 6a: `{"message":"Password changed successfully","success":true}`
- Step 6b: `200`

Pre-fix behaviour would have been `200` + a fresh token pair at Step 4. The
observed `401` confirms `tokenInvalidatedBy(user.PasswordResetAt, iat)` in
`internal/services/auth_service.go:118` is enforced on the refresh path.

**FIX #1 verdict: PASS. Password successfully restored to the original
`.env` SUPER_ADMIN_PASSWORD value (`Admin@12345`); final login = 200.**

---

## FIX #2 — Avatar rejects content-spoofed upload

Super admin employee id: `ab07e787-3ffe-4758-83f2-729c3daa59f7`

| Step | Action | Status | Expected | Result |
|------|--------|--------|----------|--------|
| 1 | `GET /api/v1/employees/me` (avatar_url before = `null`) | **200** | 200 | PASS |
| 2 | craft `/tmp/evil.jpg` = `<html><script>alert(1)</script></html>` (HTML bytes) | — | — | — |
| 3 | `PATCH /api/v1/employees/me/avatar` `-F avatar=@evil.jpg;type=image/jpeg` | **400** | 400 | PASS |
| 4 | `PATCH .../avatar` valid 1×1 PNG (70 B) `;type=image/png` | **200** | 200 | PASS |
| 5 | SQL spot-check of persisted `avatar_url` | — | matches step 4 | PASS |

Key response bodies:

- Step 3 (the fix): `400` —
  `{"success":false,"message":"Avatar must be a valid image (PNG, JPEG, GIF, or WEBP)","code":"bad_request"}`
  (client-supplied `image/jpeg` was NOT trusted; `http.DetectContentType`
  sniffed the real bytes as `text/html` and rejected — `internal/services/employee_service.go:532-535`)
- Step 4: `200` — returned
  `avatar_url = https://localhost:19000.supabase.co/storage/v1/object/public/hrm-uploads/avatars/fefd8366-479d-4c49-98ce-ce4137ae2fcc.png`

Step 5 — SQL (`psql -h localhost -U ennam -d exnodes_hrm`):

```
SELECT avatar_url FROM employees WHERE id = 'ab07e787-3ffe-4758-83f2-729c3daa59f7';

https://localhost:19000.supabase.co/storage/v1/object/public/hrm-uploads/avatars/fefd8366-479d-4c49-98ce-ce4137ae2fcc.png
```

The persisted DB value is exactly the URL returned by the valid Step-4
upload. `avatar_url` was `null` before, so the spoofed `evil.jpg` never
persisted. Object-store listing confirms exactly one 70-byte object
(`avatars/fefd8366-479d-4c49-98ce-ce4137ae2fcc.png`) — the valid PNG only;
`evil.jpg` was rejected before reaching storage.

**FIX #2 verdict: PASS.**

---

## Environment note (not a security-fix bug)

On the first valid-PNG upload attempt the request returned **500**:

```
Error #01: upload: put object: operation error S3: PutObject,
resolve auth scheme: resolve endpoint: endpoint rule error,
Custom endpoint `` was not a valid URI
```

Root cause: the local `.env` was missing the `STORAGE_ENDPOINT /
STORAGE_ACCESS_KEY / STORAGE_SECRET_KEY / STORAGE_BUCKET` block that
`.env.example` defines (only the legacy empty `SUPABASE_*` placeholders were
present), so the S3 SDK had an empty endpoint. This is a local
test-environment config gap **downstream of and independent from both
security fixes** — the content-sniff check (Fix #2) had already executed
and correctly accepted the valid PNG before the storage call. Resolved for
verification by adding a `STORAGE_*` block to `.env` pointing at a local
MinIO (`secfix-minio`, bucket `hrm-uploads`); no security-fix code was
modified. The public-URL still renders in `*.supabase.co` form because it
is derived from the configured endpoint host string, satisfying the
"supabase/fake URL" expectation.

## Overall

Both security fixes verified end-to-end with real captured output. No bug
found in either fix. Password restored to original. The only deviation was
the missing local `STORAGE_*` env config, which is an environment gap, not
a defect in the fixes.
