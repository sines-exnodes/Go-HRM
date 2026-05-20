package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

type EmployeeSkillRepository interface {
	// ListByEmployee returns live employee_skills rows for the given employee
	// with the Skill side preloaded (live skills only).
	ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.EmployeeSkill, error)
	// CountEmployeesBySkill counts how many distinct live employees are
	// currently assigned to a given skill. Used by SkillService.Delete to
	// produce the 409 conflict body.
	CountEmployeesBySkill(ctx context.Context, skillID uuid.UUID) (int64, error)
	// ReplaceForEmployee atomically replaces the live skill set assigned to
	// an employee. Rows for IDs not in the new set are soft-deleted; rows
	// for IDs in the new set are inserted if missing, or re-activated if
	// previously soft-deleted (matches the partial unique index).
	ReplaceForEmployee(ctx context.Context, employeeID uuid.UUID, skillIDs []uuid.UUID) error
}

type employeeSkillRepository struct{ db *gorm.DB }

func NewEmployeeSkillRepository(db *gorm.DB) EmployeeSkillRepository {
	return &employeeSkillRepository{db: db}
}

func (r *employeeSkillRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *employeeSkillRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.EmployeeSkill, error) {
	var rows []models.EmployeeSkill
	err := r.base(ctx).
		Preload("Skill", models.NotDeleted).
		Where("employee_id = ?", employeeID).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *employeeSkillRepository) CountEmployeesBySkill(ctx context.Context, skillID uuid.UUID) (int64, error) {
	var count int64
	err := r.base(ctx).
		Model(&models.EmployeeSkill{}).
		Where("skill_id = ?", skillID).
		Distinct("employee_id").
		Count(&count).Error
	return count, err
}

func (r *employeeSkillRepository) ReplaceForEmployee(ctx context.Context, employeeID uuid.UUID, skillIDs []uuid.UUID) error {
	// De-duplicate the request set first to keep the diff arithmetic clean.
	want := make(map[uuid.UUID]struct{}, len(skillIDs))
	for _, id := range skillIDs {
		if id == uuid.Nil {
			continue
		}
		want[id] = struct{}{}
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Snapshot the current state — include soft-deleted rows so we can
		// reactivate them rather than insert and trip the unique index.
		var existing []models.EmployeeSkill
		if err := tx.Where("employee_id = ?", employeeID).Find(&existing).Error; err != nil {
			return err
		}
		have := make(map[uuid.UUID]models.EmployeeSkill, len(existing))
		for _, row := range existing {
			have[row.SkillID] = row
		}

		// 1) Soft-delete rows for skill_ids that should no longer be assigned.
		var toRemove []uuid.UUID
		for sid, row := range have {
			if row.IsDeleted {
				continue
			}
			if _, keep := want[sid]; !keep {
				toRemove = append(toRemove, sid)
			}
		}
		if len(toRemove) > 0 {
			if err := tx.Model(&models.EmployeeSkill{}).
				Where("employee_id = ? AND skill_id IN ?", employeeID, toRemove).
				Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error; err != nil {
				return err
			}
		}

		// 2) For each desired skill: insert if missing; reactivate if soft-deleted.
		for sid := range want {
			row, ok := have[sid]
			switch {
			case !ok:
				// Insert fresh row.
				newRow := models.EmployeeSkill{EmployeeID: employeeID, SkillID: sid}
				if err := tx.Create(&newRow).Error; err != nil {
					return err
				}
			case row.IsDeleted:
				// Reactivate.
				if err := tx.Model(&models.EmployeeSkill{}).
					Where("id = ?", row.ID).
					Updates(map[string]any{"is_deleted": false, "deleted_at": gorm.Expr("NULL")}).Error; err != nil {
					return err
				}
			}
			// else: already live and desired — no-op.
		}
		return nil
	})
}
