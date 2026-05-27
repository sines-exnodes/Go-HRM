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
	repo repositories.PositionRepository
}

func NewPositionService(repo repositories.PositionRepository) *PositionService {
	return &PositionService{repo: repo}
}

func positionToRead(p *models.Position, employeeCount int64) dto.PositionRead {
	return dto.PositionRead{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		EmployeeCount: employeeCount,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// checkNameUnique enforces the global case-insensitive uniqueness of
// position names (post-migration 000014). excludeID, when non-nil, allows
// the owning row to keep its current name.
func (s *PositionService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
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
	return apperrors.ErrConflict("Position name already exists")
}

func (s *PositionService) Create(ctx context.Context, in dto.PositionCreate) (*dto.PositionRead, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperrors.ErrBadRequest("Position name cannot be blank")
	}
	if err := s.checkNameUnique(ctx, name, nil); err != nil {
		return nil, err
	}
	p := &models.Position{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	out := positionToRead(p, 0) // freshly created → zero employees
	return &out, nil
}

func (s *PositionService) Update(ctx context.Context, id uuid.UUID, in dto.PositionUpdate) (*dto.PositionRead, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Position")
		}
		return nil, err
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperrors.ErrBadRequest("Position name cannot be blank")
		}
		if err := s.checkNameUnique(ctx, name, &p.ID); err != nil {
			return nil, err
		}
		p.Name = name
	}
	if in.Description != nil {
		p.Description = strings.TrimSpace(*in.Description)
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	count, err := s.repo.CountEmployees(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	out := positionToRead(p, count)
	return &out, nil
}

func (s *PositionService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
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
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Position")
		}
		return nil, err
	}
	count, err := s.repo.CountEmployees(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	out := positionToRead(p, count)
	return &out, nil
}

func (s *PositionService) List(ctx context.Context, q dto.PositionListQuery) (*dto.PaginatedData[dto.PositionRead], error) {
	items, total, err := s.repo.List(ctx, repositories.PositionFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	counts, err := s.repo.CountEmployeesByPositionIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	reads := make([]dto.PositionRead, 0, len(items))
	for i := range items {
		reads = append(reads, positionToRead(&items[i], counts[items[i].ID]))
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
