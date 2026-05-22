# Phase 9 — Email + Invite + Push: End-to-End Verification Log

**Date:** 2026-05-22
**Branch:** `main`
**Migration version:** `12`
**Server:** `PORT=8082 SMTP_HOST=localhost SMTP_PORT=11025 ... go run ./cmd/server`
**Mailpit:** `docker run -d --rm -p 11025:1025 -p 18025:8025 --name mailpit-phase09 axllent/mailpit`
**Base URL:** `http://localhost:8082/api/v1`
**Mailpit UI:** `http://localhost:18025`

---

## Summary

23 end-to-end steps exercising all 7 new endpoints (5 admin invite + 1
public accept + 1 push test) plus the SMTP-via-Mailpit pipeline and a
DB spot-check. All green.

Highlights:

- **End-to-end SMTP delivery via Mailpit** verified — invite creation
  produces a real email visible at `http://localhost:18025`. The
  template renders with `AppName`, `FullName`, `AcceptURL`, and
  `ExpiresAt` populated. Both HTML and plain-text alternatives.
- **Token round-trip works** — extracted from the email body, accepted
  via the public endpoint, creates a real user that can log in.
- **Public accept endpoint** is genuinely unauthenticated — no
  `Authorization` header was sent and the call succeeded.
- **Permission gating** verified live: non-admin (`newbie09`, Employee
  role) gets `403 forbidden` with `missing: ["invites:manage"]` on
  `POST /invites`. Same shape for `/notifications/test` →
  `missing: ["users:manage_roles"]`.
- **`PermInviteManage`** seed merge worked on boot — the Super Admin's
  wildcard already covered the test, but the constant + group are now
  exposed via `GET /roles/permissions`.
- **SMTP graceful degradation** path (unit-tested) records
  `last_email_error` on the row; with Mailpit configured live, the
  e2e walk saw `last_email_error=null` (real delivery happened) and
  the resend message count incremented to 2.
- **`/notifications/test` no-op path** — FCM stays disabled in dev
  (no `FIREBASE_CREDENTIALS_PATH`), so registered tokens count as
  `skipped` and 0 actual pushes go out. Endpoint returns 200 with the
  diagnostic envelope.

---

## Endpoints exercised

| # | Endpoint | Auth | Status | Notes |
|---|---|---|---|---|
| 1 | `POST /auth/login` (super admin) | – | 200 | seeded |
| 2 | `DELETE /api/v1/messages` (Mailpit) | – | 200 | clear inbox |
| 3 | `POST /invites` (basic) | admin | 201 | `last_email_error=null` |
| 4 | `GET /api/v1/messages` (Mailpit) | – | 200 | 1 message to `newbie09@example.com`, subject "You're invited to Exnodes HRM" |
| 5 | `GET /api/v1/message/:id` (Mailpit) | – | 200 | token extracted from body (len=43, matches base64-url(32 bytes)) |
| 6 | `GET /invites/:id` | admin | 200 | `status="pending"`, inviter resolved to "Super Admin" |
| 7 | `GET /invites` | admin | 200 | `total=1`, list shape correct |
| 8 | `POST /invites/:id/resend` | admin | 200 | second email arrives (Mailpit total=2), token unchanged |
| 9 | `POST /invites/accept` (PUBLIC) | – | 200 | `user_id` returned, `message="Account created…"` |
| 10 | `POST /auth/login` (newly created user) | – | 200 | `has_token=true` — empSvc.Create + ReplaceRoles worked |
| 11 | `GET /invites/:id` (after accept) | admin | 200 | `status="accepted"`, `accepted_user_id` matches step 9 |
| 12 | `POST /invites/accept` (replay) | – | **409** | `"Invite has already been used"` |
| 13 | `POST /invites/accept` (unknown token) | – | **404** | `"Invite not found"` |
| 14 | `POST /invites/accept` (short password) | – | **400** | binding-tag `min=8` |
| 15 | `POST /invites` + `DELETE /:id` | admin | 201 + 200 | revoke soft-deletes; subsequent GET → 404 |
| 16 | `POST /invites` for existing user email | admin | **409** | `"A user with this email already exists"` |
| 17 | `POST /invites` duplicate pending | admin | **409** | `"A pending invite for this email already exists"` |
| 18 | `POST /invites` (non-admin) | newbie | **403** | `missing: ["invites:manage"]` |
| 19 | `POST /invites/accept` (empty body) | – | **400** | binding-tag `required` |
| 20 | `POST /notifications/test` (no devices) | admin | 200 | `sent=0, skipped=0` |
| 21 | `POST /users/me/device-tokens` + `POST /notifications/test` | admin | 200 + 200 | `sent=0, skipped=1` (FCM disabled) |
| 22 | `POST /notifications/test` (non-admin) | newbie | **403** | `missing: ["users:manage_roles"]` |
| 23 | psql spot-check | – | – | see "DB spot-check" below |

