package services_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Helpers ----

// captureHub is a tiny HubBroadcaster mock that records every broadcast.
type captureHub struct {
	mu     sync.Mutex
	events []capturedEvent
}

type capturedEvent struct {
	Type   string
	Data   any
	Filter func(uuid.UUID) bool
}

func (h *captureHub) Broadcast(t string, data any, filter func(uuid.UUID) bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, capturedEvent{Type: t, Data: data, Filter: filter})
}

func (h *captureHub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

func (h *captureHub) Last() capturedEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.events[len(h.events)-1]
}

func newAnnouncementSvc(t *testing.T) (*services.AnnouncementService, *captureHub) {
	t.Helper()
	hub := &captureHub{}
	svc := services.NewAnnouncementService(
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewLabelRepository(testDB),
		hub,
		nil, // notifier — tests that don't need dispatch use nil
	)
	return svc, hub
}

// ---- captureNotifier mock ----

type captureNotifier struct {
	mu    sync.Mutex
	calls []notifyCall
}

type notifyCall struct {
	UserIDs     []uuid.UUID
	ID          uuid.UUID
	Title       string
	Description string
}

func (n *captureNotifier) NotifyAnnouncement(_ context.Context, userIDs []uuid.UUID, id uuid.UUID, title, description string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{UserIDs: userIDs, ID: id, Title: title, Description: description})
}

func (n *captureNotifier) callCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.calls)
}

func (n *captureNotifier) lastCall() notifyCall {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls[len(n.calls)-1]
}

// svcWithNotifier builds a service wired to a specific notifier.
func svcWithNotifier(t *testing.T, notifier services.AnnouncementNotifier) *services.AnnouncementService {
	t.Helper()
	return services.NewAnnouncementService(
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewLabelRepository(testDB),
		&captureHub{},
		notifier,
	)
}

func ptrAudience(v models.AnnouncementTargetAudience) *models.AnnouncementTargetAudience {
	return &v
}

// makeLabel inserts a label and returns it.
func makeLabel(t *testing.T, name string) *models.Label {
	t.Helper()
	l := &models.Label{Name: name}
	require.NoError(t, testDB.Create(l).Error)
	return l
}

// makeRawDept inserts a department and returns it.
func makeRawDept(t *testing.T, name string) *models.Department {
	t.Helper()
	d := &models.Department{Name: name}
	require.NoError(t, testDB.Create(d).Error)
	return d
}

// makeEmpUserInDept creates an employee linked to a department.
func makeEmpUserInDept(t *testing.T, email, fullName string, deptID uuid.UUID) (*models.User, *models.Employee) {
	t.Helper()
	u, e := makeEmpUser(t, email, fullName)
	e.DepartmentID = &deptID
	require.NoError(t, testDB.Save(e).Error)
	return u, e
}

// ---- Create + Publish ----

func TestAnnouncement_Create_DraftDefault(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	out, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:       "Hello",
		Description: "world",
	})
	require.NoError(t, err)
	assert.Equal(t, models.AnnouncementStatusDraft, out.Status)
	assert.Nil(t, out.PublishedAt)
	assert.Equal(t, 0, hub.Count(), "draft must NOT broadcast")
}

func TestAnnouncement_Create_PublishesAndBroadcasts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	published := models.AnnouncementStatusPublished
	out, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:       "Live",
		Description: "hello",
		Status:      &published,
	})
	require.NoError(t, err)
	assert.Equal(t, models.AnnouncementStatusPublished, out.Status)
	require.NotNil(t, out.PublishedAt)
	require.Equal(t, 1, hub.Count())
	assert.Equal(t, "announcement_published", hub.Last().Type)
}

func TestAnnouncement_Create_NoEmployeeProfile_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	// User without an Employee row.
	u := makeUser(t, "noemp@example.com", "pw-Aa123456")

	_, err := svc.Create(ctx, u.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No employee record")
}

