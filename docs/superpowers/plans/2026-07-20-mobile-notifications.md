# Mobile In-App Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the mobile in-app notification list — a per-employee reverse-chronological feed with read state — plus the producers that generate rows on announcement publish and leave approve/reject.

**Architecture:** Fan-out on write. A `notifications` table holds one row per (recipient, event) with a snapshot of title/body, so notifications are independent of their source records. Producers hook into the existing `AnnouncementNotifier` seam and a new nil-safe `LeaveNotifier` seam. Reads are a single indexed query scoped to the JWT's `user_id`.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, golang-migrate, testify.

**Spec:** [`docs/superpowers/specs/2026-07-20-mobile-notifications-design.md`](../specs/2026-07-20-mobile-notifications-design.md)

---

## Before You Start

Read the spec first. Two things in it are load-bearing and easy to undo by accident:

1. **The unique index `uq_notifications_user_source` is what enforces DR Rule 5** (one notification per event). Producers rely on `ON CONFLICT DO NOTHING`. Don't "simplify" it away.
2. **Mark-read scopes its `UPDATE` by `id AND user_id`**, never by `id` alone. This is the data-isolation boundary (AC-01).

### Environment

The `migrate` CLI is not on PATH and `.env`'s `DATABASE_URL` uses the compose hostname `postgres`, which does not resolve from a host shell. Use the full path plus an explicit localhost URL:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
export TEST_DATABASE_URL="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
```

---

## Task 0: Capture the red baseline

The services suite is already failing on `main` for reasons unrelated to this work, and this plan modifies `LeaveService.Approve`/`Reject` — the exact code those failures sit under. Capture the failure set now so you can prove you didn't add to it.

**Files:** none (read-only)

- [ ] **Step 1: Run the leave suite and record what fails**

```bash
cd /Users/panda/work/Go-HRM
export TEST_DATABASE_URL="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
go test ./internal/services -run 'TestApprove|TestReject|TestUpdate_EmptyPatch' -v 2>&1 | grep -E '^(--- )?(FAIL|PASS|ok)' | sort | uniq -c
```

Expected: exactly these four FAIL, each with `forbidden: Only an admin can edit an approved leave request`:

- `TestApprove_TeamScope_CanApproveSubordinate`
- `TestApprove_TeamScope_RejectsNonSubordinate`
- `TestReject_TeamScope_CanRejectSubordinate`
- `TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus`

- [ ] **Step 2: Save the baseline to the scratchpad**

```bash
go test ./internal/services -run 'TestApprove|TestReject|TestUpdate_EmptyPatch' 2>&1 | tail -40 > /tmp/leave-baseline.txt
```

If **more** than those four fail, stop and report before writing any code — the baseline has drifted and you need to know why first.

---

## Task 1: Migration 000028

**Files:**
- Create: `migrations/000028_notifications.up.sql`
- Create: `migrations/000028_notifications.down.sql`

- [ ] **Step 1: Confirm 000028 is actually free**

```bash
ls migrations/ | grep -oE '^[0-9]{6}' | sort -u | tail -3
```

Expected: `000025`, `000026`, `000027`. If `000028` already exists, stop and re-number this task's files to the next free number, and tell the user.

- [ ] **Step 2: Write the up migration**

Create `migrations/000028_notifications.up.sql`:

```sql
-- migrations/000028_notifications.up.sql
--
-- Mobile in-app notification feed (DR-MOB-005-001-01). Fan-out on write:
-- one row per (recipient, event), carrying a snapshot of the title and body
-- taken at creation time.
--
-- The snapshot is deliberate (DR Rule 12) — editing an announcement after
-- publish must NOT rewrite what the employee was already told. It is also
-- why source_id carries no foreign key: the row must survive its source
-- being deleted (DR Rule 13), and source_id points into two different
-- tables depending on `type`.
--
-- user_id targets users(id), not employees(id). Notifications are an
-- auth-level surface consumed by the logged-in session, the same exception
-- already made for announcement_views and device_tokens.

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

-- List: WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC)
    WHERE is_deleted = FALSE;

-- Unread count for the dashboard header bell.
CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id)
    WHERE read_at IS NULL AND is_deleted = FALSE;

-- DR Rule 5 — one notification per event. Producers insert with
-- ON CONFLICT DO NOTHING, so re-publishing an announcement or retrying an
-- approve silently no-ops instead of duplicating. Removing this index
-- silently breaks Rule 5.
CREATE UNIQUE INDEX uq_notifications_user_source
    ON notifications (user_id, type, source_id)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_notifications_set_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 3: Write the down migration**

Create `migrations/000028_notifications.down.sql`:

```sql
-- migrations/000028_notifications.down.sql
DROP TRIGGER IF EXISTS trg_notifications_set_updated_at ON notifications;
DROP TABLE IF EXISTS notifications;
```

Dropping the table drops its indexes, so they need no explicit statements.

- [ ] **Step 4: Apply up, then down, then up again against the test DB**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
DB="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
migrate -path migrations -database "$DB" up
migrate -path migrations -database "$DB" version
migrate -path migrations -database "$DB" down 1
migrate -path migrations -database "$DB" up
migrate -path migrations -database "$DB" version
```

Expected: version `28` after the first `up`, `27` after the `down`, `28` again. No `dirty` in any output. The round-trip proves the down migration actually works — a down that errors leaves the DB dirty and blocks everyone.

- [ ] **Step 5: Commit**

```bash
git add migrations/000028_notifications.up.sql migrations/000028_notifications.down.sql
git commit -m "feat(notifications): migration 000028 notifications table"
```

---

## Task 2: Model

**Files:**
- Create: `internal/models/notification.go`

- [ ] **Step 1: Write the model**

Create `internal/models/notification.go`:

```go
package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType is the kind of event a notification represents. It is a
// stored, extensible value with no DB CHECK constraint — DR Rule 15 wants a
// new type to be a data change, not a schema migration. Validation lives
// here, mirroring AnnouncementStatus.
type NotificationType string

const (
	NotificationTypeAnnouncement NotificationType = "announcement"
	NotificationTypeLeaveRequest NotificationType = "leave_request"
)

// IsValid reports whether t is a known notification type.
func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeAnnouncement, NotificationTypeLeaveRequest:
		return true
	}
	return false
}

// Notification is one delivered event for one recipient.
//
// Title and Body are a SNAPSHOT taken at creation time (DR Rule 12) — the
// list is a faithful log of what the employee was told, even if the source
// record is later edited.
//
// SourceID has no foreign key on purpose: it points into announcements or
// leave_requests depending on Type, and the row must outlive its source
// (DR Rule 13).
//
// UserID targets users(id), not employees(id) — notifications are an
// auth-level surface, the same exception made for AnnouncementView.
type Notification struct {
	BaseModel
	UserID   uuid.UUID        `gorm:"type:uuid;not null;index"        json:"user_id"`
	Type     NotificationType `gorm:"type:text;not null"              json:"type"`
	Title    string           `gorm:"type:text;not null"              json:"title"`
	Body     string           `gorm:"type:text;not null;default:''"   json:"body"`
	SourceID uuid.UUID        `gorm:"type:uuid;not null"              json:"source_id"`

	// ReadAt is nil while unread. Read is terminal (DR Rule 8) — nothing
	// ever sets this back to nil. Storing the timestamp rather than a bool
	// costs the same and records when.
	ReadAt *time.Time `json:"read_at,omitempty"`
}

