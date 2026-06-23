package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// PasswordResetTokenRepository handles persistence for password reset tokens.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, t *models.PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	// InvalidatePendingForUser soft-deletes all unused, unexpired tokens for
	// the given user so at most one active reset is in flight at a time.
	InvalidatePendingForUser(ctx context.Context, userID uuid.UUID) error
	MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error
	UpdateEmailError(ctx context.Context, id uuid.UUID, errMsg string) error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository constructs the GORM-backed implementation.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, t *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *passwordResetTokenRepository) FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND is_deleted = FALSE", token).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *passwordResetTokenRepository) InvalidatePendingForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ? AND is_deleted = FALSE", userID, now).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *passwordResetTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("id = ?", id).
		Update("used_at", usedAt).Error
}

func (r *passwordResetTokenRepository) UpdateEmailError(ctx context.Context, id uuid.UUID, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("id = ?", id).
		Update("last_email_error", errMsg).Error
}