func TestAnnouncement_Create_DepartmentAudienceRequiresTargets(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	dept := models.AnnouncementAudienceDepartment

	_, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:          "deptonly",
		Description:    "x",
		TargetAudience: &dept,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least one department_id")
}

func TestAnnouncement_Create_WithLabelsAndDepartments(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	l1 := makeLabel(t, "General")
	d1 := makeRawDept(t, "Engineering")

	dept := models.AnnouncementAudienceDepartment
	out, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:          "for-eng",
		Description:    "x",
		TargetAudience: &dept,
		LabelIDs:       []uuid.UUID{l1.ID},
		DepartmentIDs:  []uuid.UUID{d1.ID},
	})
	require.NoError(t, err)
	require.Len(t, out.Labels, 1)
	require.Len(t, out.TargetDepartments, 1)
	assert.Equal(t, l1.ID, out.Labels[0].ID)
	assert.Equal(t, d1.ID, out.TargetDepartments[0].ID)
}

func TestAnnouncement_Create_UnknownLabel_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	_, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:       "x",
		Description: "y",
		LabelIDs:    []uuid.UUID{uuid.New()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown label_id")
}

func TestAnnouncement_Publish_BroadcastsOnce(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	draft, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)
	require.Equal(t, 0, hub.Count())

	out, err := svc.Publish(ctx, draft.ID, admin.ID, true)
	require.NoError(t, err)
	assert.Equal(t, models.AnnouncementStatusPublished, out.Status)
	require.NotNil(t, out.PublishedAt)
	require.Equal(t, 1, hub.Count())

	// Second publish — already published, no rebroadcast.
	_, err = svc.Publish(ctx, draft.ID, admin.ID, true)
	require.NoError(t, err)
	assert.Equal(t, 1, hub.Count(), "already-published rows must NOT rebroadcast")
}

func TestAnnouncement_Publish_NonOwner_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	stranger, _ := makeEmpUser(t, "stranger@example.com", "Stranger")

	draft, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)

	_, err = svc.Publish(ctx, draft.ID, stranger.ID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "do not own")
}

// ---- Update + Delete ----

func TestAnnouncement_Update_PublishTransition_Broadcasts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	draft, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)
	require.Equal(t, 0, hub.Count())

	published := models.AnnouncementStatusPublished
	_, err = svc.Update(ctx, draft.ID, admin.ID, true, dto.AnnouncementUpdate{
		Status: &published,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, hub.Count(), "draft→published in Update must broadcast")
}

func TestAnnouncement_Update_AlreadyPublished_DoesNotRebroadcast(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	// Create a draft, update title (no broadcast), then publish (1 broadcast).
	// A second publish call must NOT rebroadcast.
	a, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)
	require.Equal(t, 0, hub.Count())

	// Promote draft → published via Update.
	pub := models.AnnouncementStatusPublished
	_, err = svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{Status: &pub})
	require.NoError(t, err)
	assert.Equal(t, 1, hub.Count(), "draft→published in Update must broadcast once")

	// Publish again via Publish — already published, must NOT rebroadcast.
	_, err = svc.Publish(ctx, a.ID, admin.ID, true)
	require.NoError(t, err)
	assert.Equal(t, 1, hub.Count(), "already-published rows must NOT rebroadcast")
}

func TestAnnouncement_Update_NonOwner_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	stranger, _ := makeEmpUser(t, "stranger@example.com", "Stranger")

	draft, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)

	newTitle := "hacked"
	_, err = svc.Update(ctx, draft.ID, stranger.ID, false, dto.AnnouncementUpdate{Title: &newTitle})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "do not own")
}