func (Notification) TableName() string { return "notifications" }
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/models/notification.go
git commit -m "feat(notifications): Notification model + NotificationType"
```

---

## Task 3: Repository

**Files:**
- Create: `internal/repositories/notification_repo.go`

- [ ] **Step 1: Write the repository**

Create `internal/repositories/notification_repo.go`:

```go
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/exnodes/hrm-api/internal/models"
)

// NotificationListQuery is the filter/pagination spec for List. UserID is
// mandatory — there is no unscoped list, by design (DR AC-01).
type NotificationListQuery struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
}

// NotificationRepository defines data access for the notifications table.
type NotificationRepository interface {
	// List returns the user's notifications newest-first, plus the total count.
	List(ctx context.Context, q NotificationListQuery) ([]models.Notification, int64, error)

	// CountUnread returns how many of the user's notifications are unread.
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)

	// CreateMany bulk-inserts notifications, skipping rows that collide with
	// uq_notifications_user_source. This is what makes DR Rule 5 (one
	// notification per event) hold on retry and re-publish.
	CreateMany(ctx context.Context, rows []models.Notification) error

	// MarkRead stamps read_at on a notification owned by userID and returns
	// the updated row. Returns gorm.ErrRecordNotFound when the row does not
	// exist OR belongs to someone else — the caller maps both to 404 so the
	// response cannot be used to probe for other users' notification IDs.
	// Marking an already-read row is a no-op that returns the row unchanged
	// (DR Rule 8: read is terminal).
	MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (*models.Notification, error)
}

type notificationRepo struct{ db *gorm.DB }

// NewNotificationRepository constructs a Postgres-backed NotificationRepository.
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *notificationRepo) List(ctx context.Context, q NotificationListQuery) ([]models.Notification, int64, error) {
	qb := r.base(ctx).Model(&models.Notification{}).Where("user_id = ?", q.UserID)

	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 50
	}

	var rows []models.Notification
	err := qb.
		Order("created_at DESC").
		Order("id DESC"). // stable tiebreak so page boundaries don't shuffle
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *notificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := r.base(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *notificationRepo) CreateMany(ctx context.Context, rows []models.Notification) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(&rows, 200).Error
}

