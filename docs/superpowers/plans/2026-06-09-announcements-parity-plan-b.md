# Announcements Parity — Plan B (G2: Push + Email Dispatch) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire email + FCM push notifications into announcement publish so every recipient is notified when an announcement goes live — matching Python's `_dispatch_notifications` behaviour.

**Architecture:** Add an `AnnouncementNotifier` interface to `announcement_service.go` (nil-safe, same pattern as `HubBroadcaster`). `broadcastPublished` launches a goroutine after the SSE broadcast. The goroutine resolves recipient `user_id`s from the preloaded `target_audience` fields, then calls `notifier.NotifyAnnouncement`. The concrete `announcementNotifier` calls the existing `PushNotificationService.SendToUser` and a new `EmailService.SendAnnouncementNotification`. Three new employee-repo query methods bridge the audience → user-ID translation. All notification errors are logged and swallowed — publish already succeeded.

**Tech Stack:** Go 1.25, Gin, GORM, testify, gomail (SMTP), FCM via existing `PushClient`

---

## ⚠️ REVISION NOTES

_None — this is the initial plan._

---

## File Map

| File | Action | What changes |
|------|--------|-------------|
| `internal/repositories/employee_repo.go` | Modify | Add `FindAllActive`, `FindByIDs`, `FindByDepartmentIDs` to interface + impl |
| `internal/services/email_service.go` | Modify | Add `SendAnnouncementNotification` method |
| `internal/services/announcement_service.go` | Modify | Add `AnnouncementNotifier` interface, `notifier` field, update constructor, add `dispatchNotifications` + `resolveRecipientUserIDs` methods, update `broadcastPublished` |
| `internal/services/announcement_notifier.go` | **Create** | Concrete `announcementNotifier` struct + `NewAnnouncementNotifier` |
| `cmd/server/main.go` | Modify | Construct `annNotifier` and pass to `NewAnnouncementService` |
| `internal/services/announcement_service_test.go` | Modify | Update `newAnnouncementSvc` helper + add dispatch/resolve tests |

---

## Task 1: Employee repo — `FindAllActive`, `FindByIDs`, `FindByDepartmentIDs`

**Files:**
- Modify: `internal/repositories/employee_repo.go`

- [ ] **Step 1.1: Add three methods to the `EmployeeRepository` interface**

In `internal/repositories/employee_repo.go`, append three method signatures to the `EmployeeRepository` interface, before the `WithTx` line:

```go
	// Notification dispatch helpers — used by AnnouncementService to
	// resolve target_audience to a slice of user_ids.
	FindAllActive(ctx context.Context) ([]models.Employee, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Employee, error)
	FindByDepartmentIDs(ctx context.Context, deptIDs []uuid.UUID) ([]models.Employee, error)
```

- [ ] **Step 1.2: Write failing tests for all three methods**

Add to `internal/services/announcement_service_test.go` (or create a new file `internal/repositories/employee_repo_notify_test.go` — pick whichever file the other employee repo tests live in; check with `grep -rl "TestEmployee" internal/`).

Add the following integration tests:

