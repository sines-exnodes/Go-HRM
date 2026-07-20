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



// ---- Announcement producer ----

// newAnnouncementSvcWithFeed wires a real announcement notifier backed by the
// notification feed. push/email are nil so only the in-app rows are written.
func newAnnouncementSvcWithFeed(t *testing.T, notifs *services.NotificationService) *services.AnnouncementService {
	t.Helper()
	return svcWithNotifier(t, services.NewAnnouncementNotifier(nil, nil, testUserRepo, notifs))
}

// AC-10 — a draft announcement generates nothing. The publish gate is what
// enforces this; this test pins it so a future refactor of broadcastPublished
// cannot silently start notifying on draft save.
func TestNotification_Announcement_DraftGeneratesNothing(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	notifSvc, _ := newNotificationSvc(t)
	annSvc := newAnnouncementSvcWithFeed(t, notifSvc)

	author, _ := makeEmpUser(t, "author-draft@x.com", "Draft Author")
	recipient, _ := makeEmpUser(t, "recipient-draft@x.com", "Draft Recipient")

	_, err := annSvc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:       "Draft only",
		Description: "<p>never sent</p>",
	})
	require.NoError(t, err)

	// Give any stray dispatch goroutine a chance to misbehave before asserting.
	time.Sleep(150 * time.Millisecond)

	out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	assert.Empty(t, out.Items, "a draft announcement must not notify anyone")
}

// AC-10 positive — publishing notifies the target audience.
func TestNotification_Announcement_PublishNotifiesAudience(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	notifSvc, _ := newNotificationSvc(t)
	annSvc := newAnnouncementSvcWithFeed(t, notifSvc)

	author, _ := makeEmpUser(t, "author-pub@x.com", "Pub Author")
	recipient, _ := makeEmpUser(t, "recipient-pub@x.com", "Pub Recipient")

	created, err := annSvc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:       "Office closed Friday",
		Description: "<p>The office is <b>closed</b> this Friday.</p>",
	})
	require.NoError(t, err)

	annID := created.ID
	_, err = annSvc.Publish(ctx, annID, author.ID, true)
	require.NoError(t, err)

	// Dispatch is asynchronous (broadcastPublished launches a goroutine).
	// Eventually beats a fixed sleep: it returns as soon as the row lands and
	// does not fail merely because the machine was busy.
	require.Eventually(t, func() bool {
		out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
		return aerr == nil && len(out.Items) == 1
	}, 3*time.Second, 25*time.Millisecond, "recipient should receive a notification on publish")

	out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "Office closed Friday", out.Items[0].Title)
	assert.Equal(t, string(models.NotificationTypeAnnouncement), out.Items[0].Type)
	assert.Equal(t, annID, out.Items[0].SourceID, "source_id must point at the announcement")
	assert.False(t, out.Items[0].IsRead, "DR Rule 6 — notifications are created unread")
	// Body is stored as plain text, not the raw HTML of the description.
	assert.NotContains(t, out.Items[0].Body, "<b>")
	assert.Contains(t, out.Items[0].Body, "closed")
}

// DR Rule 12 against a REAL published announcement: editing the source row
// afterwards must not rewrite the notification the employee already received.
func TestNotification_Announcement_SnapshotSurvivesRealSourceEdit(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	notifSvc, _ := newNotificationSvc(t)
	annSvc := newAnnouncementSvcWithFeed(t, notifSvc)

	author, _ := makeEmpUser(t, "author-snap@x.com", "Snap Author")
	recipient, _ := makeEmpUser(t, "recipient-snap@x.com", "Snap Recipient")

	created, err := annSvc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:       "Original announcement title",
		Description: "<p>original body</p>",
	})
	require.NoError(t, err)

	annID := created.ID
	_, err = annSvc.Publish(ctx, annID, author.ID, true)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
		return aerr == nil && len(out.Items) == 1
	}, 3*time.Second, 25*time.Millisecond)

	// Edit the real source row.
	require.NoError(t, testDB.Exec(
		`UPDATE announcements SET title = 'Edited announcement title' WHERE id = ?`, annID,
	).Error)

	var liveTitle string
	require.NoError(t, testDB.Raw(`SELECT title FROM announcements WHERE id = ?`, annID).Scan(&liveTitle).Error)
	require.Equal(t, "Edited announcement title", liveTitle, "the source edit must actually have applied")

	out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "Original announcement title", out.Items[0].Title,
		"notification text is a snapshot and must not follow source edits")
}

// DR Rule 13 against a REAL announcement: hard-deleting the source leaves the
// notification listed and markable.
func TestNotification_Announcement_SurvivesRealSourceDeletion(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	notifSvc, _ := newNotificationSvc(t)
	annSvc := newAnnouncementSvcWithFeed(t, notifSvc)

	author, _ := makeEmpUser(t, "author-orphan@x.com", "Orphan Author")
	recipient, _ := makeEmpUser(t, "recipient-orphan@x.com", "Orphan Recipient")

	created, err := annSvc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:       "Soon to be deleted",
		Description: "<p>body</p>",
	})
	require.NoError(t, err)

	annID := created.ID
	_, err = annSvc.Publish(ctx, annID, author.ID, true)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
		return aerr == nil && len(out.Items) == 1
	}, 3*time.Second, 25*time.Millisecond)

	// Hard delete the source. source_id carries no FK, so nothing cascades.
	require.NoError(t, testDB.Exec(`DELETE FROM announcement_views WHERE announcement_id = ?`, annID).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM announcements WHERE id = ?`, annID).Error)

	var remaining int64
	require.NoError(t, testDB.Raw(`SELECT COUNT(*) FROM announcements WHERE id = ?`, annID).Scan(&remaining).Error)
	require.Equal(t, int64(0), remaining, "the source delete must actually have applied")

	out, aerr := notifSvc.List(ctx, recipient.ID, dto.NotificationListQuery{})
	require.Nil(t, aerr)
	require.Len(t, out.Items, 1, "deleting the source must not remove the notification")
	assert.Equal(t, "Soon to be deleted", out.Items[0].Title)

	_, aerr = notifSvc.MarkRead(ctx, out.Items[0].ID, recipient.ID)
	require.Nil(t, aerr, "an orphaned notification must still be markable")
}