func (r *notificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (*models.Notification, error) {
	var row models.Notification
	// Scoped by user_id: a foreign notification ID is indistinguishable from
	// a missing one.
	if err := r.base(ctx).First(&row, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	if row.ReadAt != nil {
		return &row, nil // already read — no-op (DR Rule 8)
	}

	res := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL AND is_deleted = ?", id, userID, false).
		Update("read_at", at)
	if res.Error != nil {
		return nil, res.Error
	}
	row.ReadAt = &at
	return &row, nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./... && go vet ./internal/repositories
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/repositories/notification_repo.go
git commit -m "feat(notifications): NotificationRepository"
```

---

## Task 4: DTOs

**Files:**
- Create: `internal/dto/notification.go`

- [ ] **Step 1: Write the DTOs**

Create `internal/dto/notification.go`:

```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// NotificationListQuery is the GET /mobile/notifications query string.
//
// No filter or sort params: DR Rule 16 makes the screen read-and-navigate
// only, and DR Rule 10 fixes the sort to newest-first.
type NotificationListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// NotificationRead is the canonical response shape for one notification.
//
// Type is returned as a bare string; the mobile app owns the icon/label
// registry from DR section 3. Shipping icon names over the wire would couple
// the API to a client icon set.
//
// Timestamps are RFC3339 UTC. AC-16's "20 July, 2026 - 10:00 AM" in employee
// local time is client-side formatting — the server has no reliable notion of
// the employee's timezone.
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

// NotificationUnreadCountRead backs the dashboard header bell (DR Rule 14).
type NotificationUnreadCountRead struct {
	UnreadCount int64 `json:"unread_count"`
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/dto/notification.go
git commit -m "feat(notifications): notification DTOs"
```

---

## Task 5: Add `notifications` to the test truncate list

This must land before any test in Task 7, or rows leak between tests and you will chase phantom failures.

**Files:**
- Modify: `internal/services/testhelper_test.go:137`

- [ ] **Step 1: Add the table to the TRUNCATE statement**

In `internal/services/testhelper_test.go`, the `truncateAll` function has one long `TRUNCATE TABLE ...` statement. Add `notifications` as the **first** table in the list (it has an FK to `users`, so it must precede `users`; `CASCADE` covers it either way but explicit order documents the dependency).

Change:

```go
	if err := testDB.Exec(`TRUNCATE TABLE password_reset_otps, password_reset_tokens, holidays, ...
```

to:

```go
	if err := testDB.Exec(`TRUNCATE TABLE notifications, password_reset_otps, password_reset_tokens, holidays, ...
```

Leave the rest of the statement exactly as it is.

- [ ] **Step 2: Add a comment above the statement documenting why**

Immediately before the `if err := testDB.Exec(...)` line, after the existing block of `// NOTE:` comments, add:

```go
	// Notifications (migration 000028) have an FK to users(id) and are
	// written by the announcement/leave notifiers, so they must be wiped
	// between tests or DR Rule 5 idempotency assertions see stale rows.
```

- [ ] **Step 3: Verify the suite still bootstraps**

```bash
export TEST_DATABASE_URL="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
go test ./internal/services -run TestHoliday_Create_HappyPath -v
```

Expected: PASS. This proves the truncate statement is still valid SQL.

- [ ] **Step 4: Commit**

```bash
git add internal/services/testhelper_test.go
git commit -m "test(notifications): truncate notifications between service tests"
```

---

## Task 6: NotificationService

**Files:**
- Create: `internal/services/notification_service.go`

- [ ] **Step 1: Write the service**

Create `internal/services/notification_service.go`:

```go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

const (
	notificationDefaultPageSize = 50
	notificationMaxPageSize     = 100
)

// NotificationService owns the in-app notification feed. Every read and
// write is scoped to a single user's ID — there is no unscoped path, which
// is how DR AC-01 (an employee never sees another's notifications) is
// enforced server-side.
type NotificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// List returns the user's notifications, newest first.
func (s *NotificationService) List(
	ctx context.Context,
	userID uuid.UUID,
	q dto.NotificationListQuery,
) (dto.PaginatedData[dto.NotificationRead], *apperrors.AppError) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = notificationDefaultPageSize
	}
	if pageSize > notificationMaxPageSize {
		pageSize = notificationMaxPageSize
	}

	rows, total, err := s.repo.List(ctx, repositories.NotificationListQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return dto.PaginatedData[dto.NotificationRead]{}, apperrors.ErrInternal(err.Error())
	}

	items := make([]dto.NotificationRead, 0, len(rows))
	for i := range rows {
		items = append(items, notificationToRead(&rows[i]))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return dto.PaginatedData[dto.NotificationRead]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UnreadCount backs the dashboard header bell.
func (s *NotificationService) UnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (dto.NotificationUnreadCountRead, *apperrors.AppError) {
	n, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return dto.NotificationUnreadCountRead{}, apperrors.ErrInternal(err.Error())
	}
	return dto.NotificationUnreadCountRead{UnreadCount: n}, nil
}

// MarkRead stamps the notification read for this user and returns it.
//
// Marking an already-read notification is a 200 no-op, not a 409: DR Rule 8
// makes read terminal, so a repeat is a successful arrival at the intended
// state, and the mobile client may legitimately retry after a dropped
// response.
func (s *NotificationService) MarkRead(
	ctx context.Context,
	id, userID uuid.UUID,
) (*dto.NotificationRead, *apperrors.AppError) {
	row, err := s.repo.MarkRead(ctx, id, userID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Notification")
		}
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := notificationToRead(row)
	return &out, nil
}

// CreateMany is the producer-facing entry point. Collisions on
// uq_notifications_user_source are skipped, so callers get DR Rule 5
// (one notification per event) for free on retry.
func (s *NotificationService) CreateMany(ctx context.Context, rows []models.Notification) error {
	return s.repo.CreateMany(ctx, rows)
}

func notificationToRead(m *models.Notification) dto.NotificationRead {
	return dto.NotificationRead{
		ID:        m.ID,
		Type:      string(m.Type),
		Title:     m.Title,
		Body:      m.Body,
		SourceID:  m.SourceID,
		IsRead:    m.ReadAt != nil,
		ReadAt:    m.ReadAt,
		CreatedAt: m.CreatedAt,
	}
}
```

- [ ] **Step 2: Check the error-helper names actually exist**

```bash
grep -n "func ErrInternal\|func ErrNotFound" internal/errors/errors.go
```

Expected: both present. If `ErrInternal` has a different name or signature, adjust the calls above to match — do not invent a helper.

- [ ] **Step 3: Verify it compiles**

```bash
go build ./... && go vet ./internal/services
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add internal/services/notification_service.go
git commit -m "feat(notifications): NotificationService with user-scoped reads"
```

---

## Task 7: Service tests

Written before the producers so the storage layer is proven independently. These are integration tests against the real test DB, matching `holiday_service_test.go` in shape.

**Files:**
- Create: `internal/services/notification_service_test.go`

- [ ] **Step 1: Write the tests**

Create `internal/services/notification_service_test.go`:

```go
package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newNotificationSvc(t *testing.T) (*services.NotificationService, repositories.NotificationRepository) {
	t.Helper()
	repo := repositories.NewNotificationRepository(testDB)
	return services.NewNotificationService(repo), repo
}

// seedNotification inserts one notification directly through the repo.
func seedNotification(t *testing.T, repo repositories.NotificationRepository, userID uuid.UUID, typ models.NotificationType, title string, sourceID uuid.UUID) {
	t.Helper()
	err := repo.CreateMany(context.Background(), []models.Notification{{
		UserID:   userID,
		Type:     typ,
		Title:    title,
		Body:     title + " body",
		SourceID: sourceID,
	}})
	require.NoError(t, err)
}

func TestNotification_List_NewestFirst(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	u := makeUser(t, "list@example.com", "pw-Aa123456")

	// Insert sequentially so created_at ordering is deterministic.
	for i := 0; i < 3; i++ {
		seedNotification(t, repo, u.ID, models.NotificationTypeAnnouncement,
			fmt.Sprintf("ann-%d", i), uuid.New())
		time.Sleep(5 * time.Millisecond)
	}

	out, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 3)
	assert.Equal(t, int64(3), out.Total)
	assert.Equal(t, 50, out.PageSize, "default page size should be 50")

	// DR Rule 10 — newest first, always.
	assert.Equal(t, "ann-2", out.Items[0].Title)
	assert.Equal(t, "ann-1", out.Items[1].Title)
	assert.Equal(t, "ann-0", out.Items[2].Title)
}

func TestNotification_List_PaginatesAcrossBoundary(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	u := makeUser(t, "page@example.com", "pw-Aa123456")

	for i := 0; i < 5; i++ {
		seedNotification(t, repo, u.ID, models.NotificationTypeAnnouncement,
			fmt.Sprintf("n-%d", i), uuid.New())
		time.Sleep(5 * time.Millisecond)
	}

	p1, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{Page: 1, PageSize: 2})
	require.Nil(t, aerr)
	p2, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{Page: 2, PageSize: 2})
	require.Nil(t, aerr)

	require.Len(t, p1.Items, 2)
	require.Len(t, p2.Items, 2)
	assert.Equal(t, int64(5), p1.Total)
	assert.Equal(t, 3, p1.TotalPages)
	assert.Equal(t, []string{"n-4", "n-3"}, []string{p1.Items[0].Title, p1.Items[1].Title})
	assert.Equal(t, []string{"n-2", "n-1"}, []string{p2.Items[0].Title, p2.Items[1].Title})
}

func TestNotification_List_CapsPageSizeAt100(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, _ := newNotificationSvc(t)
	u := makeUser(t, "cap@example.com", "pw-Aa123456")

	out, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{PageSize: 5000})
	require.Nil(t, aerr)
	// An unbounded page size is the footgun the cap exists to prevent —
	// DR Rule 11 keeps notifications forever.
	assert.Equal(t, 100, out.PageSize)
}

// AC-01 / "Data isolation" scenario. The second assertion is the one that
// matters: a rejected mark-read must not have side effects on the victim.
func TestNotification_DataIsolation_ForeignRowInvisibleAndUntouched(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	alice := makeUser(t, "alice@example.com", "pw-Aa123456")
	bob := makeUser(t, "bob@example.com", "pw-Aa123456")

	seedNotification(t, repo, bob.ID, models.NotificationTypeAnnouncement, "bob-only", uuid.New())

	// Alice cannot see Bob's notification.
	out, aerr := svc.List(context.Background(), alice.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	assert.Empty(t, out.Items)
	assert.Equal(t, int64(0), out.Total)

	// Fetch Bob's row to get its real ID.
	bobList, aerr := svc.List(context.Background(), bob.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, bobList.Items, 1)
	bobNotifID := bobList.Items[0].ID

	// Alice marking Bob's ID gets a 404 — indistinguishable from missing.
	_, aerr = svc.MarkRead(context.Background(), bobNotifID, alice.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.HTTP)

	// And Bob's row is still unread. A 404 that still mutated would be worse
	// than a 403 that didn't.
	bobList, listErr := svc.List(context.Background(), bob.ID, dto.NotificationListQuery{})
	require.Nil(t, listErr)
	assert.False(t, bobList.Items[0].IsRead, "rejected mark-read must not touch the owner's row")
}

func TestNotification_MarkRead_SetsReadState(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	u := makeUser(t, "read@example.com", "pw-Aa123456")
	seedNotification(t, repo, u.ID, models.NotificationTypeAnnouncement, "unread", uuid.New())

	list, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, list.Items, 1)
	assert.False(t, list.Items[0].IsRead)

	out, aerr := svc.MarkRead(context.Background(), list.Items[0].ID, u.ID)
	require.Nil(t, aerr)
	assert.True(t, out.IsRead)
	require.NotNil(t, out.ReadAt)

	count, aerr := svc.UnreadCount(context.Background(), u.ID)
	require.Nil(t, aerr)
	assert.Equal(t, int64(0), count.UnreadCount)
}

// DR Rule 8 — read is terminal. A repeat is a successful no-op, and it must
// not move the original timestamp.
func TestNotification_MarkRead_IsIdempotentAndDoesNotMoveTimestamp(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	u := makeUser(t, "twice@example.com", "pw-Aa123456")
	seedNotification(t, repo, u.ID, models.NotificationTypeAnnouncement, "once", uuid.New())

	list, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	id := list.Items[0].ID

	first, aerr := svc.MarkRead(context.Background(), id, u.ID)
	require.Nil(t, aerr)
	require.NotNil(t, first.ReadAt)

	time.Sleep(10 * time.Millisecond)

	second, aerr := svc.MarkRead(context.Background(), id, u.ID)
	require.Nil(t, aerr, "re-marking must be a 200 no-op, not a 409")
	require.NotNil(t, second.ReadAt)
	assert.WithinDuration(t, *first.ReadAt, *second.ReadAt, time.Millisecond,
		"read_at must not move on repeat")
}

func TestNotification_MarkRead_UnknownIDIs404(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, _ := newNotificationSvc(t)
	u := makeUser(t, "missing@example.com", "pw-Aa123456")

	_, aerr := svc.MarkRead(context.Background(), uuid.New(), u.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.HTTP)
}

// DR Rule 5 — one notification per event. This is the test that fails if
// someone later drops uq_notifications_user_source.
func TestNotification_CreateMany_DedupesOnRepeat(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	u := makeUser(t, "dedupe@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	row := models.Notification{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "same event",
		Body:     "body",
		SourceID: sourceID,
	}

	require.NoError(t, svc.CreateMany(context.Background(), []models.Notification{row}))
	// Same (user, type, source) again — e.g. a re-publish or a retried dispatch.
	require.NoError(t, svc.CreateMany(context.Background(), []models.Notification{{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "same event",
		Body:     "body",
		SourceID: sourceID,
	}}))

	out, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	assert.Len(t, out.Items, 1, "re-dispatching the same event must not duplicate")

	_ = repo
}

func TestNotification_UnreadCount_CountsOnlyOwnUnread(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, repo := newNotificationSvc(t)
	alice := makeUser(t, "ca@example.com", "pw-Aa123456")
	bob := makeUser(t, "cb@example.com", "pw-Aa123456")

	seedNotification(t, repo, alice.ID, models.NotificationTypeAnnouncement, "a1", uuid.New())
	seedNotification(t, repo, alice.ID, models.NotificationTypeLeaveRequest, "a2", uuid.New())
	seedNotification(t, repo, bob.ID, models.NotificationTypeAnnouncement, "b1", uuid.New())

	count, aerr := svc.UnreadCount(context.Background(), alice.ID)
	require.Nil(t, aerr)
	assert.Equal(t, int64(2), count.UnreadCount)
}
```

- [ ] **Step 2: Run the tests**

```bash
export TEST_DATABASE_URL="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
go test ./internal/services -run 'TestNotification_' -v
```

Expected: all PASS. If `aerr.HTTP` does not compile, check the actual field name on `apperrors.AppError` with `grep -n "type AppError struct" -A 8 internal/errors/errors.go` and use the real one.

- [ ] **Step 3: Commit**

```bash
git add internal/services/notification_service_test.go
git commit -m "test(notifications): service tests for list, isolation, dedupe, read state"
```

---

## Task 8: Announcement producer

**Files:**
- Modify: `internal/services/announcement_notifier.go`
- Modify: `cmd/server/main.go:121`

The existing `AnnouncementNotifier` seam already delivers recipient user IDs, the announcement ID, title, and description. `AnnouncementService` needs no change at all.

- [ ] **Step 1: Add the notifications collaborator**

In `internal/services/announcement_notifier.go`, change the struct and constructor:

```go
type announcementNotifier struct {
	push   *PushNotificationService
	email  *EmailService
	users  repositories.UserRepository
	notifs *NotificationService // optional — nil disables the in-app feed
}

func NewAnnouncementNotifier(
	push *PushNotificationService,
	email *EmailService,
	users repositories.UserRepository,
	notifs *NotificationService,
) AnnouncementNotifier {
	return &announcementNotifier{push: push, email: email, users: users, notifs: notifs}
}
```

- [ ] **Step 2: Write notification rows at the top of NotifyAnnouncement**

In the same file, insert this block as the **first** thing inside `NotifyAnnouncement`, before the existing `for _, uid := range userIDs` loop:

```go
	// In-app feed rows first: this is the durable surface. Push and email are
	// best-effort side channels, so a failure there must not cost the
	// employee the notification itself.
	//
	// AC-10 (drafts generate nothing) is already guaranteed upstream —
	// broadcastPublished returns early when PublishedAt is nil, and
	// dispatchNotifications is only reachable from there.
	if n.notifs != nil {
		rows := make([]models.Notification, 0, len(userIDs))
		for _, uid := range userIDs {
			rows = append(rows, models.Notification{
				UserID:   uid,
				Type:     models.NotificationTypeAnnouncement,
				Title:    title,
				Body:     plainTextPreview(description, 512),
				SourceID: id,
			})
		}
		if err := n.notifs.CreateMany(ctx, rows); err != nil {
			log.Printf("announcements: create in-app notifications for %s: %v", id, err)
		}
	}
```

Add `"github.com/exnodes/hrm-api/internal/models"` to the file's import block.

The body is stored as plain text via the existing `plainTextPreview` helper — the DR clamps to 3 lines visually, but storing raw HTML per recipient to render 3 lines would be wasteful. 512 runes is comfortably more than 3 lines at any phone width.

- [ ] **Step 3: Update the call site in main.go**

Construct the notification service **above line 100** (where `leaveSvc` is built), not next to the notifier — Task 9 needs it there too, and building it once at the top avoids a second move:

```go
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationSvc := services.NewNotificationService(notificationRepo)
```

Then at `cmd/server/main.go:121`, change:

```go
	annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo)
```

to:

```go
	annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo, notificationSvc)
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 5: Add the announcement trigger test**

Append to `internal/services/notification_service_test.go`:

```go
// AC-10 — a draft announcement generates nothing. The announcement service's
// publish gate is what enforces this; this test pins it so a future refactor
// of broadcastPublished cannot silently start notifying on draft save.
func TestNotification_Announcement_DraftGeneratesNothing(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)

	author := makeUser(t, "author-draft@example.com", "pw-Aa123456")
	makeEmployee(t, author, "Draft Author")
	recipient := makeUser(t, "recipient-draft@example.com", "pw-Aa123456")
	makeEmployee(t, recipient, "Draft Recipient")

	annSvc := svcWithNotifier(t, services.NewAnnouncementNotifier(nil, nil, testUserRepo, notifSvc))

	_, err := annSvc.Create(context.Background(), author.ID, dto.AnnouncementCreate{
		Title:       "Draft only",
		Description: "<p>never sent</p>",
	})
	require.NoError(t, err)

	out, aerr := notifSvc.List(context.Background(), recipient.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	assert.Empty(t, out.Items, "a draft announcement must not notify anyone")
}
```

Note: `svcWithNotifier` already exists in `announcement_service_test.go` in the same `services_test` package — reuse it, don't redefine it.

- [ ] **Step 6: Run it**

```bash
go test ./internal/services -run 'TestNotification_Announcement' -v
```

Expected: PASS. If `dto.AnnouncementCreate` requires fields beyond `Title`/`Description`, check the struct with `grep -n "type AnnouncementCreate struct" -A 15 internal/dto/announcement.go` and fill in the required ones.

- [ ] **Step 7: Commit**

```bash
git add internal/services/announcement_notifier.go internal/services/notification_service_test.go cmd/server/main.go
git commit -m "feat(notifications): create in-app rows on announcement publish"
```

---

## Task 9: Leave producer

**Files:**
- Create: `internal/services/notification_notifier.go`
- Modify: `internal/services/leave_service.go` (struct, constructor, Approve, Reject)
- Modify: `internal/services/leave_service_test.go:23-32`
- Modify: `cmd/server/main.go:100`

- [ ] **Step 1: Add the LeaveNotifier seam to LeaveService**

In `internal/services/leave_service.go`, add the interface immediately above `type LeaveService struct`:

```go
// LeaveNotifier receives leave decision events. Optional — a nil notifier
// disables in-app notifications, which keeps every existing test that
// constructs a LeaveService without one working unchanged.
type LeaveNotifier interface {
	NotifyLeaveDecision(ctx context.Context, employeeID, leaveID uuid.UUID, approved bool, from, to time.Time)
}
```

Add the field to the struct:

```go
type LeaveService struct {
	leaves   repositories.LeaveRequestRepository
	emps     repositories.EmployeeRepository
	depts    repositories.DepartmentRepository
	pos      repositories.PositionRepository
	quota    repositories.LeaveQuotaRepository
	uploads  Uploader // optional; nil means attachment upload is unavailable
	holidays repositories.HolidayRepository
	notifier LeaveNotifier // optional; nil disables in-app notifications
}
```

And extend the constructor:

```go
func NewLeaveService(
	leaves repositories.LeaveRequestRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
	holidays repositories.HolidayRepository,
	notifier LeaveNotifier,
) *LeaveService {
	return &LeaveService{
		leaves:   leaves,
		emps:     emps,
		depts:    depts,
		pos:      pos,
		quota:    quota,
		uploads:  uploads,
		holidays: holidays,
		notifier: notifier,
	}
}
```

- [ ] **Step 2: Call the notifier from Approve and Reject**

`Approve` currently ends at line ~611 with:

```go
	return s.transitionStatus(ctx, id, models.LeaveStatusApproved, []models.LeaveStatus{models.LeaveStatusPending})
```

Replace that single line with:

```go
	read, err := s.transitionStatus(ctx, id, models.LeaveStatusApproved, []models.LeaveStatus{models.LeaveStatusPending})
	if err != nil {
		return nil, err
	}
	// After the transition commits, never before — a rejected or conflicting
	// transition must not produce a notification claiming it succeeded.
	if s.notifier != nil {
		s.notifier.NotifyLeaveDecision(ctx, row.EmployeeID, row.ID, true, row.FromDate, row.ToDate)
	}
	return read, nil
```

Apply the same change in `Reject`, with `models.LeaveStatusRejected` and `false`:

```go
	read, err := s.transitionStatus(ctx, id, models.LeaveStatusRejected, []models.LeaveStatus{models.LeaveStatusPending})
	if err != nil {
		return nil, err
	}
	if s.notifier != nil {
		s.notifier.NotifyLeaveDecision(ctx, row.EmployeeID, row.ID, false, row.FromDate, row.ToDate)
	}
	return read, nil
```

Both methods already have `row` in scope from the `s.leaves.FindByID` call at the top, so `EmployeeID`, `FromDate`, and `ToDate` are available without a second query.

`Cancel` gets no notifier call — DR Rule 4 excludes cancellation, and that exclusion is flagged as an open question for the PO.

- [ ] **Step 3: Write the leave notifier**

Create `internal/services/notification_notifier.go`:

```go
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// leaveDecisionDateFormat matches the date rendering in the notification
// body copy specified by DR section 3.
const leaveDecisionDateFormat = "2006-01-02"

// leaveNotifier writes an in-app notification when a leave request is
// approved or rejected (DR Rule 4).
//
// The leave aggregate keys on employees(id) but notifications key on
// users(id), so this type owns the employee → user resolution. That
// translation is exactly why the notifier is a separate collaborator rather
// than a method on LeaveService.
type leaveNotifier struct {
	notifs *NotificationService
	emps   repositories.EmployeeRepository
}

// NewLeaveNotifier constructs the concrete LeaveNotifier.
func NewLeaveNotifier(notifs *NotificationService, emps repositories.EmployeeRepository) LeaveNotifier {
	return &leaveNotifier{notifs: notifs, emps: emps}
}

func (n *leaveNotifier) NotifyLeaveDecision(
	ctx context.Context,
	employeeID, leaveID uuid.UUID,
	approved bool,
	from, to time.Time,
) {
	if n.notifs == nil || n.emps == nil {
		return
	}

	emp, err := n.emps.Get(ctx, employeeID)
	if err != nil {
		log.Printf("leave: resolve employee %s for notification: %v", employeeID, err)
		return
	}

	title, body := leaveDecisionCopy(approved, from, to)

	err = n.notifs.CreateMany(ctx, []models.Notification{{
		UserID:   emp.UserID,
		Type:     models.NotificationTypeLeaveRequest,
		Title:    title,
		Body:     body,
		SourceID: leaveID,
	}})
	if err != nil {
		log.Printf("leave: create notification for leave %s: %v", leaveID, err)
	}
}

// leaveDecisionCopy renders the DR section 3 / AC-09 copy. Extracted so the
// exact wording is testable without a database.
func leaveDecisionCopy(approved bool, from, to time.Time) (title, body string) {
	verb := "rejected"
	title = "Leave Request Rejected"
	if approved {
		verb = "approved"
		title = "Leave Request Approved"
	}
	body = fmt.Sprintf(
		"Your leave request from %s to %s has been %s.",
		from.Format(leaveDecisionDateFormat),
		to.Format(leaveDecisionDateFormat),
		verb,
	)
	return title, body
}
```

- [ ] **Step 4: Check the employee repo getter name**

```bash
grep -n "Get(ctx context.Context, id uuid.UUID)\|FindByID(ctx context.Context, id uuid.UUID)" internal/repositories/employee_repo.go
```

Use whichever method actually exists on `EmployeeRepository` and adjust `n.emps.Get(...)` above to match. Do not add a new method to the interface.

- [ ] **Step 5: Fix the two broken call sites**

`internal/services/leave_service_test.go:31` — add a trailing `nil` for the notifier:

```go
	return services.NewLeaveService(lr, emps, depts, pos, quotaRepo, up, holidayRepo, nil), lr, quotaRepo
```

`cmd/server/main.go:100` — `notificationSvc` is already in scope from Task 8 Step 3, which placed it above this line. Add the notifier and extend the call:

```go
	leaveNotifier := services.NewLeaveNotifier(notificationSvc, employeeRepo)
	leaveSvc := services.NewLeaveService(leaveRepo, employeeRepo, departmentRepo, positionRepo, quotaRepo, uploadSvc, holidayRepo, leaveNotifier)
```

- [ ] **Step 6: Verify everything compiles**

```bash
go build ./... && go vet ./...
```

Expected: no output. A compile error here almost certainly means a call site of `NewLeaveService` was missed — `grep -rn "NewLeaveService(" cmd internal` lists all of them.

- [ ] **Step 7: Add the leave trigger tests**

Append to `internal/services/notification_service_test.go`:

```go
// AC-09 — approve produces the approved title and a body naming the range.
func TestNotification_Leave_ApproveProducesNotification(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)
	empRepo := repositories.NewEmployeeRepository(testDB)

	leaveSvc := services.NewLeaveService(
		repositories.NewLeaveRequestRepository(testDB),
		empRepo,
		repositories.NewDepartmentRepository(testDB),
		repositories.NewPositionRepository(testDB),
		repositories.NewLeaveQuotaRepository(testDB),
		nil,
		repositories.NewHolidayRepository(testDB),
		services.NewLeaveNotifier(notifSvc, empRepo),
	)

	u := makeUser(t, "leave-approve@example.com", "pw-Aa123456")
	emp := makeEmployee(t, u, "Leave Taker")
	makeLeaveQuota(t, emp.ID, 12, 6)

	from := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)

	created, err := leaveSvc.Create(context.Background(), u.ID, dto.LeaveRequestCreate{
		FromDate:  from,
		ToDate:    to,
		LeaveType: models.LeaveTypeAnnual,
		Reason:    "family trip",
	})
	require.NoError(t, err)

	_, err = leaveSvc.Approve(context.Background(), created.ID, u.ID, services.ApproveScopeAll)
	require.NoError(t, err)

	out, aerr := notifSvc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "Leave Request Approved", out.Items[0].Title)
	assert.Equal(t, string(models.NotificationTypeLeaveRequest), out.Items[0].Type)
	assert.Equal(t, created.ID, out.Items[0].SourceID, "source_id must point at the leave request")
	assert.Contains(t, out.Items[0].Body, "2026-08-03")
	assert.Contains(t, out.Items[0].Body, "2026-08-05")
	assert.Contains(t, out.Items[0].Body, "approved")
}

// DR Rule 5 at the leave layer — a repeat approve must not double-notify.
func TestNotification_Leave_ReApproveDoesNotDuplicate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)
	empRepo := repositories.NewEmployeeRepository(testDB)

	leaveSvc := services.NewLeaveService(
		repositories.NewLeaveRequestRepository(testDB),
		empRepo,
		repositories.NewDepartmentRepository(testDB),
		repositories.NewPositionRepository(testDB),
		repositories.NewLeaveQuotaRepository(testDB),
		nil,
		repositories.NewHolidayRepository(testDB),
		services.NewLeaveNotifier(notifSvc, empRepo),
	)

	u := makeUser(t, "leave-twice@example.com", "pw-Aa123456")
	emp := makeEmployee(t, u, "Twice Taker")
	makeLeaveQuota(t, emp.ID, 12, 6)

	created, err := leaveSvc.Create(context.Background(), u.ID, dto.LeaveRequestCreate{
		FromDate:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		ToDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		LeaveType: models.LeaveTypeAnnual,
		Reason:    "appointment",
	})
	require.NoError(t, err)

	_, err = leaveSvc.Approve(context.Background(), created.ID, u.ID, services.ApproveScopeAll)
	require.NoError(t, err)

	// Second approve is rejected by the status guard, so no second
	// notification should exist either way.
	_, _ = leaveSvc.Approve(context.Background(), created.ID, u.ID, services.ApproveScopeAll)

	out, aerr := notifSvc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	assert.Len(t, out.Items, 1)
}

