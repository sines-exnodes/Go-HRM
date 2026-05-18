package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// DependentService manages an employee's dependents with owner-or-admin
// authorization. Both repositories are consumed via their interfaces for
// mockability.
type DependentService struct {
	deps repositories.DependentRepository
	emps repositories.EmployeeRepository
}

func NewDependentService(deps repositories.DependentRepository, emps repositories.EmployeeRepository) *DependentService {
	return &DependentService{deps: deps, emps: emps}
}

// AuthorizeOwnerOrAdmin returns nil if caller may manage employeeID's dependents.
// canManageAll == true when the caller has PermDependentsManage; otherwise we
// require employees.user_id == callerUserID.
func (s *DependentService) AuthorizeOwnerOrAdmin(ctx context.Context, callerUserID uuid.UUID, employeeID uuid.UUID, canManageAll bool) error {
	if canManageAll {
		return nil
	}
	e, err := s.emps.FindByIDWithFull(ctx, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Employee")
		}
		return err
	}
	if e.UserID != callerUserID {
		return apperrors.ErrForbidden("You may only manage your own dependents")
	}
	return nil
}

func (s *DependentService) List(ctx context.Context, employeeID uuid.UUID, q dto.DependentListQuery) ([]dto.DependentRead, int64, error) {
	rows, total, err := s.deps.ListByEmployee(ctx, employeeID, q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.DependentRead, 0, len(rows))
	for i := range rows {
		out = append(out, toDependentRead(&rows[i]))
	}
	return out, total, nil
}

func (s *DependentService) Create(ctx context.Context, employeeID uuid.UUID, in dto.DependentCreate) (*dto.DependentRead, error) {
	if _, err := s.emps.FindByIDWithFull(ctx, employeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	d := &models.Dependent{
		EmployeeID:   employeeID,
		FullName:     in.FullName,
		DOB:          in.DOB,
		Gender:       in.Gender,
		Relationship: in.Relationship,
	}
	if err := s.deps.Create(ctx, d); err != nil {
		return nil, err
	}
	view := toDependentRead(d)
	return &view, nil
}

func (s *DependentService) Update(ctx context.Context, id uuid.UUID, in dto.DependentUpdate) (*dto.DependentRead, error) {
	existing, err := s.deps.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Dependent")
		}
		return nil, err
	}
	fields := map[string]any{}
	if in.FullName != nil {
		fields["full_name"] = *in.FullName
	}
	if in.DOB != nil {
		fields["dob"] = *in.DOB
	}
	if in.Gender != nil {
		fields["gender"] = *in.Gender
	}
	if in.Relationship != nil {
		fields["relationship"] = *in.Relationship
	}
	if err := s.deps.Update(ctx, existing.ID, fields); err != nil {
		return nil, err
	}
	fresh, err := s.deps.FindByID(ctx, existing.ID)
	if err != nil {
		return nil, err
	}
	view := toDependentRead(fresh)
	return &view, nil
}

func (s *DependentService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.deps.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Dependent")
		}
		return err
	}
	return s.deps.SoftDelete(ctx, id)
}

// OwnerEmployeeIDForDependent returns the employee_id that owns the given dependent,
// for ownership checks on /dependents/:id update/delete routes.
func (s *DependentService) OwnerEmployeeIDForDependent(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	d, err := s.deps.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.ErrNotFound("Dependent")
		}
		return uuid.Nil, err
	}
	return d.EmployeeID, nil
}

func toDependentRead(d *models.Dependent) dto.DependentRead {
	return dto.DependentRead{
		ID:           d.ID,
		EmployeeID:   d.EmployeeID,
		FullName:     d.FullName,
		DOB:          d.DOB,
		Gender:       d.Gender,
		Relationship: d.Relationship,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
