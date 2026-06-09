# Announcements Parity Audit — Python ↔ Go

**Date:** 2026-06-09
**Status:** Decisions locked — ready for implementation plans
**Migration:** `main` — no new migration needed
**Plans:** Plan A (G1+G3) · Plan B (G2)

---

## Full API Inventory

### Python web (`/announcements`) — all endpoints require `ANNOUNCEMENTS_MANAGE`

| # | Method | Path | Notes |
|---|--------|------|-------|
| W1 | GET | `/announcements/` | Filters: `search`, `status` (`draft\|sent`) |
| W2 | POST | `/announcements/` | Body: `{title, description, everyone, recipient_ids, label_id, send_now}` |
| W3 | GET | `/announcements/{id}` | |
| W4 | PATCH | `/announcements/{id}` | **409 if `status == sent`** |
| W5 | POST | `/announcements/{id}/send` | Optional body `{everyone?, recipient_ids?}` — atomic audience update + send |
| W6 | DELETE | `/announcements/{id}` | Hard delete |

### Python mobile (`/mobile/announcements`) — auth-only

| # | Method | Path | Notes |
|---|--------|------|-------|
| M1 | GET | `/mobile/announcements/` | Top-5 home widget; returns `{id, title, description, sent_at, is_read, label}` |
| M2 | GET | `/mobile/announcements/list` | Paginated; same shape as M1 |
| M3 | GET | `/mobile/announcements/{id}` | Full `MobileAnnouncementRead` (includes `description`) |
| M4 | POST | `/mobile/announcements/{id}/read` | Mark as read (idempotent) |

### Go web (`/api/v1/announcements`)

| # | Method | Path | Permission | Notes |
|---|--------|------|------------|-------|
| W1 | GET | `/api/v1/announcements` | auth | Visibility-filtered; filters: `search`, `status`, `label_id`, `pinned`, `scope`, `department_id` |
| W2 | POST | `/api/v1/announcements` | `announcements:manage` | Body: `{title, description, summary?, status?, scheduled_at?, target_audience?, pinned?, cover_image_url?, label_ids?, department_ids?, recipient_ids?, send_now?}` |
| W3 | GET | `/api/v1/announcements/:id` | auth | 403 if row not visible to caller |
| W4 | PATCH | `/api/v1/announcements/:id` | `announcements:manage` | **Currently allows editing published rows** |
| W5 | POST | `/api/v1/announcements/:id/publish` | `announcements:manage` | No body; SSE broadcast only — no push/email |
| W6 | DELETE | `/api/v1/announcements/:id` | `announcements:manage` | Soft-delete |
| W7 | POST | `/api/v1/announcements/:id/view` | auth | Mark viewed (idempotent) |

### Go mobile (`/api/v1/mobile/announcements`) — auth-only

| # | Method | Path | Notes |
|---|--------|------|-------|
| M1 | GET | `/api/v1/mobile/announcements` | Top-5 `MobileAnnouncementBrief` (no `description`) |
| M2 | GET | `/api/v1/mobile/announcements/list` | Paginated `MobileAnnouncementBrief` |
| M3 | GET | `/api/v1/mobile/announcements/:id` | Full `AnnouncementRead` (richer than Python) |

