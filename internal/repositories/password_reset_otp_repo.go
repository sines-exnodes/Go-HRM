package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/exnodes/hrm-api/internal/models"
)

// PasswordResetOTPRepository handles persistence for mobile password-reset
// OTP codes.
type PasswordResetOTPRepository interface {
	Create(ctx context.Context, o *models.PasswordResetOTP) error
	// FindLatestActiveForUser returns the most recent unconsumed, un-deleted
	// code for the user. Expiry is checked by the caller so it can return the
	// "code expired" message rather than "no code found".
	FindLatestActiveForUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetOTP, error)
	// FindLatestForUser returns the most recent row for the user regardless of
	// consumed/deleted state. Used to enforce the resend cooldown, which must
	// apply even after a code has been superseded or burned.
	FindLatestForUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetOTP, error)
	// CountCreatedSince counts every row created for the user since `since`,
	// INCLUDING soft-deleted and consumed ones. Superseding a code must not
	// hand the caller a fresh rate-limit budget (SR-04).
	CountCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
	// InvalidatePendingForUser soft-deletes all unconsumed codes for the user
	// so only the newest code is ever valid (SR-03).
	InvalidatePendingForUser(ctx context.Context, userID uuid.UUID) error
	MarkConsumed(ctx context.Context, id uuid.UUID, consumedAt time.Time) error
	// IncrementAttempts bumps attempt_count and returns the stored value.
	IncrementAttempts(ctx context.Context, id uuid.UUID) (int, error)
	// Burn soft-deletes a single code (used when the attempt limit is hit).
	Burn(ctx context.Context, id uuid.UUID) error
	UpdateEmailError(ctx context.Context, id uuid.UUID, errMsg string) error
}

type passwordResetOTPRepository struct {
	db *gorm.DB
}

// NewPasswordResetOTPRepository constructs the GORM-backed implementation.
func NewPasswordResetOTPRepository(db *gorm.DB) PasswordResetOTPRepository {
	return &passwordResetOTPRepository{db: db}
}

func (r *passwordResetOTPRepository) Create(ctx context.Context, o *models.PasswordResetOTP) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *passwordResetOTPRepository) FindLatestActiveForUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetOTP, error) {
	var o models.PasswordResetOTP
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND consumed_at IS NULL AND is_deleted = FALSE", userID).
		Order("created_at DESC").
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *passwordResetOTPRepository) FindLatestForUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetOTP, error) {
	var o models.PasswordResetOTP
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *passwordResetOTPRepository) CountCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.PasswordResetOTP{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&n).Error
	return n, err
}

func (r *passwordResetOTPRepository) InvalidatePendingForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetOTP{}).
		Where("user_id = ? AND consumed_at IS NULL AND is_deleted = FALSE", userID).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *passwordResetOTPRepository) MarkConsumed(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetOTP{}).
		Where("id = ?", id).
		Update("consumed_at", consumedAt).Error
}

func (r *passwordResetOTPRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) (int, error) {
	var updated models.PasswordResetOTP
	err := r.db.WithContext(ctx).
		Model(&updated).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "attempt_count"}}}).
		Where("id = ?", id).
		Update("attempt_count", gorm.Expr("attempt_count + 1")).Error
	if err != nil {
		return 0, err
	}
	return updated.AttemptCount, nil
}

func (r *passwordResetOTPRepository) Burn(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetOTP{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *passwordResetOTPRepository) UpdateEmailError(ctx context.Context, id uuid.UUID, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetOTP{}).
		Where("id = ?", id).
		Update("last_email_error", errMsg).Error
}