// AC-09 rejected variant.
func TestNotification_Leave_RejectProducesRejectedCopy(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)
	empRepo := repositories.NewEmployeeRepository(testDB)

	leaveSvc := services.NewLeaveService(
		repositories.NewLeaveRequestRepository(testDB),
		empRepo,
		repositories.NewDepartmentRepository(testDB),
		repositories.NewPositionRepository(testDB),
		repositories.NewLeaveQuotaRepository(testDB),
		nil,
		repositories.NewHolidayRepository(testDB),
		services.NewLeaveNotifier(notifSvc, empRepo),
	)

	u := makeUser(t, "leave-reject@example.com", "pw-Aa123456")
	emp := makeEmployee(t, u, "Reject Taker")
	makeLeaveQuota(t, emp.ID, 12, 6)

	created, err := leaveSvc.Create(context.Background(), u.ID, dto.LeaveRequestCreate{
		FromDate:  time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC),
		ToDate:    time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC),
		LeaveType: models.LeaveTypeAnnual,
		Reason:    "personal",
	})
	require.NoError(t, err)

	_, err = leaveSvc.Reject(context.Background(), created.ID, u.ID, services.ApproveScopeAll)
	require.NoError(t, err)

	out, aerr := notifSvc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "Leave Request Rejected", out.Items[0].Title)
	assert.Contains(t, out.Items[0].Body, "rejected")
}
```

- [ ] **Step 8: Run the notification tests plus the leave baseline**

```bash
go test ./internal/services -run 'TestNotification_' -v
go test ./internal/services -run 'TestApprove|TestReject|TestUpdate_EmptyPatch' 2>&1 | tail -40
```

Expected: all `TestNotification_*` PASS. The leave run must show **exactly the same four failures** as `/tmp/leave-baseline.txt` from Task 0 — no more. Diff them if unsure. If a fifth test now fails, you broke it; fix before continuing.

If `dto.LeaveRequestCreate` field names or `models.LeaveTypeAnnual` differ, check with `grep -n "type LeaveRequestCreate struct" -A 15 internal/dto/leave.go` and `grep -n "LeaveType.*=" internal/models/leave_request.go`.

- [ ] **Step 9: Commit**

```bash
git add internal/services/notification_notifier.go internal/services/leave_service.go internal/services/leave_service_test.go internal/services/notification_service_test.go cmd/server/main.go
git commit -m "feat(notifications): notify on leave approve/reject"
```

---

## Task 10: Handler endpoints

**Files:**
- Modify: `internal/handlers/notification_handler.go`
- Modify: `cmd/server/main.go:159`

- [ ] **Step 1: Extend the handler struct**

In `internal/handlers/notification_handler.go`, change the struct and constructor:

```go
// NotificationHandler owns the mobile in-app notification feed plus the
// legacy /notifications/test admin push-debug endpoint.
type NotificationHandler struct {
	svc      *services.PushNotificationService
	notifSvc *services.NotificationService
}