func TestAnnouncement_Update_LabelsReplaceSet(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	l1 := makeLabel(t, "L1")
	l2 := makeLabel(t, "L2")
	l3 := makeLabel(t, "L3")

	a, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:       "x",
		Description: "y",
		LabelIDs:    []uuid.UUID{l1.ID, l2.ID},
	})
	require.NoError(t, err)
	require.Len(t, a.Labels, 2)

	newSet := []uuid.UUID{l2.ID, l3.ID}
	out, err := svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{LabelIDs: &newSet})
	require.NoError(t, err)
	require.Len(t, out.Labels, 2)
	ids := []uuid.UUID{out.Labels[0].ID, out.Labels[1].ID}
	assert.Contains(t, ids, l2.ID)
	assert.Contains(t, ids, l3.ID)
	assert.NotContains(t, ids, l1.ID, "L1 must be soft-deleted from the join")
}

func TestAnnouncement_Delete_OwnerOnly(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	stranger, _ := makeEmpUser(t, "stranger@example.com", "Stranger")
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	a, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)

	// Stranger (non-admin) → forbidden.
	require.Error(t, svc.Delete(ctx, a.ID, stranger.ID, false))

	// Admin → ok.
	require.NoError(t, svc.Delete(ctx, a.ID, admin.ID, true))

	// Subsequent Get → 404.
	_, err = svc.Get(ctx, a.ID, admin.ID, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---- Visibility ----

func TestAnnouncement_Get_NonOwner_AllAudience_VisibleWhenPublished(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	reader, _ := makeEmpUser(t, "reader@example.com", "Reader")
	pub := models.AnnouncementStatusPublished

	a, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y", Status: &pub})
	require.NoError(t, err)

	out, err := svc.Get(ctx, a.ID, reader.ID, false)
	require.NoError(t, err)
	assert.Equal(t, a.ID, out.ID)
}

func TestAnnouncement_Get_NonOwner_DraftAllAudience_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	reader, _ := makeEmpUser(t, "reader@example.com", "Reader")

	draft, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y"})
	require.NoError(t, err)

	_, err = svc.Get(ctx, draft.ID, reader.ID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot view")
}

func TestAnnouncement_Get_DepartmentMatch_Visible(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	d1 := makeRawDept(t, "Eng")
	d2 := makeRawDept(t, "HR")
	author, _ := makeEmpUserInDept(t, "author@example.com", "Author", d1.ID)
	engReader, _ := makeEmpUserInDept(t, "eng-reader@example.com", "EngR", d1.ID)
	hrReader, _ := makeEmpUserInDept(t, "hr-reader@example.com", "HrR", d2.ID)

	pub := models.AnnouncementStatusPublished
	dept := models.AnnouncementAudienceDepartment
	a, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title: "for-eng", Description: "x",
		Status:         &pub,
		TargetAudience: &dept,
		DepartmentIDs:  []uuid.UUID{d1.ID},
	})
	require.NoError(t, err)

	// Eng reader → visible.
	_, err = svc.Get(ctx, a.ID, engReader.ID, false)
	require.NoError(t, err)

	// HR reader → forbidden.
	_, err = svc.Get(ctx, a.ID, hrReader.ID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot view")
}

// ---- View tracking ----

func TestAnnouncement_MarkViewed_Idempotent(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	reader, _ := makeEmpUser(t, "reader@example.com", "Reader")
	pub := models.AnnouncementStatusPublished

	a, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "x", Description: "y", Status: &pub})
	require.NoError(t, err)

	require.NoError(t, svc.MarkViewed(ctx, a.ID, reader.ID, false))
	// Second call must be a no-op (idempotent).
	require.NoError(t, svc.MarkViewed(ctx, a.ID, reader.ID, false))

	out, err := svc.Get(ctx, a.ID, reader.ID, false)
	require.NoError(t, err)
	assert.True(t, out.HasViewed)
}

// ---- List ----

func TestAnnouncement_List_AdminSeesAll(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	_, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "draft1", Description: "x"})
	require.NoError(t, err)
	pub := models.AnnouncementStatusPublished
	_, err = svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "live1", Description: "x", Status: &pub})
	require.NoError(t, err)

	out, err := svc.List(ctx, admin.ID, true, dto.AnnouncementListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(out.Items), 2)
}

