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

// SkillFilter mirrors dto.SkillListQuery in a service-agnostic shape.
type SkillFilter struct {
	Page     int
	PageSize int
	Search   string
}

type SkillRepository interface {
	Create(ctx context.Context, s *models.Skill) error
	Update(ctx context.Context, s *models.Skill) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Skill, error)
	// FindByName returns (nil, nil) when no live row matches — case-insensitive.
	FindByName(ctx context.Context, name string) (*models.Skill, error)
	List(ctx context.Context, f SkillFilter) ([]models.Skill, int64, error)
}

type skillRepository struct{ db *gorm.DB }

func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *skillRepository) Create(ctx context.Context, s *models.Skill) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *skillRepository) Update(ctx context.Context, s *models.Skill) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *skillRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Skill{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *skillRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Skill, error) {
	var s models.Skill
	if err := r.base(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepository) FindByName(ctx context.Context, name string) (*models.Skill, error) {
	var s models.Skill
	err := r.base(ctx).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *skillRepository) List(ctx context.Context, f SkillFilter) ([]models.Skill, int64, error) {
	q := r.base(ctx).Model(&models.Skill{})
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
		size = 20
	}
	var items []models.Skill
	err := q.
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}
