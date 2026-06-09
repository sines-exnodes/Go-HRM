# Announcements Parity — Plan A (G3 + G1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a 409 guard blocking edits on published/archived announcements (G3) and register the mobile mark-as-read route alias (G1).

**Architecture:** Two independent one-line changes. G3 is a guard inserted at the top of `AnnouncementService.Update` before any field mutations. G1 is a single route registration in `main.go` that reuses the existing `MarkViewed` handler.

**Tech Stack:** Go 1.25, Gin, GORM, testify (integration tests against `exnodes_hrm_test` DB)

---

## ⚠️ REVISION NOTES

_None — this is the initial plan._

---

## File Map

| File | Action | What changes |
|------|--------|-------------|
| `internal/services/announcement_service.go` | Modify | Add status guard at top of `Update` method |
| `internal/services/announcement_service_test.go` | Modify | Add `TestAnnouncement_Update_PublishedRows_Returns409` + `TestAnnouncement_Update_ArchivedRows_Returns409` |
| `cmd/server/main.go` | Modify | Add `mobileAnnounce.POST(":id/read", announcementH.MarkViewed)` |

---

## Task 1: G3 — Edit guard for published/archived announcements

**Files:**
- Modify: `internal/services/announcement_service.go` (inside `Update`, after the row fetch block)
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1.1: Write the failing tests**

Add to `internal/services/announcement_service_test.go`, after the existing `TestAnnouncement_Update_*` tests:

```go
func TestAnnouncement_Update_PublishedRow_Returns409(t *testing.T) {
    svc, _ := newAnnouncementSvc(t)
    _, emp := makeEmpUser(t, "author-upd-pub@test.com")
    ann := makeAnnouncement(t, emp.ID, models.AnnouncementStatusPublished)

    _, err := svc.Update(context.Background(), ann.ID, emp.UserID, false, dto.AnnouncementUpdate{})
    require.Error(t, err)
    var appErr *apperrors.AppError
    require.ErrorAs(t, err, &appErr)
    assert.Equal(t, 409, appErr.Code)
}

func TestAnnouncement_Update_ArchivedRow_Returns409(t *testing.T) {
    svc, _ := newAnnouncementSvc(t)
    _, emp := makeEmpUser(t, "author-upd-arch@test.com")
    ann := makeAnnouncement(t, emp.ID, models.AnnouncementStatusArchived)

    _, err := svc.Update(context.Background(), ann.ID, emp.UserID, false, dto.AnnouncementUpdate{})
    require.Error(t, err)
    var appErr *apperrors.AppError
    require.ErrorAs(t, err, &appErr)
    assert.Equal(t, 409, appErr.Code)
}
```

Check that `makeAnnouncement` and `makeEmpUser` helpers exist in the test file. If `makeAnnouncement` does not exist, add it below the other `make*` helpers:

```go
func makeAnnouncement(t *testing.T, authorID uuid.UUID, status models.AnnouncementStatus) *models.Announcement {
    t.Helper()
    now := time.Now().UTC()
    ann := &models.Announcement{
        Title:          "Test announcement",
        Description:    "Test description",
        Status:         status,
        TargetAudience: models.AnnouncementAudienceAll,
        AuthorID:       authorID,
    }
    if status == models.AnnouncementStatusPublished || status == models.AnnouncementStatusArchived {
        ann.PublishedAt = &now
    }
    require.NoError(t, testDB.Create(ann).Error)
    return ann
}
```

Add the missing import if needed:
```go
"time"
```

- [ ] **Step 1.2: Run the failing tests**

```bash
cd e:\Work\Go-HRM
go test ./internal/services/... -run "TestAnnouncement_Update_PublishedRow_Returns409|TestAnnouncement_Update_ArchivedRow_Returns409" -v
```

Expected: both tests **FAIL** — `Update` currently allows edits on published/archived rows so `err` is `nil`.

- [ ] **Step 1.3: Add the status guard to `Update`**

In `internal/services/announcement_service.go`, inside `Update`, after the row-fetch error block (the `if errors.Is(err, gorm.ErrRecordNotFound)` block) and before the `resolveCurrentEmployee` call, insert:

```go
	// Block edits on terminal states — matches Python's 409-on-sent guard.
	if row.Status == models.AnnouncementStatusPublished ||
		row.Status == models.AnnouncementStatusArchived {
		return nil, apperrors.ErrConflict("Cannot edit a published or archived announcement")
	}
```

The full `Update` method opening should now read:

