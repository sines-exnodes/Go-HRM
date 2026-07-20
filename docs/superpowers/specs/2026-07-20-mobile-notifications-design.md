# Mobile In-App Notifications — Design

**Date:** 2026-07-20
**Source requirement:** [`DR-MOB-005-001-01`](../../../ba-requirements/docs/PLATFORMS/MOBILE-APP/EP-005-notifications/US-001-notifications/details/DR-005-001-01-notification-screen.md) v1.0 (status: draft)
**Status:** approved for planning
**Migration:** 000028

---

## 1. Scope

Build the in-app notification list for the mobile app: a per-employee, reverse-chronological feed of events generated elsewhere in the system, with per-notification read state.

**In scope:**

- `notifications` table + model + repository + service
- Three mobile endpoints: list, unread count, mark-read
- `unread_notification_count` added to the existing dashboard response
- Notification generation on announcement publish (DR Rules 2, 3)
- Notification generation on leave request approve/reject (DR Rule 4)

**Out of scope** (per DR §8): notification preferences and mute settings, delete/archive/hide, mark-all-as-read, grouping/filtering/search, OS-level push delivery, reverting to unread, notification types beyond the two below.

**Not our concern:** authoring announcements (WEB-APP EP-010) and making leave decisions (WEB-APP EP-002) already exist. This module consumes those events.

**Types in this release:** `announcement` and `leave_request`. The type is a stored, extensible value (DR Rule 15).

---

## 2. Approach — fan-out on write

One row per (recipient, event), carrying a **snapshot** of the title and body at creation time.

Two alternatives were considered and rejected:

- **Derived union at read time** (no table; `UNION` over visible announcements and own leave decisions, plus a separate read-marker table). Rejected: it breaks DR Rule 12 — editing an announcement would retroactively rewrite what the employee was told — and Rule 13, since deleting a source would make the notification vanish instead of showing "no longer available". Each new type would also mean editing the query rather than adding data, which is the opposite of Rule 15.
- **Thin rows, content resolved on read** (store only `user_id, type, source_id, read_at`; join to source for title/body). Rejected: inherits both Rule 12 and Rule 13 failures above, and adds an N+1 join across two differently-shaped source tables.

Fan-out costs one row per recipient per event — an announcement to 200 employees writes 200 rows. That is the correct trade here: reads vastly outnumber writes, and it turns three DR rules into table facts rather than query logic.

---

## 3. Data model

### Migration `000028_notifications`

```sql
CREATE TABLE notifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id),
    type       TEXT        NOT NULL,
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL DEFAULT '',
    source_id  UUID        NOT NULL,
    read_at    TIMESTAMPTZ,
    is_deleted BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC) WHERE is_deleted = FALSE;

CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id) WHERE read_at IS NULL AND is_deleted = FALSE;

CREATE UNIQUE INDEX uq_notifications_user_source
    ON notifications (user_id, type, source_id) WHERE is_deleted = FALSE;

CREATE TRIGGER trg_notifications_set_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### Design decisions

**`user_id → users(id)`, not `employees(id)`.** This is the documented exception to the schema-split rule: notifications are an auth-level surface consumed by the logged-in session, exactly like `announcement_views` and `device_tokens`. It also matches what the code already produces — `AnnouncementService.resolveRecipientUserIDs` returns user IDs.

**`source_id` carries no foreign key.** It points into two different tables depending on `type`, and Rule 13 requires the notification to survive its source's deletion. A FK would either be impossible or would cascade the row away.

**`type` is plain TEXT with no CHECK constraint.** Rule 15 wants a new type to be a data change, not a schema change. Validation lives in Go as a `NotificationType` constant set, mirroring `AnnouncementStatus`.

**Read state is `read_at TIMESTAMPTZ`, not a boolean.** Same storage, but it records *when*, and the one-way transition of Rule 8 becomes `WHERE read_at IS NULL` — a no-op on replay rather than a flag that could be flipped back.

**The unique index is what makes Rule 5 true** rather than hoped-for. Producers insert with `clause.OnConflict{DoNothing: true}` (the repo's existing idempotent-marker convention), so re-publishing an announcement or a retried approve silently no-ops instead of duplicating.

### Model

`internal/models/notification.go` — `Notification` embedding `BaseModel`, plus:

```go
type NotificationType string