func NewNotificationHandler(
	svc *services.PushNotificationService,
	notifSvc *services.NotificationService,
) *NotificationHandler {
	return &NotificationHandler{svc: svc, notifSvc: notifSvc}
}
```

Leave the existing `SendTest` method untouched.

- [ ] **Step 2: Add the three endpoints**

Append to `internal/handlers/notification_handler.go`:

```go
// List godoc
// @Summary      List the caller's in-app notifications
// @Description  Reverse-chronological feed of the authenticated employee's notifications. Always scoped to the caller — there is no way to read another employee's feed.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "page number (default 1)"
// @Param        page_size  query  int  false  "page size (default 50, max 100)"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.NotificationRead]]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Router       /mobile/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var q dto.NotificationListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.notifSvc.List(c.Request.Context(), u.ID, q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.NotificationRead]]{Success: true, Data: out})
}

// UnreadCount godoc
// @Summary      Count the caller's unread notifications
// @Description  Backs the dashboard header notification bell. Cheap to poll after marking a notification read.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[dto.NotificationUnreadCountRead]
// @Failure      401  {object}  dto.Response[any]
// @Router       /mobile/notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	out, aerr := h.notifSvc.UnreadCount(c.Request.Context(), u.ID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.NotificationUnreadCountRead]{Success: true, Data: out})
}

