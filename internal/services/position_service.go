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

type PositionService struct {
	repo     repositories.PositionRepository
	deptRepo repositories.DepartmentRepository
}

func NewPositionService(repo repositories.PositionRepository, deptRepo repositories.DepartmentRepository) *PositionService {
	return &PositionService{repo: repo, deptRepo: deptRepo}
}

func positionToRead(p *models.Position) dto.PositionRead {
	out := dto.PositionRead{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		DepartmentID: p.DepartmentID,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.Department != nil {
		d := departmentToRead(p.Department)
		out.Department = &d
	}
	return out
}

func (s *PositionService) assertDept(ctx context.Context, deptID uuid.UUID) error {
	if _, err := s.deptRepo.FindByID(ctx, deptID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrBadRequest("Department not found")
		}
		return err
	}
	return nil
}

func (s *PositionService) checkNameUniqueInDept(ctx context.Context, name string, deptID uuid.UUID, excludeID *uuid.UUID) error {
	existing, err := s.repo.FindByNameInDept(ctx, name, deptID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	if excludeID != nil && existing.ID == *excludeID {
		return nil
	}
	return apperrors.ErrConflict("Position name already exists in this department")
}

func (s *PositionService) Create(ctx context.Context, in dto.PositionCreate) (*dto.PositionRead, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperrors.ErrBadRequest("Position name cannot be blank")
	}
	if err := s.assertDept(ctx, in.DepartmentID); err != nil {
		return nil, err
	}
	if err := s.checkNameUniqueInDept(ctx, name, in.DepartmentID, nil); err != nil {
		return nil, err
	}
	p := &models.Position{
		Name:         name,
		Description:  strings.TrimSpace(in.Description),
		DepartmentID: in.DepartmentID,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.repo.FindByID(ctx, p.ID, true)
	if err != nil {
		return nil, err
	}
	out := positionToRead(fresh)
	return &out, nil
}

func (s *PositionService) Update(ctx context.Context, id uuid.UUID, in dto.PositionUpdate) (*dto.PositionRead, error) {
	p, err := s.repo.FindByID(ctx, id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Position")
		}
		return nil, err
	}
	newDept := p.DepartmentID
	if in.DepartmentID != nil {
		if err := s.assertDept(ctx, *in.DepartmentID); err != nil {
			return nil, err
		}
		newDept = *in.DepartmentID
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperrors.ErrBadRequest("Position name cannot be blank")
		}
		if err := s.checkNameUniqueInDept(ctx, name, newDept, &p.ID); err != nil {
			return nil, err
		}
		p.Name = name
	}
	if in.Description != nil {
		p.Description = strings.TrimSpace(*in.Description)
	}
	p.DepartmentID = newDept
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.repo.FindByID(ctx, p.ID, true)
	if err != nil {
		return nil, err
	}
	out := positionToRead(fresh)
	return &out, nil
}

func (s *PositionService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Position")
		}
		return err
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
			"Cannot delete — %d %s assigned to this position. Reassign all employees before deleting.", empCount, word))
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *PositionService) Get(ctx context.Context, id uuid.UUID) (*dto.PositionRead, error) {
	p, err := s.repo.FindByID(ctx, id, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Position")
		}
		return nil, err
	}
	out := positionToRead(p)
	return &out, nil
}

func (s *PositionService) List(ctx context.Context, q dto.PositionListQuery) (*dto.PaginatedData[dto.PositionRead], error) {
	items, total, err := s.repo.List(ctx, repositories.PositionFilter{
		Page:         q.Page,
		PageSize:     q.PageSize,
		Search:       q.Search,
		DepartmentID: q.DepartmentID,
	})
	if err != nil {
		return nil, err
	}
	reads := make([]dto.PositionRead, 0, len(items))
	for i := range items {
		reads = append(reads, positionToRead(&items[i]))
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
	return &dto.PaginatedData[dto.PositionRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