---

## Mailpit delivery proof

```text
$ curl -fsS http://localhost:18025/api/v1/messages
{
  "total": 2,
  "messages": [
    {
      "To": [{"Address":"newbie09@example.com"}],
      "Subject": "You're invited to Exnodes HRM",
      "Snippet": "Welcome to Exnodes HRM Hi New User, You've been invited to join Exnodes HRM. Cli…"
    },
    ...
  ]
}
```

Token extracted with `grep -oE 'token=[^[:space:]]+'` from the plain-text
alternative — 43 characters of URL-safe base64 (32 random bytes), as
the service intended.

## DB spot-check

```
         email          | accepted | is_deleted | has_email_err | has_user_link
------------------------+----------+------------+---------------+---------------
 newbie09@example.com   | t        | f          | f             | t
 revokeme09@example.com | f        | t          | f             | f
 dup09@example.com      | f        | f          | f             | f
```

- `newbie09` — fully accepted: `accepted_at` set, `accepted_user_id`
  populated, no email error (real Mailpit delivery succeeded).
- `revokeme09` — soft-deleted (step 15): `is_deleted=true`. API GET
  returns 404. The DB row is preserved for audit but invisible to the
  service layer.
- `dup09` — created in step 17 but the duplicate attempt was rejected
  before insert; the original pending row is still there.

## Push notification path

FCM is intentionally disabled in this dev environment
(`FIREBASE_CREDENTIALS_PATH=""`). The `NewPushClient` constructor logs:

```
push: FCM disabled — FIREBASE_CREDENTIALS_PATH or FIREBASE_PROJECT_ID is empty
```

…and returns the no-op client. The `/notifications/test` endpoint
remains live and reports `sent=0, skipped=<token-count>`. Once a
service-account JSON is supplied in production, the same endpoint will
deliver real FCM payloads without code changes — `IsConfigured()`
flips to `true` and `Send()` calls the FCM HTTP v1 endpoint.

## Test summary

```
$ go test ./internal/services -run 'TestInvite_|TestPush_|TestEmail_' -count=1
ok  github.com/exnodes/hrm-api/internal/services    9.855s
```

17 service tests (11 invite + 4 push + 2 email), all PASS. Full project
suite remains green:

```
$ go test ./... -count=1
ok  github.com/exnodes/hrm-api/internal/permissions  0.352s
ok  github.com/exnodes/hrm-api/internal/services    73.188s
ok  github.com/exnodes/hrm-api/internal/sse          1.218s
ok  github.com/exnodes/hrm-api/pkg/utils             1.172s
```

## Operational notes

- **Port 8080 conflict (carried over from Phases 7-8)**: `ennam-kg-server`
  container holds host port 8080 locally. Phase 9 verification ran on
  `PORT=8082`. CI default stays 8080.
- **Mailpit on 11025/18025**: avoids the typical 1025/8025 collision
  with other dev SMTP installs.
- **FCM production rollout**: set `FIREBASE_CREDENTIALS_PATH` to a
  service-account JSON path + `FIREBASE_PROJECT_ID` to the GCP project.
  The PushClient hot-swaps to the real FCM impl at boot.

## What's deferred

- **Password-reset email flow** — Python source has it; Phase 9 only
  ships invites + push. Reuse the EmailService + a new
  `password-reset.html` template if/when BA confirms the scope.
- **`/invites/accept` rate-limiting** — public endpoint that creates
  users; consider IP-based rate limit (reverse proxy or middleware)
  before production exposure.
- **FCM topic-based fanout** — current impl is per-device. Topic
  subscription (e.g. "all-android", "department-eng") would need a
  small extension to the `PushMessage` shape + a topic field in
  `device_tokens`.
- **Email retry / queue** — current path is best-effort fire-and-record.
  A real retry queue (with backoff) is out of scope.
