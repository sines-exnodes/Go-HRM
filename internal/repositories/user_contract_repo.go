package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// UserContractRepository defines data-access operations for the user_contracts table.
type UserContractRepository interface {
	List(ctx context.Context, employeeID uuid.UUID, q dto.UserContractListQuery) ([]models.UserContract, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*models.UserContract, error)
	Create(ctx context.Context, c *models.UserContract) error
	Update(ctx context.Context, c *models.UserContract) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userContractRepository struct{ db *gorm.DB }

// NewUserContractRepository returns a UserContractRepository backed by GORM.
func NewUserContractRepository(db *gorm.DB) UserContractRepository {
	return &userContractRepository{db: db}
}

func (r *userContractRepository) Create(ctx context.Context, c *models.UserContract) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *userContractRepository) Get(ctx context.Context, id uuid.UUID) (*models.UserContract, error) {
	var c models.UserContract
	if err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = false", id).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *userContractRepository) Update(ctx context.Context, c *models.UserContract) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *userContractRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.UserContract{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *userContractRepository) List(ctx context.Context, employeeID uuid.UUID, q dto.UserContractListQuery) ([]models.UserContract, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.UserContract{}).
		Where("employee_id = ? AND is_deleted = false", employeeID)

	if q.SignedFrom != nil {
		db = db.Where("signed_date >= ?", *q.SignedFrom)
	}
	if q.SignedTo != nil {
		db = db.Where("signed_date <= ?", *q.SignedTo)
	}
	if q.ExpiryFrom != nil {
		db = db.Where("expiry_date >= ?", *q.ExpiryFrom)
	}
	if q.ExpiryTo != nil {
		db = db.Where("expiry_date <= ?", *q.ExpiryTo)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var contracts []models.UserContract
	err := db.Order("signed_date DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&contracts).Error
	return contracts, total, err
}