func TestAnnouncement_List_NonAdmin_OnlyPublishedVisible(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	reader, _ := makeEmpUser(t, "reader@example.com", "Reader")
	pub := models.AnnouncementStatusPublished

	_, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "draft", Description: "x"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "live", Description: "x", Status: &pub})
	require.NoError(t, err)

	out, err := svc.List(ctx, reader.ID, false, dto.AnnouncementListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "live", out.Items[0].Title)
}

func TestAnnouncement_List_Mine_IncludesDrafts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")

	_, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "draft", Description: "x"})
	require.NoError(t, err)

	out, err := svc.List(ctx, author.ID, false, dto.AnnouncementListQuery{Page: 1, PageSize: 50, Scope: "mine"})
	require.NoError(t, err)
	require.Len(t, out.Items, 1, "mine scope must include own drafts")
}

// ---- Mobile ----

func TestAnnouncement_MobileList_OnlyPublished(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	reader, _ := makeEmpUser(t, "reader@example.com", "Reader")
	pub := models.AnnouncementStatusPublished

	_, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "draft", Description: "x"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, author.ID, dto.AnnouncementCreate{Title: "live", Description: "x", Status: &pub})
	require.NoError(t, err)

	out, err := svc.MobileList(ctx, reader.ID, dto.MobileAnnouncementListQuery{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "live", out.Items[0].Title)
}

// ---- Custom (per-user) audience — closes parity audit decision #6 ----

func TestAnnouncement_Create_CustomAudienceRequiresRecipients(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	custom := models.AnnouncementAudienceCustom

	_, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:          "custom-no-recipients",
		Description:    "x",
		TargetAudience: &custom,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least one recipient_id")
}

func TestAnnouncement_Create_WithRecipients(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	_, target1 := makeEmpUser(t, "t1@example.com", "Target One")
	_, target2 := makeEmpUser(t, "t2@example.com", "Target Two")
	l1 := makeLabel(t, "Personal")
	custom := models.AnnouncementAudienceCustom

	out, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:          "named",
		Description:    "x",
		TargetAudience: &custom,
		LabelIDs:       []uuid.UUID{l1.ID},
		RecipientIDs:   []uuid.UUID{target1.ID, target2.ID},
	})
	require.NoError(t, err)
	require.Len(t, out.Labels, 1)
	require.Len(t, out.TargetRecipients, 2)
	assert.Equal(t, models.AnnouncementAudienceCustom, out.TargetAudience)
	ids := []uuid.UUID{out.TargetRecipients[0].ID, out.TargetRecipients[1].ID}
	assert.Contains(t, ids, target1.ID)
	assert.Contains(t, ids, target2.ID)
}

func TestAnnouncement_Visibility_TargetedAtUser(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	userA, empA := makeEmpUser(t, "a@example.com", "A")
	userB, _ := makeEmpUser(t, "b@example.com", "B")
	custom := models.AnnouncementAudienceCustom
	pub := models.AnnouncementStatusPublished

	a, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:          "for-A-only",
		Description:    "x",
		Status:         &pub,
		TargetAudience: &custom,
		RecipientIDs:   []uuid.UUID{empA.ID},
	})
	require.NoError(t, err)

	// A is targeted → visible via Get and via targeted-at-me scope.
	_, err = svc.Get(ctx, a.ID, userA.ID, false)
	require.NoError(t, err)

	listA, err := svc.List(ctx, userA.ID, false, dto.AnnouncementListQuery{
		Page: 1, PageSize: 20, Scope: "targeted-at-me",
	})
	require.NoError(t, err)
	require.Len(t, listA.Items, 1)
	assert.Equal(t, a.ID, listA.Items[0].ID)

	// B is not targeted → forbidden on Get, absent from targeted-at-me list.
	_, err = svc.Get(ctx, a.ID, userB.ID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot view")

	listB, err := svc.List(ctx, userB.ID, false, dto.AnnouncementListQuery{
		Page: 1, PageSize: 20, Scope: "targeted-at-me",
	})
	require.NoError(t, err)
	assert.Len(t, listB.Items, 0)
}