// MarkRead godoc
// @Summary      Mark a notification read
// @Description  Marks the notification read for the caller and returns it. Read is terminal — re-marking an already-read notification is a 200 no-op, not a conflict. A notification belonging to another employee returns 404, indistinguishable from a missing one.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "notification id"
// @Success      200  {object}  dto.Response[dto.NotificationRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /mobile/notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, aerr := h.notifSvc.MarkRead(c.Request.Context(), id, u.ID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.NotificationRead]{Success: true, Data: out})
}
```

Note the `parseIDParam` result is named `err`, not `aerr`. This is deliberate and load-bearing: reusing an `aerr` variable across `parseIDParam` (which returns `error`) and a service call (which returns `*apperrors.AppError`) boxes a typed-nil into a non-nil interface and panics the error middleware. That exact bug was already found and fixed once in the holiday handler.

- [ ] **Step 3: Update the constructor call site**

`cmd/server/main.go:159`:

```go
	notifH := handlers.NewNotificationHandler(pushSvc, notificationSvc)
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./... && go vet ./...
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/notification_handler.go cmd/server/main.go
git commit -m "feat(notifications): list, unread-count, mark-read endpoints"
```

---

## Task 11: Routes

**Files:**
- Modify: `cmd/server/main.go` (after the `mobileAnnounce` block, ~line 388)

- [ ] **Step 1: Register the routes**

In `cmd/server/main.go`, immediately after the `mobileAnnounce.POST(":id/read", ...)` line, add:

```go
		// ---- /mobile/notifications (EP-005 — in-app notification feed) ----
		// JWT-only, no permission gate: every authenticated employee is
		// entitled to their own notifications by definition, and the service
		// scopes every query to the caller's user_id. A notifications:read
		// permission would have to be seeded to all five roles immediately,
		// making it a permission that can never be false.
		//
		// Gin route precedence: the literal /unread-count is registered
		// before any wildcard segment on this group.
		mobileNotif := authed.Group("/mobile/notifications")
		mobileNotif.GET("/unread-count", notifH.UnreadCount)
		mobileNotif.GET("", notifH.List)
		mobileNotif.POST(":id/read", notifH.MarkRead)
