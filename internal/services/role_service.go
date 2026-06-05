package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// roleNamePattern mirrors the Python source: letters, digits, spaces, hyphens,
// ampersands. The trailing '-' in the class is a literal hyphen.
var roleNamePattern = regexp.MustCompile(`^[a-zA-Z0-9 &-]+$`)

// RoleService owns role-management business logic.
type RoleService struct {
	repo repositories.RoleRepository
}

func NewRoleService(repo repositories.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func roleToRead(r *models.Role) dto.RoleRead {
	perms := make([]string, 0, len(r.Permissions))
	perms = append(perms, r.Permissions...)
	return dto.RoleRead{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description,
		Level:           r.Level,
		Permissions:     perms,
		PermissionCount: len(perms),
		IsSystem:        r.IsSystem,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func validateRoleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apperrors.ErrBadRequest("Role name is required")
	}
	if len(name) > 100 {
		return "", apperrors.ErrBadRequest("Role name must not exceed 100 characters")
	}
	if !roleNamePattern.MatchString(name) {
		return "", apperrors.ErrBadRequest("Role name can only contain letters, numbers, spaces, hyphens, and ampersands")
	}
	return name, nil
}

// validatePermissions rejects unknown permission strings. The wildcard '*' is
// allowed (permissions.IsValid returns true for it).
func validatePermissions(perms []string) (models.StringSlice, error) {
	out := make(models.StringSlice, 0, len(perms))
	var invalid []string
	for _, p := range perms {
		if !permissions.IsValid(permissions.Permission(p)) {
			invalid = append(invalid, p)
			continue
		}
		out = append(out, p)
	}
	if len(invalid) > 0 {
		return nil, apperrors.ErrBadRequest("Unknown permissions: " + strings.Join(invalid, ", "))
	}
	return out, nil
}

func (s *RoleService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if excludeID != nil && existing.ID == *excludeID {
		return nil
	}
	return apperrors.ErrConflict("Role name already exists")
}

func (s *RoleService) Create(ctx context.Context, in dto.RoleCreate) (*dto.RoleRead, error) {
	name, err := validateRoleName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := s.checkNameUnique(ctx, name, nil); err != nil {
		return nil, err
	}
	perms, err := validatePermissions(in.Permissions)
	if err != nil {
		return nil, err
	}
	r := &models.Role{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		Level:       in.Level,
		IsSystem:    false,
		Permissions: perms,
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Get(ctx context.Context, id uuid.UUID) (*dto.RoleRead, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Role")
		}
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Update(ctx context.Context, id uuid.UUID, in dto.RoleUpdate) (*dto.RoleRead, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Role")
		}
		return nil, err
	}

	if in.Name != nil {
		name, err := validateRoleName(*in.Name)
		if err != nil {
			return nil, err
		}
		if r.IsSystem && name != r.Name {
			return nil, apperrors.ErrBadRequest("Cannot rename a system role")
		}
		if err := s.checkNameUnique(ctx, name, &r.ID); err != nil {
			return nil, err
		}
		r.Name = name
	}
	if in.Level != nil {
		if r.IsSystem && *in.Level != r.Level {
			return nil, apperrors.ErrBadRequest("Cannot change the level of a system role")
		}
		r.Level = *in.Level
	}
	if in.Description != nil {
		r.Description = strings.TrimSpace(*in.Description)
	}
	if in.Permissions != nil {
		perms, err := validatePermissions(*in.Permissions)
		if err != nil {
			return nil, err
		}
		r.Permissions = perms
	}

	if err := s.repo.Update(ctx, r); err != nil {
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Delete(ctx context.Context, id uuid.UUID) error {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Role")
		}
		return err
	}
	if r.IsSystem {
		return apperrors.ErrBadRequest("Cannot delete a system role")
	}
	count, err := s.repo.CountUsersWithRole(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		word := "user is"
		if count > 1 {
			word = "users are"
		}
		return apperrors.ErrConflict(fmt.Sprintf(
			"Cannot delete role '%s' — %d %s currently assigned. Please reassign them before deleting.", r.Name, count, word))
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *RoleService) List(ctx context.Context, q dto.RoleListQuery) (*dto.PaginatedData[dto.RoleRead], error) {
	items, total, err := s.repo.List(ctx, repositories.RoleFilter{Page: q.Page, PageSize: q.PageSize, Search: q.Search})
	if err != nil {
		return nil, err
	}
	reads := make([]dto.RoleRead, 0, len(items))
	for i := range items {
		reads = append(reads, roleToRead(&items[i]))
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
	return &dto.PaginatedData[dto.RoleRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