### Go SSE (Python has none)

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/sse/announcements` | JWT via `?token=` |

---

## Gap Analysis

| Gap | Python | Go | Decision |
|-----|--------|----|----------|
| **G1** Mobile mark-as-read | `POST /mobile/:id/read` | **MISSING** | **D1 = A: add route alias** |
| **G2** Push + email on publish | `_dispatch_notifications` fires email + FCM after send | **MISSING** — SSE only | **D2 = A: implement full dispatch** |
| **G3** Edit-after-publish guard | 409 if `status == sent` | Allows editing published rows | **D3 = A: add 409 guard** |
| G4 | List/Get require `ANNOUNCEMENTS_MANAGE` | Auth-only (visibility-filtered) | Intentional Go design — not a gap |
| G5 | `/send` + optional audience body | `/publish` + no body | Intentional — not a gap |
| G6 | No attachments | Model + DTO present, handler deferred | Same state vs Python — not a gap |
| G7 | `is_read` field name | `has_viewed` field name | Rename — mobile client update only, not a backend gap |
| G8 | `description` in mobile list | Omitted from brief (in detail only) | Intentional Go optimization — not a gap |

---

## Locked Decisions

### D1 — Mobile mark-as-read (G1) = **Route alias**

Register `POST /api/v1/mobile/announcements/:id/read` that calls the existing `MarkViewed`
handler. Zero new logic — the handler already writes to `announcement_views` (idempotent,
`ON CONFLICT DO NOTHING`) and is auth-only. Mobile clients get their canonical path; web
clients keep `/view`. Both write to the same table.

### D2 — Push + email on publish (G2) = **Full dispatch via `AnnouncementNotifier` interface**

Add a `AnnouncementNotifier` interface to `announcement_service.go`:

```go
type AnnouncementNotifier interface {
    NotifyAnnouncement(ctx context.Context, userIDs []uuid.UUID, title, description string)
}
```

- `nil` is valid (tests, same nil-safe pattern as `HubBroadcaster`).
- `broadcastPublished` launches `go s.dispatchNotifications(ann)` after the SSE broadcast.
- `dispatchNotifications` calls `resolveRecipientUserIDs` (translates `target_audience` +
  preloaded join rows to `[]uuid.UUID` user IDs using the existing `emps` repo), then calls
  `s.notifier.NotifyAnnouncement` with `context.Background()`.
- Concrete implementation in new file `internal/services/announcement_notifier.go` wraps
  `PushNotificationService` + `EmailService` + `UserRepository`. Errors are logged, never
  propagated — publish already succeeded.
- `EmailService` gains `SendAnnouncementNotification(ctx, toEmail, title, description string) error`
  (same HTML-template pattern as `SendInviteEmail`).
- The `announcementNotifier` calls the existing `PushNotificationService.SendToUser(ctx, uid, dto.NotificationTestRequest{Title: title, Body: description})` directly — no new push method needed.
- `main.go`: construct `annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo)`
  and pass to `NewAnnouncementService`.

**Recipient resolution by audience:**

| `target_audience` | Query |
|-------------------|-------|
| `all` | `emps.FindAllActive(ctx)` → collect `emp.UserID` |
| `department` | `emps.FindByDepartmentIDs(ctx, deptIDs)` → collect `emp.UserID` |
| `custom` | `emps.FindByIDs(ctx, empIDs)` → collect `emp.UserID` |

`FindAllActive`, `FindByIDs`, and `FindByDepartmentIDs` must be added to the `EmployeeRepository` interface and impl — **none of the three currently exist**.

### D3 — Edit-after-publish guard (G3) = **409 on published/archived**

In `AnnouncementService.Update`, immediately after fetching the row:

```go
if row.Status == models.AnnouncementStatusPublished ||
   row.Status == models.AnnouncementStatusArchived {
    return nil, apperrors.ErrConflict("Cannot edit a published or archived announcement")
}
```

Matches Python's guard. `archived` is included because it is a terminal state and editing
it makes no sense semantically.

---

## Implementation Plans

### Plan A — G3 + G1 (trivial changes, no new deps)

| Task | File | Change |
|------|------|--------|
| T1 | `internal/services/announcement_service.go` | Add status guard at top of `Update` |
| T2 | `cmd/server/main.go` | Add `mobileAnnounce.POST(":id/read", announcementH.MarkViewed)` |

Tests: existing service suite covers the guard via status-gated operations. Route alias
has no logic to test — verified by smoke `curl`.

### Plan B — G2 (new deps, goroutine dispatch, tests)

| Task | File | Change |
|------|------|--------|
| T1 | `internal/services/announcement_service.go` | Add `AnnouncementNotifier` interface + `notifier` field + constructor param + `dispatchNotifications` + `resolveRecipientUserIDs` |
| T2 | `internal/repositories/employee_repo.go` | Add `FindAllActive(ctx)`, `FindByIDs(ctx, []uuid.UUID)`, `FindByDepartmentIDs(ctx, []uuid.UUID)` — **none exist yet** |
| T3 | `internal/services/email_service.go` | Add `SendAnnouncementNotification(ctx, toEmail, title, description string) error` |
| T4 | `internal/services/announcement_notifier.go` | New file — concrete `announcementNotifier` impl; calls existing `PushNotificationService.SendToUser` with `dto.NotificationTestRequest{Title, Body}` |
| T5 | `cmd/server/main.go` | Wire `annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo)` into `NewAnnouncementService` |
| T6 | `internal/services/announcement_service_test.go` | Tests: `dispatchNotifications` + `resolveRecipientUserIDs` for all 3 audience modes |

---

## Schema / Migration

No migration needed. `announcement_views` and all join tables already exist.

---

## Intentional Divergences (do not revert)

| Topic | Python | Go | Rationale |
|-------|--------|----|-----------|
| List/Get access | Admin-only | Auth + visibility filter | Employees can read their own targeted announcements |
| `/send` body | Optional audience override at send time | No body — PATCH first, then publish | Separation of concerns |
| Delete | Hard delete | Soft delete | House style |
| Mobile list shape | Full `description` in list | Brief only (description in detail) | Bandwidth optimization |
| Status enum | `draft\|sent` | `draft\|scheduled\|published\|archived` | Go is a superset |
| Labels | Single `label_id` | Multi `label_ids[]` | Go enhancement |
| Author field | `created_by: string` | Nested `author: {id, full_name, avatar_url}` | Go richer read |
