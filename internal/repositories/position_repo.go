package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

type PositionFilter struct {
	Page     int
	PageSize int
	Search   string
}

type PositionRepository interface {
	Create(ctx context.Context, p *models.Position) error
	Update(ctx context.Context, p *models.Position) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Position, error)
	FindByName(ctx context.Context, name string) (*models.Position, error)
	List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error)
	// CountEmployees counts non-deleted employees whose position_id == id.
	CountEmployees(ctx context.Context, id uuid.UUID) (int64, error)
	// CountEmployeesByPositionIDs returns a map of position id → employee
	// count. Positions with zero employees are absent from the map so
	// callers must default-zero on lookup. Used by PositionService.List for
	// batch hydration without N+1 queries.
	CountEmployeesByPositionIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error)
}

type positionRepository struct{ db *gorm.DB }

func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *positionRepository) Create(ctx context.Context, p *models.Position) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *positionRepository) Update(ctx context.Context, p *models.Position) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *positionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Position{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *positionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Position, error) {
	var p models.Position
	if err := r.base(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// FindByName returns (nil, nil) when no active row matches — callers treat
// that as "name available".
func (r *positionRepository) FindByName(ctx context.Context, name string) (*models.Position, error) {
	var p models.Position
	err := r.base(ctx).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *positionRepository) List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error) {
	q := r.base(ctx).Model(&models.Position{})
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 10
	}
	var items []models.Position
	err := q.
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *positionRepository) CountEmployees(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Employee{}).
		Where("position_id = ? AND is_deleted = ?", id, false).
		Count(&count).Error
	return count, err
}

func (r *positionRepository) CountEmployeesByPositionIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	out := make(map[uuid.UUID]int64, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	type row struct {
		PositionID uuid.UUID `gorm:"column:position_id"`
		Count      int64     `gorm:"column:count"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&models.Employee{}).
		Select("position_id, COUNT(*) AS count").
		Where("position_id IN ? AND is_deleted = ?", ids, false).
		Group("position_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.PositionID] = row.Count
	}
	return out, nil
}
