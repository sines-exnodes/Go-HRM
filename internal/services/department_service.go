package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// DepartmentService owns department business logic. Positions are now a
// flat global catalog (see migration 000014), so the service no longer
// holds a position-repo dependency and Delete no longer guards against
// positions assigned to the department.
type DepartmentService struct {
	repo repositories.DepartmentRepository
}

func NewDepartmentService(repo repositories.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func departmentToRead(d *models.Department, employeeCount int64) dto.DepartmentRead {
	out := dto.DepartmentRead{
		ID:            d.ID,
		Name:          d.Name,
		Description:   d.Description,
		ParentID:      d.ParentID,
		EmployeeCount: employeeCount,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	if d.Parent != nil {
		// Parent's own employee count is not hydrated here; the parent is
		// shown as a denormalised reference, not a fully-formed record.
		p := departmentToRead(d.Parent, 0)
		out.Parent = &p
	}
	return out
}

func (s *DepartmentService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	if excludeID != nil && existing.ID == *excludeID {
		return nil
	}
	return apperrors.ErrConflict("Department name already exists")
}

// assertParent verifies the proposed parent exists and that setting it
// would not create a cycle (only relevant when updating an existing node).
func (s *DepartmentService) assertParent(ctx context.Context, parentID uuid.UUID, selfID *uuid.UUID) error {
	if selfID != nil && parentID == *selfID {
		return apperrors.ErrBadRequest("Department cannot be its own parent")
	}
	parent, err := s.repo.FindByID(ctx, parentID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrBadRequest("Parent department not found")
		}
		return err
	}
	if selfID != nil {
		current := parent
		for current.ParentID != nil {
			if *current.ParentID == *selfID {
				return apperrors.ErrBadRequest("Setting this parent would create a cycle")
			}
			next, err := s.repo.FindByID(ctx, *current.ParentID, false)
			if err != nil {
				return err
			}
			current = next
		}
	}
	return nil
}

func (s *DepartmentService) Create(ctx context.Context, in dto.DepartmentCreate) (*dto.DepartmentRead, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperrors.ErrBadRequest("Department name cannot be blank")
	}
	if err := s.checkNameUnique(ctx, name, nil); err != nil {
		return nil, err
	}
	if in.ParentID != nil {
		if err := s.assertParent(ctx, *in.ParentID, nil); err != nil {
			return nil, err
		}
	}
	d := &models.Department{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		ParentID:    in.ParentID,
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	fresh, err := s.repo.FindByID(ctx, d.ID, true)
	if err != nil {
		return nil, err
	}
	out := departmentToRead(fresh, 0) // freshly created → zero employees
	return &out, nil
}

func (s *DepartmentService) Update(ctx context.Context, id uuid.UUID, in dto.DepartmentUpdate) (*dto.DepartmentRead, error) {
	d, err := s.repo.FindByID(ctx, id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Department")
		}
		return nil, err
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperrors.ErrBadRequest("Department name cannot be blank")
		}
		if err := s.checkNameUnique(ctx, name, &d.ID); err != nil {
			return nil, err
		}
		d.Name = name
	}
	if in.Description != nil {
		d.Description = strings.TrimSpace(*in.Description)
	}
	switch {
	case in.ClearParent:
		d.ParentID = nil
	case in.ParentID != nil:
		if err := s.assertParent(ctx, *in.ParentID, &d.ID); err != nil {
			return nil, err
		}
		d.ParentID = in.ParentID
	}
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	fresh, err := s.repo.FindByID(ctx, d.ID, true)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountEmployees(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	out := departmentToRead(fresh, count)
	return &out, nil
}

// Delete soft-deletes the department after verifying it has no child
// departments and no assigned employees. The pre-000014 positions blocker
// is gone — positions are a flat global catalog now, not bound to any
// department.
func (s *DepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Department")
		}
		return err
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return apperrors.ErrConflict("Cannot delete department — it has child departments. Move or delete them first.")
	}

	empCount, err := s.repo.CountEmployees(ctx, id)
	if err != nil {
		return err
	}
	if empCount > 0 {
		word := "employee is"
		if empCount > 1 {
			word = "employees are"
		}
		return apperrors.ErrConflict(fmt.Sprintf(
			"Cannot delete — %d %s assigned to this department. Reassign all employees before deleting.", empCount, word))
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *DepartmentService) Get(ctx context.Context, id uuid.UUID) (*dto.DepartmentRead, error) {
	d, err := s.repo.FindByID(ctx, id, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Department")
		}
		return nil, err
	}
	count, err := s.repo.CountEmployees(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	out := departmentToRead(d, count)
	return &out, nil
}

func (s *DepartmentService) List(ctx context.Context, q dto.DepartmentListQuery) (*dto.PaginatedData[dto.DepartmentRead], error) {
	f := repositories.DepartmentFilter{Page: q.Page, PageSize: q.PageSize, Search: q.Search}
	switch strings.ToLower(strings.TrimSpace(q.ParentID)) {
	case "":
		// no parent filter
	case "root", "null":
		nilUUID := uuid.Nil
		f.ParentID = &nilUUID
	default:
		parsed, err := uuid.Parse(q.ParentID)
		if err != nil {
			return nil, apperrors.ErrBadRequest("Invalid parent_id")
		}
		f.ParentID = &parsed
	}

	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	counts, err := s.repo.CountEmployeesByDepartmentIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	reads := make([]dto.DepartmentRead, 0, len(items))
	for i := range items {
		reads = append(reads, departmentToRead(&items[i], counts[items[i].ID]))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return &dto.PaginatedData[dto.DepartmentRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