```go
func (s *AnnouncementService) Update(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool, in dto.AnnouncementUpdate) (*dto.AnnouncementRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Announcement")
		}
		return nil, err
	}
	// Block edits on terminal states — matches Python's 409-on-sent guard.
	if row.Status == models.AnnouncementStatusPublished ||
		row.Status == models.AnnouncementStatusArchived {
		return nil, apperrors.ErrConflict("Cannot edit a published or archived announcement")
	}
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	// ... rest unchanged
```

Verify `apperrors.ErrConflict` exists — grep for it:

```bash
grep -r "ErrConflict" internal/errors/
```

If it does not exist, add it to `internal/errors/errors.go` following the same pattern as `ErrNotFound`:

```go
func ErrConflict(msg string) *AppError {
    return &AppError{Code: 409, Message: msg}
}
```

- [ ] **Step 1.4: Run the tests — expect PASS**

```bash
go test ./internal/services/... -run "TestAnnouncement_Update_PublishedRow_Returns409|TestAnnouncement_Update_ArchivedRow_Returns409" -v
```

Expected: both tests **PASS**.

- [ ] **Step 1.5: Run full service test suite to check for regressions**

```bash
go test ./internal/services/... -v 2>&1 | tail -20
```

Expected: all existing tests pass. If any `TestAnnouncement_Update_*` test now fails because it uses a published row, update its fixture to use `AnnouncementStatusDraft`.

- [ ] **Step 1.6: Commit**

```bash
cd e:\Work\Go-HRM
git add internal/services/announcement_service.go internal/services/announcement_service_test.go internal/errors/errors.go
git commit -m "$(cat <<'EOF'
fix(announcements): 409 guard on edit of published/archived rows (G3)

Matches Python's ConflictException guard — PATCH on a published or
archived announcement now returns 409. Guard fires before ownership
check so neither owner nor admin can mutate terminal-state rows.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: G1 — Mobile mark-as-read route alias

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 2.1: Add the route**

In `cmd/server/main.go`, inside the `mobileAnnounce` group (around line 322–325), add one line:

```go
mobileAnnounce.GET("", announcementH.MobileBrief)
mobileAnnounce.GET("/list", announcementH.MobileList)
mobileAnnounce.GET(":id", announcementH.MobileGet)
mobileAnnounce.POST(":id/read", announcementH.MarkViewed)  // ← ADD THIS
```

- [ ] **Step 2.2: Build to confirm no compile errors**

```bash
cd e:\Work\Go-HRM
go build ./...
```

Expected: exits 0, no output.

- [ ] **Step 2.3: Smoke test the route**

Start the server (adjust port if 8080 is occupied):

```bash
PORT=8082 go run ./cmd/server &
sleep 2

# Login to get a token
TOKEN=$(curl -s -X POST http://localhost:8082/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.data.access_token')

# Get any announcement ID from the list
ANN_ID=$(curl -s http://localhost:8082/api/v1/announcements \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.items[0].id')

# Call the new mobile mark-as-read route
curl -s -X POST http://localhost:8082/api/v1/mobile/announcements/$ANN_ID/read \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected: `{"success":true,"message":"Marked as viewed"}` (or similar — same response as `POST /announcements/:id/view`).

Kill the dev server after smoke test:

```bash
kill %1
```

- [ ] **Step 2.4: Commit**

```bash
git add cmd/server/main.go
git commit -m "$(cat <<'EOF'
feat(announcements): mobile mark-as-read route alias (G1)

POST /api/v1/mobile/announcements/:id/read now maps to the existing
MarkViewed handler — same idempotent announcement_views write as the
web POST /announcements/:id/view. Mobile clients get their canonical
path; web clients keep /view. No new logic.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Final verification

- [ ] **Step 3.1: Run full test suite**

```bash
cd e:\Work\Go-HRM
make fmt && make vet && make test
```

Expected: `make fmt` and `make vet` exit 0. `make test` — all tests pass (0 failures, 0 skips outside of known skipped integration tests).

- [ ] **Step 3.2: Regenerate Swagger if annotations changed**

The guard and route alias don't change Swagger annotations, but verify:

```bash
grep -n "archived" internal/handlers/announcement_handler.go
```

If the `Update` godoc doesn't mention the 409 behaviour, add a `@Failure 409` annotation to it. Then run `make swag`.

- [ ] **Step 3.3: Plan A complete**

Plan A is done. Proceed to Plan B (`2026-06-09-announcements-parity-plan-b.md`) for G2 (push + email dispatch).
