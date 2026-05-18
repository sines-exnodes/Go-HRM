package repositories

import (
	"context"
	"errors"

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
	// The DB has a plain UNIQUE (user_id, device_id) constraint (it is NOT a
	// partial index on is_deleted), so a soft-deleted row still occupies the
	// slot. Update the existing row in place (re-activating it if it had been
	// soft-deleted); insert only when no row exists for the device at all.
	var existing models.DeviceToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", t.UserID, t.DeviceID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(t).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("id = ?", existing.ID).
		Updates(map[string]any{
			"token":      t.Token,
			"platform":   t.Platform,
			"is_deleted": false,
			"deleted_at": nil,
		}).Error
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