```go
func TestEmployeeRepo_FindAllActive_ReturnsNonDeleted(t *testing.T) {
    repo := repositories.NewEmployeeRepository(testDB)
    u1, e1 := makeEmpUser(t, "findall-active-1@test.com")
    _, e2  := makeEmpUser(t, "findall-active-2@test.com")
    // soft-delete e2
    require.NoError(t, repo.SoftDelete(context.Background(), e2.ID))

    emps, err := repo.FindAllActive(context.Background())
    require.NoError(t, err)

    ids := make(map[uuid.UUID]bool)
    for _, e := range emps {
        ids[e.ID] = true
    }
    assert.True(t, ids[e1.ID], "active employee should be returned")
    assert.False(t, ids[e2.ID], "deleted employee should not be returned")
    _ = u1
}

func TestEmployeeRepo_FindByIDs_ReturnsMatchingNonDeleted(t *testing.T) {
    repo := repositories.NewEmployeeRepository(testDB)
    _, e1 := makeEmpUser(t, "findbyids-1@test.com")
    _, e2 := makeEmpUser(t, "findbyids-2@test.com")
    _, e3 := makeEmpUser(t, "findbyids-3@test.com")
    require.NoError(t, repo.SoftDelete(context.Background(), e3.ID))

    emps, err := repo.FindByIDs(context.Background(), []uuid.UUID{e1.ID, e2.ID, e3.ID})
    require.NoError(t, err)

    ids := make(map[uuid.UUID]bool)
    for _, e := range emps {
        ids[e.ID] = true
    }
    assert.True(t, ids[e1.ID])
    assert.True(t, ids[e2.ID])
    assert.False(t, ids[e3.ID], "soft-deleted employee should be excluded")
}

func TestEmployeeRepo_FindByIDs_EmptySlice_ReturnsNil(t *testing.T) {
    repo := repositories.NewEmployeeRepository(testDB)
    emps, err := repo.FindByIDs(context.Background(), []uuid.UUID{})
    require.NoError(t, err)
    assert.Empty(t, emps)
}

func TestEmployeeRepo_FindByDepartmentIDs_ReturnsMatchingNonDeleted(t *testing.T) {
    repo := repositories.NewEmployeeRepository(testDB)
    dept := makeRawDept(t, "notify-dept")
    _, e1 := makeEmpUserInDept(t, "notifydept-1@test.com", "Emp One", dept.ID)
    _, e2 := makeEmpUserInDept(t, "notifydept-2@test.com", "Emp Two", dept.ID)
    _, e3 := makeEmpUser(t, "notifydept-nodept@test.com") // different dept

    emps, err := repo.FindByDepartmentIDs(context.Background(), []uuid.UUID{dept.ID})
    require.NoError(t, err)

    ids := make(map[uuid.UUID]bool)
    for _, e := range emps {
        ids[e.ID] = true
    }
    assert.True(t, ids[e1.ID])
    assert.True(t, ids[e2.ID])
    assert.False(t, ids[e3.ID], "employee in different dept should not be returned")
}
```

Check that `makeEmpUser` and `makeEmpUserInDept` helpers are available in the test package. They are already defined in `announcement_service_test.go` (or nearby); if the new tests are in a separate file, move or duplicate the helpers as needed.

- [ ] **Step 1.3: Run failing tests**

```bash
cd e:\Work\Go-HRM
go test ./internal/... -run "TestEmployeeRepo_FindAllActive|TestEmployeeRepo_FindByIDs|TestEmployeeRepo_FindByDepartmentIDs" -v
```

Expected: compile error — methods not yet on interface/impl.

- [ ] **Step 1.4: Implement the three methods on `employeeRepository`**

Add to `internal/repositories/employee_repo.go` (after the last existing `func (r *employeeRepository)` block):

```go
func (r *employeeRepository) FindAllActive(ctx context.Context) ([]models.Employee, error) {
	var emps []models.Employee
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Find(&emps).Error
	return emps, err
}

func (r *employeeRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Employee, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var emps []models.Employee
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("id IN ?", ids).
		Find(&emps).Error
	return emps, err
}

func (r *employeeRepository) FindByDepartmentIDs(ctx context.Context, deptIDs []uuid.UUID) ([]models.Employee, error) {
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var emps []models.Employee
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("department_id IN ?", deptIDs).
		Find(&emps).Error
	return emps, err
}
```

- [ ] **Step 1.5: Run tests — expect PASS**

```bash
go test ./internal/... -run "TestEmployeeRepo_FindAllActive|TestEmployeeRepo_FindByIDs|TestEmployeeRepo_FindByDepartmentIDs" -v
```

Expected: all 4 tests **PASS**.

- [ ] **Step 1.6: Commit**

