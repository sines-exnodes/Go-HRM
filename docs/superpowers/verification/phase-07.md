# Phase 7 — Announcements + Mobile + SSE: End-to-End Verification Log

**Date:** 2026-05-20
**Branch:** `main`
**Migration version:** `10`
**Server:** `PORT=8082 go run ./cmd/server` (port 8080 was occupied by an
unrelated `ennam-kg-server` container; user opted for the port-swap rather
than stopping that container)
**Base URL:** `http://localhost:8082/api/v1`

---

## Summary

17 end-to-end steps exercising the full announcement surface (10 web
endpoints + 2 mobile + 1 SSE) plus a DB spot-check. All green. SSE
end-to-end (subscribe → publish → receive `announcement_published` event)
works as designed.

Highlights:

- `announcements.author_id` is `employees(id)` (verified at the DB level
  via `psql \d announcements`).
- `target_audience` enum is `('all','department')` only — 'custom' was
  dropped per REVISION NOTES #4. CHECK constraint applied.
- SSE event includes a JSON-marshaled `SSEAnnouncementPublishedEvent`
  inside the `data:` line; FE refetches the visible list on receipt
  (broadcast filter is nil, visibility is applied on the GET path).
- Non-admin reads are silently scoped to published + audience-match rows
  (visibility predicate evaluated at the SQL layer for List, in-process
  for Get).
- View tracking is idempotent (`ON CONFLICT (announcement_id, user_id)
  DO NOTHING` — preserves the FIRST view time).
- `JWTFromQueryOrHeader` middleware accepts `?token=` for SSE
  (EventSource limitation, REVISION NOTES #8). No token → 401.

---

## Endpoints exercised

| # | Endpoint | Auth | Status | Notes |
|---|---|---|---|---|
| 1 | `POST /auth/login` (super admin) | – | 200 | seeded |
| 2 | `POST /announcement-labels` (Phase 4) | admin | 200 | reused for label_ids |
| 3 | `POST /announcements` (draft + label) | admin | 201 | `status=draft, published_at=null` |
| 4 | `GET /sse/announcements?token=…` | admin | 200 | initial `event: connected` frame |
| 5 | `POST /announcements/:id/publish` | admin | 200 | `status=published`, SSE event received on the stream in <1s |
| 6 | `POST /announcements` (inline published) | admin | 201 | SSE event #2 received |
| 7 | `POST /employees` (non-admin) | admin | 201 | Reader user for visibility tests |
| 8 | `GET /announcements` (non-admin) | reader | 200 | `total=2`, both `status="published"` (draft hidden) |
| 9 | `GET /announcements/:draft` (non-admin) | reader | **403** | `"You cannot view this announcement"` |
| 10 | `POST /announcements/:id/view` × 2 | reader | 200 × 2 | idempotent |
| 11 | `GET /announcements/:id` (after mark) | reader | 200 | `has_viewed=true` |
| 12 | `POST /announcements` (non-admin) | reader | **403** | route-level RequirePerms; `missing: ["announcements:manage"]` |
| 13 | `PATCH /announcements/:id` (non-admin) | reader | **403** | same RequirePerms gate |
| 14 | `GET /mobile/announcements` | reader | 200 | `total=2`, only published, `has_viewed` per item |
| 15 | `GET /sse/announcements` (no token) | – | **401** | JWTFromQueryOrHeader correctly refused |
| 16 | `DELETE /announcements/:id` (admin) | admin | 200 | soft delete |
| 16b | `GET /announcements/:id` (after delete) | admin | **404** | `"Announcement not found"` |
| 17 | psql spot-check | – | – | see "DB spot-check" below |

---

## SSE end-to-end proof

After step 4 subscribed and step 5 published `46f55df1-…`:

```
event: connected
data: {"connection_id":"7c6ec0cd-36d7-47cc-8533-bb524300e5dd"}

event: announcement_published
data: {"type":"announcement_published","data":{"id":"46f55df1-be99-440a-bf1a-75a46e37ad4d","title":"Phase 7 draft","target_audience":"all","pinned":false,"published_at":"2026-05-20T18:11:37.660748+07:00"}}
```

The event arrives < 1 second after `POST /:id/publish` returns. Second
publish (step 6) produced a second event on the same stream — keep-alive
ticker held the connection open between the two.

---

## DB spot-check

```
                  id                  |     title     |  status   | target_audience | is_deleted | has_deleted_at | pinned
--------------------------------------+---------------+-----------+-----------------+------------+----------------+--------
 bbaa0f0a-504d-424c-a979-91860e2d7b7a | to delete     | published | all             | t          | t              | f
 9a84cd30-2dd6-4d48-8a57-e0cb9c718aab | Secret draft  | draft     | all             | f          | f              | f
 bd320cc1-303b-48ab-a512-76b8f5ae248c | Second        | published | all             | f          | f              | f
 46f55df1-be99-440a-bf1a-75a46e37ad4d | Phase 7 draft | published | all             | f          | f              | f
```

- "to delete" was step-16 admin DELETE → `is_deleted=t`,
  `deleted_at IS NOT NULL`. API GET returns 404 (step 16b).
- "Secret draft" stays as draft + `is_deleted=f` (admin-only-visible);
  reader can't list or get it.
- The 2 published rows surface to non-admin readers (step 8).

Child-table counts (live rows only):

```
      t       | live
--------------+------
 labels       |    1   ← step 3 attached P7-General to "Phase 7 draft"
 target_depts |    0   ← target_audience=all everywhere; no dept rows
 attachments  |    0   ← attachment upload path NOT exercised in this walk
 views        |    1   ← step 10 mark-viewed by reader on "Phase 7 draft"
```

---

## Test summary

```
$ go test ./internal/services -run 'TestAnnouncement_' -count=1
ok  github.com/exnodes/hrm-api/internal/services    43.938s
$ go test ./internal/sse -count=1
ok  github.com/exnodes/hrm-api/internal/sse          0.887s
```

21 announcement service tests + 7 SSE hub tests = 28 tests, all PASS.

## Operational note: port 8080 conflict

`ennam-kg-server` Docker container is currently bound to host port 8080.
Phase 7 verification ran on `PORT=8082` instead. CI / production wiring
is unaffected (the .env default stays at 8080); the swap was a local-
dev convenience.

## What's deferred

- **Multipart attachment upload endpoint** — model + repo + DB plumbing
  are in place (`announcement_attachments`), but `POST /announcements/
  :id/attachments` HTTP handler is intentionally not wired yet. BA will
  confirm whether attachments are a Phase 7 must-ship or a follow-up.
  Adding the route is a small commit: reuse the
  `http.DetectContentType` + MIME allowlist pattern from Phase 5
  `readLeaveAttachment`.
- **`target_audience='custom'`** — needs `announcement_target_users`
  join table; deferred until BA confirms (REVISION NOTES #4).
- **Scheduled publish cron** — the `scheduled_at` column is wired but
  no background job activates rows when their scheduled time arrives.
  Same hook-point as Phase 6's deferred auto-checkout cron.
- **SSE backplane for >1 replica** — `internal/sse/hub.go` is in-process
  only. Horizontal scaling needs Redis pub/sub (or NATS); package
  comment documents the constraint.
