package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// SystemConfigRepository exposes only the operations needed for the
// singleton system_config row. NotDeleted scope is intentionally NOT
// applied (REVISION NOTES #10) — the row has no soft-delete path.
type SystemConfigRepository interface {
	// Get returns the singleton row, or (nil, nil) when no row exists
	// (only possible if seed hasn't run yet).
	Get(ctx context.Context) (*models.SystemConfig, error)
	// EnsureExists performs an idempotent INSERT ... ON CONFLICT DO
	// NOTHING of the sentinel row. Safe to call repeatedly.
	EnsureExists(ctx context.Context) error
	// UpdateFields applies a partial update to the singleton row. The
	// caller is responsible for setting only the columns it wants to
	// change; nil-pointer fields are not written.
	UpdateFields(ctx context.Context, fields map[string]any) error
}

type systemConfigRepository struct{ db *gorm.DB }

// NewSystemConfigRepository constructs a Postgres-backed
// SystemConfigRepository.
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

func (r *systemConfigRepository) Get(ctx context.Context) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	err := r.db.WithContext(ctx).
		Where("id = ?", models.SystemConfigSingletonID).
		First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *systemConfigRepository) EnsureExists(ctx context.Context) error {
	// Raw SQL avoids GORM trying to populate fields on a struct insert
	// and lets us rely on the column defaults.
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO system_config (id) VALUES (?) ON CONFLICT (id) DO NOTHING`,
		models.SystemConfigSingletonID,
	).Error
}

func (r *systemConfigRepository) UpdateFields(ctx context.Context, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&models.SystemConfig{}).
		Where("id = ?", models.SystemConfigSingletonID).
		Updates(fields).Error
}