func TestAnnouncement_SSE_BroadcastIncludesRecipientIDs(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, hub := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "author@example.com", "Author")
	_, empA := makeEmpUser(t, "a@example.com", "A")
	pub := models.AnnouncementStatusPublished
	custom := models.AnnouncementAudienceCustom

	_, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:          "ping-A",
		Description:    "x",
		Status:         &pub,
		TargetAudience: &custom,
		RecipientIDs:   []uuid.UUID{empA.ID},
	})
	require.NoError(t, err)

	require.Equal(t, 1, hub.Count())
	ev := hub.Last()
	assert.Equal(t, "announcement_published", ev.Type)
	payload, ok := ev.Data.(dto.SSEAnnouncementPublishedEvent)
	require.True(t, ok, "expected SSEAnnouncementPublishedEvent payload, got %T", ev.Data)
	assert.Equal(t, models.AnnouncementAudienceCustom, payload.TargetAudience)
	require.Len(t, payload.RecipientIDs, 1)
	assert.Equal(t, empA.ID, payload.RecipientIDs[0])
}

func TestAnnouncement_Update_ReplaceRecipients(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	_, e1 := makeEmpUser(t, "e1@example.com", "E1")
	_, e2 := makeEmpUser(t, "e2@example.com", "E2")
	_, e3 := makeEmpUser(t, "e3@example.com", "E3")
	custom := models.AnnouncementAudienceCustom

	a, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{
		Title:          "x",
		Description:    "y",
		TargetAudience: &custom,
		RecipientIDs:   []uuid.UUID{e1.ID, e2.ID},
	})
	require.NoError(t, err)
	require.Len(t, a.TargetRecipients, 2)

	// Replace with a different set.
	newSet := []uuid.UUID{e2.ID, e3.ID}
	out, err := svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{
		RecipientIDs: &newSet,
	})
	require.NoError(t, err)
	require.Len(t, out.TargetRecipients, 2)
	ids := []uuid.UUID{out.TargetRecipients[0].ID, out.TargetRecipients[1].ID}
	assert.Contains(t, ids, e2.ID)
	assert.Contains(t, ids, e3.ID)
	assert.NotContains(t, ids, e1.ID, "e1 must be soft-deleted from the join")

	// Nil pointer → leave unchanged.
	newTitle := "still-x"
	out2, err := svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{Title: &newTitle})
	require.NoError(t, err)
	require.Len(t, out2.TargetRecipients, 2, "nil RecipientIDs must leave the set unchanged")

	// Empty slice → clear all.
	empty := []uuid.UUID{}
	out3, err := svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{
		RecipientIDs: &empty,
	})
	require.NoError(t, err)
	assert.Len(t, out3.TargetRecipients, 0, "empty slice must clear the set")
}

// ---- Employee repo helpers (notification dispatch) ----

func TestEmployeeRepo_FindAllActive_ReturnsNonDeleted(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	skipIfNoDB(t)
	truncateAll(t)
	repo := repositories.NewEmployeeRepository(testDB)
	_, e1 := makeEmpUser(t, "findall-active-1@test.com", "Find Active One")
	_, e2 := makeEmpUser(t, "findall-active-2@test.com", "Find Active Two")
	// soft-delete e2 directly in DB (bypass service to avoid auth)
	require.NoError(t, testDB.Model(&models.Employee{}).Where("id = ?", e2.ID).Updates(map[string]interface{}{
		"is_deleted": true,
	}).Error)

	emps, err := repo.FindAllActive(context.Background())
	require.NoError(t, err)
	ids := make(map[uuid.UUID]bool)
	for _, e := range emps {
		ids[e.ID] = true
	}
	assert.True(t, ids[e1.ID], "active employee should be returned")
	assert.False(t, ids[e2.ID], "deleted employee should not be returned")
}