```

- [ ] **Step 2: Verify the server boots and the routes are registered**

```bash
go build -o /tmp/hrm-server ./cmd/server && go vet ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat(notifications): wire /mobile/notifications routes"
```

---

## Task 12: Dashboard unread count

**Files:**
- Modify: `internal/dto/dashboard.go`
- Modify: `internal/services/dashboard_service.go`
- Modify: `internal/services/dashboard_service_test.go:21`
- Modify: `cmd/server/main.go:136`

- [ ] **Step 1: Add the field to the DTO**

In `internal/dto/dashboard.go`, add to `DashboardRead`:

```go
type DashboardRead struct {
	Greeting     DashboardGreetingRead `json:"greeting"`
	Widgets      []DashboardWidgetRead `json:"widgets"`
	Empty        bool                  `json:"empty"`
	EmptyMessage string                `json:"empty_message,omitempty"`

	// UnreadNotificationCount backs the mobile dashboard header bell
	// (DR-MOB-005-001-01 Rule 14). Unlike the widgets, this is never omitted:
	// notifications carry no permission gate, so every caller has a count.
	UnreadNotificationCount int64 `json:"unread_notification_count"`
}
```

The spec calls this field `unread_count`; `unread_notification_count` is used here because at the dashboard root, `unread_count` does not say unread *what*. Update the spec's §5 wording to match when this task lands.

- [ ] **Step 2: Add the repository to DashboardService**

In `internal/services/dashboard_service.go`, add the field and constructor parameter:

```go
type DashboardService struct {
	cfg           *config.Config
	emps          repositories.EmployeeRepository
	leaves        repositories.LeaveRequestRepository
	quota         repositories.LeaveQuotaRepository
	attendance    repositories.AttendanceRepository
	announcements repositories.AnnouncementRepository
	holidays      repositories.HolidayRepository
	notifications repositories.NotificationRepository
}

func NewDashboardService(
	cfg *config.Config,
	emps repositories.EmployeeRepository,
	leaves repositories.LeaveRequestRepository,
	quota repositories.LeaveQuotaRepository,
	attendance repositories.AttendanceRepository,
	announcements repositories.AnnouncementRepository,
	holidays repositories.HolidayRepository,
	notifications repositories.NotificationRepository,
) *DashboardService {
	return &DashboardService{
		cfg:           cfg,
		emps:          emps,
		leaves:        leaves,
		quota:         quota,
		attendance:    attendance,
		announcements: announcements,
		holidays:      holidays,
		notifications: notifications,
	}
}
```

- [ ] **Step 3: Populate the count in Get**

In `DashboardService.Get`, immediately before the final `out.Empty = len(out.Widgets) == 0` line, add:

```go
	// The bell count is independent of the widget list — an employee with no
	// visible widgets can still have unread notifications.
	if s.notifications != nil {
		unread, err := s.notifications.CountUnread(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		out.UnreadNotificationCount = unread
	}
```

- [ ] **Step 4: Fix the two call sites**

`cmd/server/main.go:136` — append `notificationRepo`:

```go
	dashboardSvc := services.NewDashboardService(cfg, employeeRepo, leaveRepo, quotaRepo, attendanceRepo, announcementRepo, holidayRepo, notificationRepo)
```

`internal/services/dashboard_service_test.go:21` — open the file, find the `services.NewDashboardService(` call, and append `repositories.NewNotificationRepository(testDB),` as the final argument.

- [ ] **Step 5: Add a dashboard test**

Append to `internal/services/notification_service_test.go`:

```go
// DR Rule 14 — the count reaches the dashboard so the header bell renders on
// first paint without a second request.
func TestNotification_DashboardCarriesUnreadCount(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	u := makeUser(t, "dash@example.com", "pw-Aa123456")
	makeEmployee(t, u, "Dash User")

	seedNotification(t, notifRepo, u.ID, models.NotificationTypeAnnouncement, "d1", uuid.New())
	seedNotification(t, notifRepo, u.ID, models.NotificationTypeLeaveRequest, "d2", uuid.New())

	dashSvc := services.NewDashboardService(
		nil,
		repositories.NewEmployeeRepository(testDB),
		repositories.NewLeaveRequestRepository(testDB),
		repositories.NewLeaveQuotaRepository(testDB),
		repositories.NewAttendanceRepository(testDB),
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewHolidayRepository(testDB),
		notifRepo,
	)

	full, err := testUserRepo.FindByID(context.Background(), u.ID)
	require.NoError(t, err)

	out, err := dashSvc.Get(context.Background(), full)
	require.NoError(t, err)
	assert.Equal(t, int64(2), out.UnreadNotificationCount)
}
```

If `NewAttendanceRepository` takes different arguments, copy the exact construction from `dashboard_service_test.go` rather than guessing.

- [ ] **Step 6: Run the tests**

```bash
go test ./internal/services -run 'TestNotification_|TestDashboard' -v
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dto/dashboard.go internal/services/dashboard_service.go internal/services/dashboard_service_test.go internal/services/notification_service_test.go cmd/server/main.go
git commit -m "feat(notifications): unread count on the dashboard response"
```

---

## Task 13: Snapshot and orphan tests

These two encode *why* fan-out was chosen over a derived view. Without them, a future refactor to a join-based read model would pass every other test.

**Files:**
- Modify: `internal/services/notification_service_test.go`

- [ ] **Step 1: Write the tests**

Append to `internal/services/notification_service_test.go`:

```go
// DR Rule 12 — notification content is a snapshot. Editing the source after
// the fact must not rewrite what the employee was already told. This test is
// the reason the table stores title/body instead of joining to the source.
func TestNotification_SnapshotSurvivesSourceEdit(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)

	u := makeUser(t, "snapshot@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	require.NoError(t, notifSvc.CreateMany(context.Background(), []models.Notification{{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "Original title",
		Body:     "Original body",
		SourceID: sourceID,
	}}))

	// Simulate the source record changing underneath the notification.
	require.NoError(t, testDB.Exec(
		`UPDATE announcements SET title = 'Edited title' WHERE id = ?`, sourceID,
	).Error)

	out, aerr := notifSvc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "Original title", out.Items[0].Title,
		"notification text is a snapshot and must not follow source edits")
}

