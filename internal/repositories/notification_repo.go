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
