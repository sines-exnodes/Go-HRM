package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

type LeaveQuotaRepository struct {
	db *gorm.DB
}

func NewLeaveQuotaRepository(db *gorm.DB) *LeaveQuotaRepository {
	return &LeaveQuotaRepository{db: db}
}

func (r *LeaveQuotaRepository) GetByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeLeaveQuota, error) {
	var q models.EmployeeLeaveQuota
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND is_deleted = ?", employeeID, false).
		First(&q).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &q, err
}

func (r *LeaveQuotaRepository) Upsert(ctx context.Context, employeeID uuid.UUID, annual, sick float64) error {
	existing, err := r.GetByEmployee(ctx, employeeID)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.db.WithContext(ctx).Create(&models.EmployeeLeaveQuota{
			EmployeeID:       employeeID,
			AnnualLeaveQuota: annual,
			SickLeaveQuota:   sick,
		}).Error
	}
	return r.db.WithContext(ctx).Model(&models.EmployeeLeaveQuota{}).
		Where("id = ?", existing.ID).
		Updates(map[string]any{"annual_leave_quota": annual, "sick_leave_quota": sick}).Error
}
