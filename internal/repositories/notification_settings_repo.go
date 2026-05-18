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
		return r.db.WithContext(ctx).Create(&models.UserNotificationSettings{
			UserID:               userID,
			NotificationsEnabled: enabled,
		}).Error
	}
	return r.db.WithContext(ctx).Model(&models.UserNotificationSettings{}).
		Where("id = ?", existing.ID).
		Update("notifications_enabled", enabled).Error
}