func TestEmployeeRepo_FindByIDs_ReturnsMatchingNonDeleted(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	skipIfNoDB(t)
	truncateAll(t)
	repo := repositories.NewEmployeeRepository(testDB)
	_, e1 := makeEmpUser(t, "findbyids-1@test.com", "Find ByIDs One")
	_, e2 := makeEmpUser(t, "findbyids-2@test.com", "Find ByIDs Two")
	_, e3 := makeEmpUser(t, "findbyids-3@test.com", "Find ByIDs Three")
	require.NoError(t, testDB.Model(&models.Employee{}).Where("id = ?", e3.ID).Updates(map[string]interface{}{
		"is_deleted": true,
	}).Error)

	emps, err := repo.FindByIDs(context.Background(), []uuid.UUID{e1.ID, e2.ID, e3.ID})
	require.NoError(t, err)
	ids := make(map[uuid.UUID]bool)
	for _, e := range emps {
		ids[e.ID] = true
	}
	assert.True(t, ids[e1.ID])
	assert.True(t, ids[e2.ID])
	assert.False(t, ids[e3.ID], "soft-deleted should be excluded")
}

func TestEmployeeRepo_FindByIDs_EmptySlice_ReturnsNil(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	skipIfNoDB(t)
	repo := repositories.NewEmployeeRepository(testDB)
	emps, err := repo.FindByIDs(context.Background(), []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, emps)
}

func TestEmployeeRepo_FindByDepartmentIDs_ReturnsMatchingNonDeleted(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	skipIfNoDB(t)
	truncateAll(t)
	repo := repositories.NewEmployeeRepository(testDB)
	dept := makeRawDept(t, "notify-dept-repo")
	_, e1 := makeEmpUserInDept(t, "notifydept-repo-1@test.com", "Notify Dept One", dept.ID)
	_, e2 := makeEmpUserInDept(t, "notifydept-repo-2@test.com", "Notify Dept Two", dept.ID)
	_, e3 := makeEmpUser(t, "notifydept-nodept@test.com", "Notify No Dept")

	emps, err := repo.FindByDepartmentIDs(context.Background(), []uuid.UUID{dept.ID})
	require.NoError(t, err)
	ids := make(map[uuid.UUID]bool)
	for _, e := range emps {
		ids[e.ID] = true
	}
	assert.True(t, ids[e1.ID])
	assert.True(t, ids[e2.ID])
	assert.False(t, ids[e3.ID], "employee in different dept should not appear")
}

// makeAnnouncement inserts an announcement row directly with the given status.
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

func TestAnnouncement_Update_PublishedRow_Returns409(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	_, emp := makeEmpUser(t, "author-upd-pub@test.com", "Author Pub")
	ann := makeAnnouncement(t, emp.ID, models.AnnouncementStatusPublished)

	_, err := svc.Update(ctx, ann.ID, emp.UserID, false, dto.AnnouncementUpdate{})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.HTTP)
}

func TestAnnouncement_Update_ArchivedRow_Returns409(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	_, emp := makeEmpUser(t, "author-upd-arch@test.com", "Author Arch")
	ann := makeAnnouncement(t, emp.ID, models.AnnouncementStatusArchived)

	_, err := svc.Update(ctx, ann.ID, emp.UserID, false, dto.AnnouncementUpdate{})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.HTTP)
}

// ---- Notifier dispatch (G2) ----

