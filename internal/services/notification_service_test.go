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

	svc, _ := newNotificationSvc(t)
	u := makeUser(t, "dedupe@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	require.NoError(t, svc.CreateMany(context.Background(), []models.Notification{{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "same event",
		Body:     "body",
		SourceID: sourceID,
	}}))
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

// DR Rule 12 — notification content is a snapshot. Editing the source after
// the fact must not rewrite what the employee was already told. This test is
// the reason the table stores title/body instead of joining to the source.
func TestNotification_SnapshotSurvivesSourceEdit(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc, _ := newNotificationSvc(t)
	u := makeUser(t, "snapshot@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	require.NoError(t, svc.CreateMany(context.Background(), []models.Notification{{
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

	out, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
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

	svc, _ := newNotificationSvc(t)
	u := makeUser(t, "orphan@example.com", "pw-Aa123456")
	sourceID := uuid.New()

	require.NoError(t, svc.CreateMany(context.Background(), []models.Notification{{
		UserID:   u.ID,
		Type:     models.NotificationTypeAnnouncement,
		Title:    "Orphaned soon",
		Body:     "body",
		SourceID: sourceID,
	}}))

	// There is no FK on source_id, so deleting the source must not cascade.
	require.NoError(t, testDB.Exec(`DELETE FROM announcements WHERE id = ?`, sourceID).Error)

	out, aerr := svc.List(context.Background(), u.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1, "deleting the source must not remove the notification")
	assert.Equal(t, "Orphaned soon", out.Items[0].Title)

	// And it is still markable.
	_, aerr = svc.MarkRead(context.Background(), out.Items[0].ID, u.ID)
	require.Nil(t, aerr)
}