// DR Rule 13 — the notification outlives its source. Tapping it should show
// "no longer available" rather than the row vanishing from the list.
func TestNotification_SurvivesSourceDeletion(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	notifRepo := repositories.NewNotificationRepository(testDB)
	notifSvc := services.NewNotificationService(notifRepo)

	u := makeUser(t, "orphan@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	require.NoError(t, notifSvc.CreateMany(context.Background(), []models.Notification{{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "Orphaned soon",
		Body:     "body",
		SourceID: sourceID,
	}}))

	// Hard-delete anything that might have shared the id. There is no FK, so
	// this must not cascade.
	require.NoError(t, testDB.Exec(`DELETE FROM announcements WHERE id = ?`, sourceID).Error)

	out, aerr := notifSvc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1, "deleting the source must not remove the notification")
	assert.Equal(t, "Orphaned soon", out.Items[0].Title)

	// And it is still markable.
	_, aerr = notifSvc.MarkRead(context.Background(), out.Items[0].ID, u.ID)
	require.Nil(t, aerr)
}
```

The `UPDATE`/`DELETE` statements target rows that do not exist (the `sourceID` is a fresh UUID), which is exactly the point: they are no-ops that prove the notification has no dependency on the source table.

- [ ] **Step 2: Run the full notification suite**

```bash
go test ./internal/services -run 'TestNotification_' -v
```

Expected: all PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/services/notification_service_test.go
git commit -m "test(notifications): snapshot and orphan-source behaviour"
```

---

## Task 14: Swagger, formatting, full suite

**Files:**
- Modify: `docs/swagger/` (generated)

- [ ] **Step 1: Format and vet**

```bash
make fmt
go vet ./...
```

Expected: no output from vet.

- [ ] **Step 2: Regenerate Swagger**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
make swag
grep -c "mobile/notifications" docs/swagger/swagger.json
```

Expected: at least 3 (list, unread-count, mark-read).

- [ ] **Step 3: Run the full suite and compare against the Task 0 baseline**

```bash
export TEST_DATABASE_URL="postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable"
go test ./... 2>&1 | tail -30
```

Expected: every package `ok` except `internal/services`, which fails with **exactly the four pre-existing leave failures** from Task 0 and nothing else.

Report the result honestly. "Tests pass" is false while those four are red — the correct statement is "all new tests pass; the four pre-existing leave failures from the Task 0 baseline are unchanged."

- [ ] **Step 4: Commit**

```bash
git add docs/swagger
git commit -m "docs(swagger): regenerate for notification endpoints"
```

---

## Task 15: Live HTTP verification

Unit tests are not verification in this repo. A phase is done when real requests have exercised the flow and DB state has been spot-checked.

**Files:**
- Create: `docs/superpowers/verification/mobile-notifications.md`

- [ ] **Step 1: Apply the migration to the dev DB**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
migrate -path migrations -database "postgres://postgres:devpassword@localhost:5432/exnodes_hrm?sslmode=disable" up
migrate -path migrations -database "postgres://postgres:devpassword@localhost:5432/exnodes_hrm?sslmode=disable" version
```

Expected: `28`, not dirty.

- [ ] **Step 2: Start the server**

Port 8080 is held by `ennam-kg-server` in this dev environment, so use 8082:

```bash
PORT=8082 go run ./cmd/server
```

- [ ] **Step 3: Run the smoke flow**

In a second shell, with `TOKEN` set to a logged-in employee's access token:

```bash
BASE=http://localhost:8082/api/v1

# 1. Empty state
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications" | jq

# 2. Unread count starts at 0
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications/unread-count" | jq

# 3. Publish an announcement as an admin, then re-list as the recipient
#    (use the admin token for the publish call)
curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"Smoke test announcement","description":"<p>hello</p>"}' \
  "$BASE/announcements" | jq -r '.data.id'
curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE/announcements/<ID>/publish" | jq

# 4. Recipient now has one unread notification
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications" | jq
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications/unread-count" | jq

# 5. Mark it read, then confirm the count drops
curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications/<NOTIF_ID>/read" | jq
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications/unread-count" | jq

# 6. Re-mark is a 200 no-op, not a 409
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H "Authorization: Bearer $TOKEN" \
  "$BASE/mobile/notifications/<NOTIF_ID>/read"

# 7. Another employee's notification ID returns 404
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H "Authorization: Bearer $OTHER_TOKEN" \
  "$BASE/mobile/notifications/<NOTIF_ID>/read"

# 8. Dashboard carries the count
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/dashboard" | jq '.data.unread_notification_count'

# 9. Create a leave request as the employee, approve it as an approver,
#    then re-list as the employee
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"from_date":"2026-08-03","to_date":"2026-08-05","leave_type":"annual","reason":"smoke test"}' \
  "$BASE/leave-requests" | jq -r '.data.id'
curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE/leave-requests/<LEAVE_ID>/approve" | jq
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/notifications" | jq '.data.items[0]'
```

Expected: step 6 prints `200`, step 7 prints `404`, and step 9's final call shows `"title": "Leave Request Approved"` with a body naming 2026-08-03 to 2026-08-05 and `source_id` equal to `<LEAVE_ID>`.

If the leave create/approve paths differ, confirm them against the registered routes rather than guessing: `grep -n 'leaves\.\|leave-requests' cmd/server/main.go`.

- [ ] **Step 4: Spot-check the DB**

```bash
psql "postgres://postgres:devpassword@localhost:5432/exnodes_hrm" -c \
  "SELECT type, title, source_id, read_at IS NOT NULL AS is_read, created_at
     FROM notifications ORDER BY created_at DESC LIMIT 10;"
```

Confirm the leave row's `title` reads `Leave Request Approved` and its `source_id` matches the leave request ID.

- [ ] **Step 5: Write the verification log**

Create `docs/superpowers/verification/mobile-notifications.md` with the actual commands run, the actual responses (not idealised ones), the DB spot-check output, and an explicit note that the four pre-existing leave test failures were present before this work and remain unchanged.

- [ ] **Step 6: Commit**

```bash
git add docs/superpowers/verification/mobile-notifications.md
git commit -m "docs(verification): mobile notifications end-to-end log"
```

---

## Task 16: Checkpoint and memories

Mandatory per AGENTS.md Rule 10 — never skipped, even if the session failed.

- [ ] **Step 1: Update CHECKPOINT.md**

Replace the relevant section of `docs/superpowers/CHECKPOINT.md` in place (do not append a sibling file). Record: the branch and commits, migration 000028 applied to test and dev, what was verified with evidence, and the follow-ups below.

While there, fix the stale line 162 — it says the latest migration is 000024 and the next is 000025, which contradicts line 35 and the actual tree. After this work the latest is **000028** and the next free is **000029**.

- [ ] **Step 2: Update the Serena memories**

```
mcp__serena__write_memory("checkpoint/main-session-2026-07-20", ...)
```

Also update `project_overview` — add the notifications module to the module list and the new files to the code map.

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/CHECKPOINT.md
git commit -m "docs(checkpoint): mobile notifications module"
```

---

## Follow-ups (not in this plan)

| Item | Why deferred |
|---|---|
| Fix the 4 pre-existing leave test failures | Unrelated to this work; blocks a clean suite |
| BA sign-off on the pagination deviation from DR §8 | Product decision |
| BA decision on client-side vs server-side icon registry (AC-20) | Changes the response shape if reversed |
| Retention/purge policy (DR Rule 11) | PO decision; table grows unbounded until then |
| OS-level push already exists but is separate from this feed | Out of scope per DR §8 |
| Announcement `target_audience` resolution runs on every publish | Existing behaviour, not made worse here |