const (
    NotificationTypeAnnouncement NotificationType = "announcement"
    NotificationTypeLeaveRequest NotificationType = "leave_request"
)
```

---

## 4. Producers

Both producers share one `NotificationService.CreateMany`. The leave notifier is new and lives in `internal/services/notification_notifier.go`; the announcement producer is an extension of the existing `announcementNotifier`, so it stays in `internal/services/announcement_notifier.go` beside the push and email dispatch it runs alongside.

### Announcements — no change to `AnnouncementService`

The existing `AnnouncementNotifier` interface ([`announcement_service.go:29-31`](../../../internal/services/announcement_service.go)) is already the right seam. The concrete `announcementNotifier` gains a `notifications` collaborator beside `push` and `email`; `NotifyAnnouncement` writes one row per recipient before dispatching push and email.

AC-10 (drafts generate nothing) is already satisfied upstream: `broadcastPublished` returns early when `PublishedAt == nil`, and `dispatchNotifications` is only reachable from there.

Announcement bodies pass through the existing `plainTextPreview` helper so the stored snapshot is already plain text. Storing raw HTML per recipient to render three clamped lines would be wasteful.

### Leave — new seam

`LeaveService` has no notifier today, so it gets one in the same shape:

```go
type LeaveNotifier interface {
    NotifyLeaveDecision(ctx context.Context, employeeID, leaveID uuid.UUID,
        approved bool, from, to time.Time)
}
```

Nil-safe optional field, called from `Approve` and `Reject` **after** `transitionStatus` returns successfully — a failed transition must not notify. The implementation resolves `employee_id → user_id` and writes the row.

Copy per DR §3 and AC-09:

| Field | Approved | Rejected |
|---|---|---|
| Title | `Leave Request Approved` | `Leave Request Rejected` |
| Body | `Your leave request from {from} to {to} has been approved.` | `…has been rejected.` |

Cancellation generates nothing (Rule 4, flagged as an open question in the DR).

---

## 5. API surface

Three endpoints under the existing mobile group, alongside `/mobile/announcements`:

```
GET  /api/v1/mobile/notifications              ?page&page_size
GET  /api/v1/mobile/notifications/unread-count
POST /api/v1/mobile/notifications/:id/read
```

### No new permission

JWT-only, following the `GET /api/v1/dashboard` precedent. AC-01 is an *ownership* constraint, not a permission one — every authenticated employee is entitled to their own notifications by definition. A `notifications:read` permission would have to be seeded to all five roles immediately, making it a permission that can never be false.

**Ownership is enforced in the service by scoping every query to the JWT's `user_id`.** Mark-read scopes its `UPDATE` by `id AND user_id`, not by `id` alone: a request for another employee's notification ID affects zero rows and returns 404, indistinguishable from a genuinely missing row, leaking nothing.

Mark-read returns `200` with the updated `NotificationRead` in the standard `Response[T]` envelope, so the client can re-render the row without a refetch. Marking an already-read notification is a `200` no-op that returns the row with its original `read_at` untouched — not a `409`. Rule 8 makes read terminal, so a repeat is a successful arrival at the intended state, and the mobile client may legitimately retry after a dropped response.

### Pagination — deviation from the DR

DR §8 says "no pagination or infinite scroll", and Rule 11 retains notifications indefinitely. Together that is an unbounded response that grows forever.

**Decision:** paginate with `page` / `page_size`, default 50, max 100, using the repo's standard `PaginatedData` envelope. The FE can ignore paging and render page 1, so the DR's continuous-scroll UX still works. **Flag for BA** as a deliberate deviation.

### DTOs — `internal/dto/notification.go`

```go
type NotificationListQuery struct {
    Page     int `form:"page"`
    PageSize int `form:"page_size"`   // default 50, max 100
}

