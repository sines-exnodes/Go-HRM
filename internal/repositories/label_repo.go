package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// LabelRepository exposes only the operations needed by the Python source's
// labels module: list all (no pagination, ordered by name) and get-by-name
// for the get-or-create write path. No update/delete by design — these are
// out of scope for Phase 4 (REVISION NOTES item #4).
type LabelRepository interface {
	Create(ctx context.Context, l *models.Label) error
	// FindByName returns (nil, nil) when no live row matches — case-insensitive.
	FindByName(ctx context.Context, name string) (*models.Label, error)
	// List returns every live label, ordered by name ASC. No pagination.
	List(ctx context.Context) ([]models.Label, error)
}

type labelRepository struct{ db *gorm.DB }

func NewLabelRepository(db *gorm.DB) LabelRepository {
	return &labelRepository{db: db}
}

func (r *labelRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *labelRepository) Create(ctx context.Context, l *models.Label) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *labelRepository) FindByName(ctx context.Context, name string) (*models.Label, error) {
	var l models.Label
	err := r.base(ctx).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *labelRepository) List(ctx context.Context) ([]models.Label, error) {
	var items []models.Label
	err := r.base(ctx).
		Order("LOWER(name) ASC").
		Find(&items).Error
	return items, err
}
