package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

type DeviceTokenRepository struct {
	db *gorm.DB
}

func NewDeviceTokenRepository(db *gorm.DB) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, t *models.DeviceToken) error {
	// Soft-delete any existing live row for (user_id, device_id), then insert new.
	if err := r.db.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("user_id = ? AND device_id = ? AND is_deleted = ?", t.UserID, t.DeviceID, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *DeviceTokenRepository) DeleteByToken(ctx context.Context, userID uuid.UUID, token string) error {
	return r.db.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("user_id = ? AND token = ? AND is_deleted = ?", userID, token, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *DeviceTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Find(&tokens).Error
	return tokens, err
}