type NotificationRead struct {
    ID        uuid.UUID  `json:"id"`
    Type      string     `json:"type"`
    Title     string     `json:"title"`
    Body      string     `json:"body"`
    SourceID  uuid.UUID  `json:"source_id"`
    IsRead    bool       `json:"is_read"`
    ReadAt    *time.Time `json:"read_at,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
}

type NotificationUnreadCountRead struct {
    UnreadCount int64 `json:"unread_count"`
}
```

No filter or sort parameters — Rule 16 makes the screen read-and-navigate only, and Rule 10 fixes the sort. Adding them now would be speculative surface.

### Deliberate response omissions

**No icon or type-label field.** The API returns `type`; the mobile app holds the icon/label registry from DR §3. Sending `"SpeakerHigh"` over the wire would couple the backend to a client icon set. Honest cost: a genuinely new type needs a mobile release to render its icon, so AC-20 holds for the *layout* but not for zero-client-change type addition. A server-side registry is the alternative if BA wants the stronger reading — see Open Questions.

**Timestamps are RFC3339 UTC.** AC-16's `"20 July, 2026 - 10:00 AM"` in employee-local time is client-side formatting; the server has no reliable notion of the employee's timezone.

### Dashboard widget

`unread_notification_count` is added to the existing `GET /dashboard` response so the header bell renders on first paint without a second request (Rule 14). The field is named in full because at the dashboard root, a bare `unread_count` does not say unread *what* — the dedicated endpoint keeps the short `unread_count` since its envelope already supplies the context. That endpoint serves cheap refreshes after a mark-read, avoiding a full dashboard refetch.

---

## 6. Testing

Integration tests against the Postgres test DB, matching `holiday_service_test.go` in shape. The tests that encode business intent rather than mechanics:

| Test | Encodes |
|---|---|
| Data isolation | A's list excludes B's rows; A marking B's ID returns 404 **and leaves B's row unread** |
| Rule 5 idempotency | Re-publishing an announcement and re-approving a request each leave exactly one row |
| Rule 8 one-way | Marking an already-read notification is a no-op success; `read_at` does not move |
| AC-10 draft | A draft announcement generates zero rows for anyone |
| AC-09 copy | Approve and reject produce the specified titles and date-range bodies |
| Rule 12 snapshot | Editing an announcement's title after publish leaves the notification text unchanged |
| Rule 13 orphan | Deleting the source announcement leaves the notification listed and readable |
| Ordering + pagination | Newest first across a page boundary |

The Rule 5 test is the one that fails if someone later drops the unique index. The Rule 12 test directly encodes why fan-out was chosen over a derived view.

**Live HTTP smoke** (required before this counts as done per the repo's verification bar): publish an announcement, approve a leave request, list as the recipient, mark one read, re-list, check the unread count, spot-check rows in the DB. Log to `docs/superpowers/verification/`.

### Baseline — fixed before this work starts

The services suite used to be red on `main` with 4 failing leave tests. Since this module modifies `LeaveService.Approve` and `Reject`, those were diagnosed and fixed first on `fix/leave-test-failures` (`1dd371b`), which must be merged before implementation begins.

CHECKPOINT.md described them as one failure mode with a shared error message. They were actually **two unrelated bugs, both test-side, neither ever passing**:

1. `setupApproveChain` wrote to `employees.line_manager_id`, a column that has never existed — it is `manager_id` (migration 000003); `line_manager_id` is the *feature* name from PR #10. Postgres rejected the UPDATE with SQLSTATE 42703, so all three approve/reject scope tests died in the helper before reaching any approval logic. They never produced the error CHECKPOINT.md attributed to them.
2. `TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus` patched as a non-admin. The Approved→Pending revert it guards is admin-only — a non-admin owner is refused earlier by `Update`'s documented `approved (owner) -> 403` contract. The test hit that 403 and proved nothing about G3. Fixed by patching as admin, then verified meaningful by mutation: disabling the no-op guard fails it on exactly the revert assertion.

Both date to the commits that introduced them (`0132dd8`, `d632a69`), whose verification recorded "full suite green (all PASS or SKIP)" — from a run with no `TEST_DATABASE_URL`, where every DB-backed test skipped. **Count passes and skips explicitly; `ok` is not evidence.**

Post-fix baseline: 314 pass, 0 fail, 1 skip (`TestUploadServiceLiveAWS`, opt-in on `RUN_AWS_S3_INTEGRATION`).

---

## 7. Open questions

| Question | Owner | Impact |
|---|---|---|
| Pagination deviates from DR §8's "no pagination" — confirm acceptable | BA | Low; FE can ignore paging |
| Icon/label registry client-side vs server-side (AC-20 strong reading) | BA / Design | Medium; changes the response shape if server-side |
| Retention/purge policy (DR Rule 11 assumes indefinite) | Product Owner | Low now, grows with table size |
| Leave cancellation generating no notification (DR Rule 4) | Product Owner | Low |
| REQUIREMENTS.md not yet authored for US-001; `parent_requirement` is the proposed FR-US-001-01 | BA | None on implementation |

---

## 8. File inventory

| Path | Change |
|---|---|
| `migrations/000028_notifications.up.sql` / `.down.sql` | new |
| `internal/models/notification.go` | new |
| `internal/dto/notification.go` | new |
| `internal/repositories/notification_repo.go` | new |
| `internal/services/notification_service.go` | new |
| `internal/services/notification_notifier.go` | new — leave producer + shared copy helpers |
| `internal/services/notification_service_test.go` | new |
| `internal/handlers/notification_handler.go` | extend — existing file owns `/notifications/test` |
| `internal/services/announcement_notifier.go` | extend — add notifications collaborator |
| `internal/services/leave_service.go` | extend — `LeaveNotifier` seam + calls in Approve/Reject |
| `internal/services/dashboard_service.go` | extend — `unread_notification_count` |
| `internal/dto/dashboard.go` | extend |
| `cmd/server/main.go` | wire + 3 routes |
| `docs/swagger/` | regenerate via `make swag` |
