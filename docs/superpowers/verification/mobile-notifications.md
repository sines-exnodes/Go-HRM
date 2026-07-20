# Verification — Mobile In-App Notifications (EP-005 / DR-MOB-005-001-01)

**Date:** 2026-07-20
**Branch:** `feat/mobile-notifications`
**Migration:** 000028 (`notifications`)
**Spec:** [`specs/2026-07-20-mobile-notifications-design.md`](../specs/2026-07-20-mobile-notifications-design.md)
**Plan:** [`plans/2026-07-20-mobile-notifications.md`](../plans/2026-07-20-mobile-notifications.md)

---

## 1. Migration round-trip

Against `exnodes_hrm_test`:

```
--- up ---      28/u notifications (159.83525ms)   → version 28
--- down 1 ---  28/d notifications (36.696541ms)   → version 27
--- up again ---28/u notifications (40.464041ms)   → version 28
```

No `dirty` state at any point. The down migration was exercised, not assumed.

Resulting schema (`\d notifications`):

```
Indexes:
    "notifications_pkey" PRIMARY KEY, btree (id)
    "idx_notifications_user_created" btree (user_id, created_at DESC) WHERE is_deleted = false
    "idx_notifications_user_unread" btree (user_id) WHERE read_at IS NULL AND is_deleted = false
    "uq_notifications_user_source" UNIQUE, btree (user_id, type, source_id) WHERE is_deleted = false
Foreign-key constraints:
    "notifications_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
Triggers:
    trg_notifications_set_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION set_updated_at()
```

`source_id` correctly carries **no** foreign key (DR Rule 13).

---

## 2. Test suite

`go test ./...` — every package `ok`.

Explicit counts for `internal/services` (`go test -v -count=1`):

```
PASS: 331
FAIL: 0
SKIP:
--- SKIP: TestUploadServiceLiveAWS (0.00s)
```

Baseline before this work was 314 passing, so **17 new tests**, zero regressions. The single skip is opt-in on `RUN_AWS_S3_INTEGRATION` (live AWS credentials).

> Counting passes and skips is deliberate. The leave-test bugs fixed in `1dd371b` shipped because a run with no `TEST_DATABASE_URL` skipped every DB-backed test and was recorded as "full suite green". `ok` is not evidence.

---

## 3. Live HTTP smoke

Server on `PORT=8082` (8080 is held by `ennam-kg-server` in this dev environment), `DB_NAME=exnodes_hrm_test`.

Two accounts: `smoke-admin@x.com` (Super Admin) and `smoke-emp@x.com` (Employee).

> Note: `POST /employees` does not accept a password — the created user must set one via the invite flow (`"Please set your password using the invite link sent to your email."`). For this smoke the hash was set directly using the project's own `utils.HashPassword`.

### Empty state (AC-11)

```
GET /mobile/notifications
{"success":true,"data":{"items":[],"total":0,"page":1,"page_size":50,"total_pages":0}}

GET /mobile/notifications/unread-count
{"success":true,"data":{"unread_count":0}}
```

Default `page_size` is 50 as designed.

### Draft announcement generates nothing (AC-10)

```
POST /announcements  → 722638cc-5bf8-4624-8bfe-34ac29702d85 (draft)
GET  /mobile/notifications → items: 0
```

### Publish notifies the audience

```
POST /announcements/{id}/publish → HTTP 200

GET /mobile/notifications
{
  "items": [{
    "id": "3ef80744-191d-44a1-9ea4-17835c1bf5cd",
    "type": "announcement",
    "title": "Office closed Friday",
    "body": "The office is closed this Friday.",
    "source_id": "722638cc-5bf8-4624-8bfe-34ac29702d85",
    "is_read": false,
    "created_at": "2026-07-20T15:03:56.66946+07:00"
  }],
  "total": 1, "page": 1, "page_size": 50, "total_pages": 1
}
```

The description was posted as `<p>The office is <b>closed</b> this Friday.</p>`; the stored body is plain text. `read_at` is absent (`omitempty`) and `is_read` is false — DR Rule 6, created unread.

### Leave approval notifies the submitter (AC-09)

```
POST /leave-requests → 32fcc6bf-9034-4388-88c5-2177c7b4197b
POST /leave-requests/{id}/approve (as admin) → HTTP 200

GET /mobile/notifications  → total: 2, newest first
 - leave_request | Leave Request Approved | Your leave request from 2026-08-03 to 2026-08-05 has been approved. | read: False
 - announcement  | Office closed Friday   | The office is closed this Friday.                                | read: False
```

> `from_date`/`to_date` must be full RFC3339 (`2026-08-03T00:00:00Z`); a bare `2026-08-03` returns 400.

### Mark read, idempotency, isolation (AC-05, AC-07, Rule 8, AC-01)

```
unread-count before                → {"unread_count":2}
POST /mobile/notifications/{id}/read
                                   → is_read: True  read_at: 2026-07-20T08:04:39.594677Z
unread-count after                 → {"unread_count":1}
POST .../read again (same user)    → HTTP 200      ← no-op, not 409 (Rule 8)
POST .../read as the OTHER user    → HTTP 404      ← indistinguishable from missing (AC-01)
GET /mobile/notifications as admin → total: 1, contains employee notif: False
GET /dashboard                     → unread_notification_count: 1   (Rule 14)
```

---

## 4. DB spot-check

```sql
SELECT n.user_id, u.email, n.type, n.source_id FROM notifications n JOIN users u ON u.id=n.user_id;
```

```
               user_id                |       email       |     type      |              source_id
--------------------------------------+-------------------+---------------+--------------------------------------
 acadc69c-2353-474b-9200-bcb8b1df0072 | smoke-admin@x.com | announcement  | 722638cc-...
 c1b77fd6-dc09-4ee9-b32c-dc2e450baa08 | smoke-emp@x.com   | announcement  | 722638cc-...
 c1b77fd6-dc09-4ee9-b32c-dc2e450baa08 | smoke-emp@x.com   | leave_request | 32fcc6bf-...
```

Two rows share a `source_id` — that is correct fan-out (the announcement targeted `all`, so both accounts received it), not duplication. Confirmed:

```
 rows | distinct_triples
------+------------------
    3 |                3
```

Every row is a distinct `(user_id, type, source_id)`, so `uq_notifications_user_source` is intact and DR Rule 5 holds.

`source_id` values match the originating announcement and leave request exactly (checked by equality against the captured IDs).

---

## 5. Swagger

`make swag` regenerated; `docs/swagger/swagger.json` contains 3 `mobile/notifications` path entries (list, unread-count, mark-read).

---

## 6. Not verified here

- **Announcement audience targeting by department / custom recipient list.** Only `all` was exercised live; the department and custom paths are covered by the pre-existing `resolveRecipientUserIDs` tests, not by this smoke.
- **Dev DB `exnodes_hrm` is still at migration 27.** This smoke ran entirely against `exnodes_hrm_test`. Applying 000028 to dev is a separate step.
- **Offline/cached-list behaviour (DR Alt 4)** is a client concern with no server surface.
- **Push and email dispatch** were disabled (`nil`) for the smoke — this DR covers the in-app list only (DR §8 out-of-scope).