func TestAnnouncement_NilNotifier_DoesNotPanic(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	truncateAll(t)
	svc := svcWithNotifier(t, nil)
	_, u := makeEmpUser(t, "nil-notifier-author@test.com", "Nil Author")
	_, err := svc.Create(context.Background(), u.UserID, dto.AnnouncementCreate{
		Title:       "Silent publish",
		Description: "body",
		SendNow:     true,
	})
	require.NoError(t, err)
	// If we reach here without panic, the nil guard works.
}

func TestAnnouncement_Publish_DispatchesNotifier_AudienceAll(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	truncateAll(t)
	notifier := &captureNotifier{}
	svc := svcWithNotifier(t, notifier)
	_, author := makeEmpUser(t, "dispatch-all-author@test.com", "Dispatch All Author")
	_, _ = makeEmpUser(t, "dispatch-all-emp1@test.com", "Emp1")
	_, _ = makeEmpUser(t, "dispatch-all-emp2@test.com", "Emp2")

	_, err := svc.Create(context.Background(), author.UserID, dto.AnnouncementCreate{
		Title:       "All notif",
		Description: "body",
		SendNow:     true,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond) // let goroutine finish
	assert.Equal(t, 1, notifier.callCount(), "notifier should be called once on publish")
	call := notifier.lastCall()
	assert.GreaterOrEqual(t, len(call.UserIDs), 3, "should notify at least the 3 employees we created")
	assert.Equal(t, "All notif", call.Title)
}

func TestAnnouncement_Publish_DispatchesNotifier_AudienceDepartment(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	truncateAll(t)
	notifier := &captureNotifier{}
	svc := svcWithNotifier(t, notifier)
	_, author := makeEmpUser(t, "dispatch-dept-author@test.com", "Dispatch Dept Author")
	dept := makeRawDept(t, "dispatch-dept")
	_, deptEmp := makeEmpUserInDept(t, "dispatch-dept-member@test.com", "Member", dept.ID)
	_, outsideEmp := makeEmpUser(t, "dispatch-dept-outside@test.com", "Outside")

	_, err := svc.Create(context.Background(), author.UserID, dto.AnnouncementCreate{
		Title:          "Dept notif",
		Description:    "body",
		SendNow:        true,
		TargetAudience: ptrAudience(models.AnnouncementAudienceDepartment),
		DepartmentIDs:  []uuid.UUID{dept.ID},
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, 1, notifier.callCount())
	call := notifier.lastCall()
	userIDSet := make(map[uuid.UUID]bool)
	for _, id := range call.UserIDs {
		userIDSet[id] = true
	}
	assert.True(t, userIDSet[deptEmp.UserID], "dept member should be notified")
	assert.False(t, userIDSet[outsideEmp.UserID], "outside member should not be notified")
}

func TestAnnouncement_Publish_DispatchesNotifier_AudienceCustom(t *testing.T) {
	if testDB == nil {
		t.Skip("no test DB")
	}
	truncateAll(t)
	notifier := &captureNotifier{}
	svc := svcWithNotifier(t, notifier)
	_, author := makeEmpUser(t, "dispatch-custom-author@test.com", "Dispatch Custom Author")
	_, recipEmp := makeEmpUser(t, "dispatch-custom-recip@test.com", "Recip")
	_, otherEmp := makeEmpUser(t, "dispatch-custom-other@test.com", "Other")

	_, err := svc.Create(context.Background(), author.UserID, dto.AnnouncementCreate{
		Title:          "Custom notif",
		Description:    "body",
		SendNow:        true,
		TargetAudience: ptrAudience(models.AnnouncementAudienceCustom),
		RecipientIDs:   []uuid.UUID{recipEmp.ID},
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, 1, notifier.callCount())
	call := notifier.lastCall()
	userIDSet := make(map[uuid.UUID]bool)
	for _, id := range call.UserIDs {
		userIDSet[id] = true
	}
	assert.True(t, userIDSet[recipEmp.UserID], "named recipient should be notified")
	assert.False(t, userIDSet[otherEmp.UserID], "other emp should not be notified")
}
