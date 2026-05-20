package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// LabelService is intentionally tiny: the Python source only exposes
// list-all and get-or-create on announcement_labels (see REVISION NOTES
// item #4). No update, no delete.
type LabelService struct {
	repo repositories.LabelRepository
}

func NewLabelService(repo repositories.LabelRepository) *LabelService {
	return &LabelService{repo: repo}
}

func labelToRead(l *models.Label) dto.LabelRead {
	return dto.LabelRead{
		ID:        l.ID,
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

// GetOrCreateResult is returned by GetOrCreate so the handler can choose
// the correct HTTP status (200 vs 201) without re-querying.
type GetOrCreateResult struct {
	Label   dto.LabelRead
	Created bool
}

// validateLabelName enforces the Python contract: trimmed, 1..50 chars.
// No regex — labels are free-form (the Python schema only sets max length).
func validateLabelName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", apperrors.ErrBadRequest("Label name cannot be blank")
	}
	if len(name) > 50 {
		return "", apperrors.ErrBadRequest(fmt.Sprintf("Label name must be at most %d characters", 50))
	}
	return name, nil
}

// GetOrCreate is the only write path. If a live label with the same
// case-insensitive name exists, it is returned (Created=false); otherwise
// a new row is inserted (Created=true). This matches the Python source's
// idempotent POST.
func (s *LabelService) GetOrCreate(ctx context.Context, in dto.LabelCreate) (*GetOrCreateResult, error) {
	name, err := validateLabelName(in.Name)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		out := labelToRead(existing)
		return &GetOrCreateResult{Label: out, Created: false}, nil
	}
	row := &models.Label{Name: name}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	out := labelToRead(row)
	return &GetOrCreateResult{Label: out, Created: true}, nil
}

// List returns every live label, sorted by name ASC. No pagination
// (matches the Python source: announcement labels are a small set used
// for filtering, so the FE wants them all in one shot).
func (s *LabelService) List(ctx context.Context) ([]dto.LabelRead, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.LabelRead, 0, len(items))
	for i := range items {
		out = append(out, labelToRead(&items[i]))
	}
	return out, nil
}
