package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

type NotificationSettingsRepository struct {
	db *gorm.DB
}

func NewNotificationSettingsRepository(db *gorm.DB) *NotificationSettingsRepository {
	return &NotificationSettingsRepository{db: db}
}

func (r *NotificationSettingsRepository) Get(ctx context.Context, userID uuid.UUID) (*models.UserNotificationSettings, error) {
	var s models.UserNotificationSettings
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *NotificationSettingsRepository) Upsert(ctx context.Context, userID uuid.UUID, enabled bool) error {
	existing, err := r.Get(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		// Insert via an explicit column map so a zero value (false) is always
		// written. A struct Create would let the column's `default:true`
		// override an intended "disable notifications" on first write, because
		// GORM omits zero-value fields that have a default tag.
		return r.db.WithContext(ctx).
			Model(&models.UserNotificationSettings{}).
			Create(map[string]any{
				"user_id":               userID,
				"notifications_enabled": enabled,
			}).Error
	}
	return r.db.WithContext(ctx).Model(&models.UserNotificationSettings{}).
		Where("id = ?", existing.ID).
		Update("notifications_enabled", enabled).Error
}