```bash
git add internal/repositories/employee_repo.go internal/services/announcement_service_test.go
git commit -m "$(cat <<'EOF'
feat(employee-repo): add FindAllActive / FindByIDs / FindByDepartmentIDs

Three query helpers needed by AnnouncementService to resolve
target_audience to user_id slices for push+email dispatch. All three
respect the NotDeleted scope. Empty-slice fast-path on FindByIDs and
FindByDepartmentIDs avoids unnecessary DB round-trips.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Email service — `SendAnnouncementNotification`

**Files:**
- Modify: `internal/services/email_service.go`

- [ ] **Step 2.1: Add `AnnouncementEmailData` and `SendAnnouncementNotification`**

Append to `internal/services/email_service.go`:

```go
// SendAnnouncementNotification sends a plain HTML announcement email to
// toEmail. Returns ErrEmailDisabled when SMTP is not configured (same
// behaviour as SendInvite — caller logs and continues).
func (s *EmailService) SendAnnouncementNotification(ctx context.Context, toEmail, title, description string) error {
	if !s.IsConfigured() {
		log.Printf("email: skipped announcement notification to %s — SMTP not configured", toEmail)
		return ErrEmailDisabled
	}

	appName := s.cfg.AppName
	if appName == "" {
		appName = "HRM"
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;line-height:1.6;color:#333;max-width:600px;margin:0 auto;padding:20px">
  <div style="background:#f8f9fa;padding:30px;border-radius:10px">
    <h2 style="color:#2563eb">%s</h2>
    <div style="line-height:1.6">%s</div>
    <hr style="border:none;border-top:1px solid #ddd;margin:20px 0">
    <p style="color:#999;font-size:12px">This is an automated message from %s.</p>
  </div>
</body>
</html>`, title, description, appName)

	plainText := fmt.Sprintf("%s\n\n%s\n\n-- %s", title, description, appName)

	from := s.cfg.SMTPFromEmail
	if from == "" {
		from = "no-reply@" + appName
	}
	fromAddr := from
	if s.cfg.SMTPFromName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, from)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromAddr)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] New Announcement: %s", appName, title))
	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)
	d.SSL = false
	if !s.cfg.SMTPUseTLS {
		d.TLSConfig = nil
	}

	done := make(chan error, 1)
	go func() { done <- d.DialAndSend(m) }()

	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("smtp send announcement: %w", err)
		}
		return nil
	case <-deadline.Done():
		return fmt.Errorf("smtp send announcement: timeout after 10s")
	}
}
```

- [ ] **Step 2.2: Build to confirm it compiles**

```bash
go build ./internal/services/...
```

Expected: exits 0.

- [ ] **Step 2.3: Commit**

```bash
git add internal/services/email_service.go
git commit -m "$(cat <<'EOF'
feat(email): add SendAnnouncementNotification

Sends an HTML+plaintext email for a published announcement. Same
gomail/SMTP pattern as SendInvite: 10s timeout, ErrEmailDisabled
when SMTP_HOST is empty. Called by AnnouncementNotifier on publish.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `AnnouncementNotifier` interface + service dispatch methods

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 3.1: Write failing tests for `resolveRecipientUserIDs`**

Add to `internal/services/announcement_service_test.go`:

```go
// ---- AnnouncementNotifier mock ----

type captureNotifier struct {
	mu    sync.Mutex
	calls []notifyCall
}

type notifyCall struct {
	UserIDs     []uuid.UUID
	Title       string
	Description string
}

func (n *captureNotifier) NotifyAnnouncement(_ context.Context, userIDs []uuid.UUID, title, description string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{UserIDs: userIDs, Title: title, Description: description})
}

func (n *captureNotifier) CallCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.calls)
}

func (n *captureNotifier) LastCall() notifyCall {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls[len(n.calls)-1]
}

func newAnnouncementSvcWithNotifier(t *testing.T, notifier services.AnnouncementNotifier) *services.AnnouncementService {
	t.Helper()
	hub := &captureHub{}
	return services.NewAnnouncementService(
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewLabelRepository(testDB),
		hub,
		notifier,
	)
}

// ---- resolveRecipientUserIDs tests ----

