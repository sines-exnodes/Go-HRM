package services_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
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
	)
	return svc, hub
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
	pub := models.AnnouncementStatusPublished
	a, err := svc.Create(ctx, admin.ID, dto.AnnouncementCreate{Title: "x", Description: "y", Status: &pub})
	require.NoError(t, err)
	require.Equal(t, 1, hub.Count())

	newTitle := "x revised"
	_, err = svc.Update(ctx, a.ID, admin.ID, true, dto.AnnouncementUpdate{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, 1, hub.Count(), "edit of already-published must NOT rebroadcast")
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