func TestAnnouncement_ResolveRecipients_All(t *testing.T) {
	notifier := &captureNotifier{}
	svc := newAnnouncementSvcWithNotifier(t, notifier)
	authorUser, authorEmp := makeEmpUser(t, "resolve-all-author@test.com")
	u1, _ := makeEmpUser(t, "resolve-all-1@test.com")
	u2, _ := makeEmpUser(t, "resolve-all-2@test.com")

	// publish with target_audience=all using send_now shortcut
	out, err := svc.Create(context.Background(), authorUser.ID, dto.AnnouncementCreate{
		Title:       "All audience",
		Description: "body",
		SendNow:     true,
	})
	require.NoError(t, err)
	require.NotNil(t, out)

	// wait briefly for goroutine
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 1, notifier.CallCount())
	call := notifier.LastCall()
	userIDSet := make(map[uuid.UUID]bool)
	for _, id := range call.UserIDs {
		userIDSet[id] = true
	}
	assert.True(t, userIDSet[authorUser.ID])
	assert.True(t, userIDSet[u1.ID])
	assert.True(t, userIDSet[u2.ID])
	_ = authorEmp
}

func TestAnnouncement_ResolveRecipients_Department(t *testing.T) {
	notifier := &captureNotifier{}
	svc := newAnnouncementSvcWithNotifier(t, notifier)
	authorUser, _ := makeEmpUser(t, "resolve-dept-author@test.com")
	dept := makeRawDept(t, "resolve-dept")
	deptUser, _ := makeEmpUserInDept(t, "resolve-dept-member@test.com", "Member", dept.ID)
	outsideUser, _ := makeEmpUser(t, "resolve-dept-outside@test.com")

	out, err := svc.Create(context.Background(), authorUser.ID, dto.AnnouncementCreate{
		Title:          "Dept audience",
		Description:    "body",
		SendNow:        true,
		TargetAudience: ptr(models.AnnouncementAudienceDepartment),
		DepartmentIDs:  []uuid.UUID{dept.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, out)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 1, notifier.CallCount())
	call := notifier.LastCall()
	userIDSet := make(map[uuid.UUID]bool)
	for _, id := range call.UserIDs {
		userIDSet[id] = true
	}
	assert.True(t, userIDSet[deptUser.ID], "department member should be notified")
	assert.False(t, userIDSet[outsideUser.ID], "outside-dept user should not be notified")
}

func TestAnnouncement_ResolveRecipients_Custom(t *testing.T) {
	notifier := &captureNotifier{}
	svc := newAnnouncementSvcWithNotifier(t, notifier)
	authorUser, _ := makeEmpUser(t, "resolve-custom-author@test.com")
	_, recipEmp := makeEmpUser(t, "resolve-custom-recip@test.com")
	_, otherEmp := makeEmpUser(t, "resolve-custom-other@test.com")

	out, err := svc.Create(context.Background(), authorUser.ID, dto.AnnouncementCreate{
		Title:          "Custom audience",
		Description:    "body",
		SendNow:        true,
		TargetAudience: ptr(models.AnnouncementAudienceCustom),
		RecipientIDs:   []uuid.UUID{recipEmp.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, out)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 1, notifier.CallCount())
	call := notifier.LastCall()
	userIDSet := make(map[uuid.UUID]bool)
	for _, id := range call.UserIDs {
		userIDSet[id] = true
	}
	assert.True(t, userIDSet[recipEmp.UserID], "named recipient should be notified")
	assert.False(t, userIDSet[otherEmp.UserID], "non-recipient should not be notified")
}

func TestAnnouncement_NilNotifier_DoesNotPanic(t *testing.T) {
	svc := newAnnouncementSvcWithNotifier(t, nil)
	_, emp := makeEmpUser(t, "nil-notifier@test.com")

	_, err := svc.Create(context.Background(), emp.UserID, dto.AnnouncementCreate{
		Title:       "Silent publish",
		Description: "body",
		SendNow:     true,
	})
	require.NoError(t, err) // must not panic
}
```

Add the `ptr` generic helper if not already present in the test file:

```go
func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 3.2: Run failing tests**

```bash
go test ./internal/services/... -run "TestAnnouncement_ResolveRecipients|TestAnnouncement_NilNotifier" -v
```

Expected: compile errors — `AnnouncementNotifier` interface and `newAnnouncementSvcWithNotifier` do not exist yet; `NewAnnouncementService` does not accept a 6th argument.

- [ ] **Step 3.3: Add `AnnouncementNotifier` interface + struct field + constructor update**

In `internal/services/announcement_service.go`, after the `HubBroadcaster` interface definition, add:

```go
// AnnouncementNotifier dispatches push + email notifications when an
// announcement is published. Nil is valid — disables notifications
// (used in tests). Same nil-safe pattern as HubBroadcaster.
type AnnouncementNotifier interface {
	NotifyAnnouncement(ctx context.Context, userIDs []uuid.UUID, title, description string)
}
```

Add `notifier AnnouncementNotifier` to the `AnnouncementService` struct:

```go
type AnnouncementService struct {
	repo     repositories.AnnouncementRepository
	emps     repositories.EmployeeRepository
	depts    repositories.DepartmentRepository
	labels   repositories.LabelRepository
	hub      HubBroadcaster
	notifier AnnouncementNotifier // nil = no push/email notifications
}
```

Update `NewAnnouncementService` to accept and store `notifier`:

```go
func NewAnnouncementService(
	repo repositories.AnnouncementRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	labels repositories.LabelRepository,
	hub HubBroadcaster,
	notifier AnnouncementNotifier,
) *AnnouncementService {
	return &AnnouncementService{repo: repo, emps: emps, depts: depts, labels: labels, hub: hub, notifier: notifier}
}
```

- [ ] **Step 3.4: Add `dispatchNotifications` and `resolveRecipientUserIDs` methods**

Add after `broadcastPublished` in `announcement_service.go`:

```go
// dispatchNotifications resolves recipient user IDs from the
// announcement's target_audience and fires NotifyAnnouncement. Runs in
// a goroutine (called from broadcastPublished) — uses context.Background
// so it outlives the request. All errors are logged; publish already succeeded.
func (s *AnnouncementService) dispatchNotifications(ann *models.Announcement) {
	ctx := context.Background()
	userIDs, err := s.resolveRecipientUserIDs(ctx, ann)
	if err != nil {
		log.Printf("announcements: dispatchNotifications: resolve for %s: %v", ann.ID, err)
		return
	}
	if len(userIDs) == 0 {
		return
	}
	s.notifier.NotifyAnnouncement(ctx, userIDs, ann.Title, ann.Description)
}

// resolveRecipientUserIDs translates the announcement's target_audience
// + preloaded join rows to a slice of user_ids for notification dispatch.
func (s *AnnouncementService) resolveRecipientUserIDs(ctx context.Context, ann *models.Announcement) ([]uuid.UUID, error) {
	switch ann.TargetAudience {
	case models.AnnouncementAudienceAll:
		emps, err := s.emps.FindAllActive(ctx)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, 0, len(emps))
		for _, e := range emps {
			ids = append(ids, e.UserID)
		}
		return ids, nil
	case models.AnnouncementAudienceDepartment:
		deptIDs := make([]uuid.UUID, 0, len(ann.TargetDepartments))
		for _, td := range ann.TargetDepartments {
			deptIDs = append(deptIDs, td.DepartmentID)
		}
		emps, err := s.emps.FindByDepartmentIDs(ctx, deptIDs)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, 0, len(emps))
		for _, e := range emps {
			ids = append(ids, e.UserID)
		}
		return ids, nil
	case models.AnnouncementAudienceCustom:
		empIDs := make([]uuid.UUID, 0, len(ann.TargetUsers))
		for _, tu := range ann.TargetUsers {
			empIDs = append(empIDs, tu.EmployeeID)
		}
		emps, err := s.emps.FindByIDs(ctx, empIDs)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, 0, len(emps))
		for _, e := range emps {
			ids = append(ids, e.UserID)
		}
		return ids, nil
	default:
		return nil, nil
	}
}
```

Add `"log"` to the import block in `announcement_service.go` if not already present.

- [ ] **Step 3.5: Update `broadcastPublished` to launch the goroutine**

Inside `broadcastPublished`, after `s.hub.Broadcast(...)`, add:

```go
	// Dispatch push + email notifications off the request path.
	if s.notifier != nil {
		go s.dispatchNotifications(a)
	}
```

The full `broadcastPublished` function should now end with:

```go
	s.hub.Broadcast("announcement_published", payload, nil)
	// Dispatch push + email notifications off the request path.
	if s.notifier != nil {
		go s.dispatchNotifications(a)
	}
}
```

- [ ] **Step 3.6: Fix the existing `newAnnouncementSvc` test helper**

In `announcement_service_test.go`, `newAnnouncementSvc` calls `NewAnnouncementService` with 5 args. Add `nil` as the 6th:

```go
func newAnnouncementSvc(t *testing.T) (*services.AnnouncementService, *captureHub) {
	t.Helper()
	hub := &captureHub{}
	svc := services.NewAnnouncementService(
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewLabelRepository(testDB),
		hub,
		nil, // notifier — nil disables push/email in unit tests
	)
	return svc, hub
}
```

- [ ] **Step 3.7: Run failing tests — expect PASS**

```bash
go test ./internal/services/... -run "TestAnnouncement_ResolveRecipients|TestAnnouncement_NilNotifier" -v -timeout 30s
```

Expected: all 4 tests **PASS**. The goroutine uses a 100ms sleep in the test; if timing is flaky on CI, increase to 200ms or use a channel-based mock.

- [ ] **Step 3.8: Run full service suite to check regressions**

```bash
go test ./internal/services/... -v -timeout 300s 2>&1 | tail -30
```

Expected: all tests pass.

- [ ] **Step 3.9: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "$(cat <<'EOF'
feat(announcements): AnnouncementNotifier interface + dispatch goroutine (G2)

Add AnnouncementNotifier interface (nil-safe, same pattern as
HubBroadcaster). broadcastPublished now launches go dispatchNotifications
after the SSE broadcast. resolveRecipientUserIDs translates
target_audience (all/department/custom) to []uuid.UUID user_ids using
the three new employee repo helpers. All errors logged, never
propagated — publish already succeeded.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Concrete `announcement_notifier.go`

**Files:**
- Create: `internal/services/announcement_notifier.go`

- [ ] **Step 4.1: Create the concrete implementation**

Create `internal/services/announcement_notifier.go`:

```go
package services

import (
	"context"
	"log"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// announcementNotifier is the production AnnouncementNotifier. It
// dispatches FCM push (via PushNotificationService) and SMTP email
// (via EmailService) to every resolved recipient. All errors are logged
// and swallowed — the caller's goroutine is fire-and-forget.
type announcementNotifier struct {
	push  *PushNotificationService
	email *EmailService
	users repositories.UserRepository
}

// NewAnnouncementNotifier constructs the production notifier.
// Pass nil for push or email to disable that channel individually.
func NewAnnouncementNotifier(
	push *PushNotificationService,
	email *EmailService,
	users repositories.UserRepository,
) AnnouncementNotifier {
	return &announcementNotifier{push: push, email: email, users: users}
}

// NotifyAnnouncement sends a push + email to every user in userIDs.
// Errors for individual users are logged; the loop continues regardless.
func (n *announcementNotifier) NotifyAnnouncement(ctx context.Context, userIDs []uuid.UUID, title, description string) {
	for _, uid := range userIDs {
		// FCM push — skips silently when client is not configured.
		if n.push != nil {
			req := dto.NotificationTestRequest{Title: title, Body: description}
			if _, err := n.push.SendToUser(ctx, uid, req); err != nil {
				log.Printf("announcements: push to user %s: %v", uid, err)
			}
		}
		// Email — look up the address, skip on lookup failure.
		if n.email != nil {
			user, err := n.users.FindByID(ctx, uid)
			if err != nil {
				log.Printf("announcements: lookup user %s for email: %v", uid, err)
				continue
			}
			if user.Email == "" {
				continue
			}
			if err := n.email.SendAnnouncementNotification(ctx, user.Email, title, description); err != nil {
				log.Printf("announcements: email to %s: %v", user.Email, err)
			}
		}
	}
}
```

- [ ] **Step 4.2: Build to confirm it compiles**

```bash
go build ./internal/services/...
```

Expected: exits 0.

- [ ] **Step 4.3: Commit**

```bash
git add internal/services/announcement_notifier.go
git commit -m "$(cat <<'EOF'
feat(announcements): concrete AnnouncementNotifier (push + email)

announcementNotifier wraps PushNotificationService + EmailService.
NotifyAnnouncement loops over user IDs: FCM push via existing
SendToUser, email via new SendAnnouncementNotification. Per-user
errors are logged and swallowed; nil push/email fields disable
that channel without panicking.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Wire in `main.go` + final verification

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 5.1: Wire `annNotifier` into `NewAnnouncementService`**

In `cmd/server/main.go`, find the line:

```go
announcementSvc := services.NewAnnouncementService(
    announcementRepo, employeeRepo, departmentRepo, labelRepo,
    sseHubAdapter{hub: sseHub},
)
```

Replace it with:

```go
annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo)
announcementSvc := services.NewAnnouncementService(
    announcementRepo, employeeRepo, departmentRepo, labelRepo,
    sseHubAdapter{hub: sseHub},
    annNotifier,
)
```

Verify that `pushSvc`, `emailSvc`, and `userRepo` are already in scope at this point in `main.go` — they are wired for Phase 9 (invite + push notification). If the variable names differ, grep for them:

```bash
grep -n "pushSvc\|emailSvc\|userRepo\|PushNotificationService\|EmailService\|NewUserRepository" cmd/server/main.go | head -20
```

Use whatever variable names `main.go` actually uses.

- [ ] **Step 5.2: Build to confirm it compiles**

```bash
go build ./...
```

Expected: exits 0.

- [ ] **Step 5.3: Run full test suite**

```bash
make fmt && make vet && make test
```

Expected: `fmt` and `vet` exit 0. All tests pass — 0 failures.

- [ ] **Step 5.4: Smoke test notification dispatch (optional — requires Mailpit + a registered device token)**

Start the server:

```bash
PORT=8082 go run ./cmd/server &
sleep 2
```

Login and publish an announcement:

```bash
TOKEN=$(curl -s -X POST http://localhost:8082/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.data.access_token')

# Create a draft
ANN_ID=$(curl -s -X POST http://localhost:8082/api/v1/announcements \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Notify","description":"Hello from Plan B"}' \
  | jq -r '.data.id')

# Publish it
curl -s -X POST http://localhost:8082/api/v1/announcements/$ANN_ID/publish \
  -H "Authorization: Bearer $TOKEN" | jq
```

Check Mailpit at `http://localhost:18025` — expect one email per employee in the system.

```bash
kill %1
```

- [ ] **Step 5.5: Commit**

```bash
git add cmd/server/main.go
git commit -m "$(cat <<'EOF'
feat(announcements): wire AnnouncementNotifier in main.go (G2 complete)

Construct annNotifier from existing pushSvc + emailSvc + userRepo
and pass to NewAnnouncementService. Announcements now dispatch
FCM push + email to all recipients on publish, matching Python's
_dispatch_notifications behaviour.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Swagger + CHECKPOINT update

- [ ] **Step 6.1: Verify Swagger annotations**

The `Update` handler's godoc should mention `@Failure 409`. Check and add if missing:

```bash
grep -A 20 "// Update godoc" internal/handlers/announcement_handler.go
```

If `@Failure 409` is absent, add it:

```go
// @Failure      409  {object}  map[string]interface{}  "published or archived — cannot edit"
```

Then regenerate:

```bash
make swag
```

- [ ] **Step 6.2: Update CHECKPOINT.md**

Update `docs/superpowers/CHECKPOINT.md` to record:
- Plans A + B committed (list commit hashes)
- What is verified
- What is next (deploy, then Request Tickets EP-003)

- [ ] **Step 6.3: Plan B complete**

Both Plan A and Plan B are done. The announcements module is now at full Python parity for the three open gaps (G1 + G2 + G3).
