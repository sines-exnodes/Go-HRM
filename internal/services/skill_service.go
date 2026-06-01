package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// ---- Skill validation + upload constants ----
//
// Mirrors the Python source's skill regex/length rules so the public
// contract is unchanged across the migration.
var skillNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 &.+#/()-]+$`)

const (
	skillNameMaxLen = 100
	skillDescMaxLen = 500

	skillIconSubdir   = "skill-icons"
	skillIconMaxBytes = 5 * 1024 * 1024
)

// allowedSkillIconMIME is the set of image content types accepted for a
// skill icon. The authoritative check sniffs the actual file bytes
// (http.DetectContentType) — never trust the client's Content-Type header
// (review-fix #2 from Phase 2 hardened this same pattern for avatars).
var allowedSkillIconMIME = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

// SkillIconUpload bundles the multipart icon fields. Nil means "no icon
// supplied in this request" (preserve any existing icon on update).
type SkillIconUpload struct {
	Content     []byte
	ContentType string // client-supplied, used only as a hint
	Ext         string // ".png" / ".jpg" / etc — lowercased by caller
}

// SkillService owns the skill catalog and the cross-aggregate delete
// guard (a skill currently assigned to one or more employees cannot be
// deleted — mirrors the Python source and the Phase 3 department guard).
type SkillService struct {
	repo      repositories.SkillRepository
	empSkills repositories.EmployeeSkillRepository
	emps      repositories.EmployeeRepository
	uploads   Uploader // optional; nil means icon upload is unavailable
}

func NewSkillService(
	repo repositories.SkillRepository,
	empSkills repositories.EmployeeSkillRepository,
	emps repositories.EmployeeRepository,
	uploads Uploader,
) *SkillService {
	return &SkillService{repo: repo, empSkills: empSkills, emps: emps, uploads: uploads}
}

func skillToRead(s *models.Skill) dto.SkillRead {
	return dto.SkillRead{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		IconURL:     s.IconURL,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// validateSkillName enforces the Python contract: trimmed, 1..100 chars,
// regex-clean. Returns the trimmed/cleaned name on success.
func validateSkillName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", apperrors.ErrBadRequest("Skill name cannot be blank")
	}
	if len(name) > skillNameMaxLen {
		return "", apperrors.ErrBadRequest(fmt.Sprintf("Skill name must be at most %d characters", skillNameMaxLen))
	}
	if !skillNameRegex.MatchString(name) {
		return "", apperrors.ErrBadRequest("Skill name contains invalid characters")
	}
	return name, nil
}

func validateSkillDescription(raw string) (string, error) {
	desc := strings.TrimSpace(raw)
	if len(desc) > skillDescMaxLen {
		return "", apperrors.ErrBadRequest(fmt.Sprintf("Skill description must be at most %d characters", skillDescMaxLen))
	}
	return desc, nil
}

func (s *SkillService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
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
	return apperrors.ErrConflict("Skill name already exists")
}

// uploadIcon validates and stores a skill icon, returning the public URL.
// Content-type is determined by sniffing the actual bytes, NOT trusting
// the client header (review-fix #2 pattern).
func (s *SkillService) uploadIcon(ctx context.Context, icon SkillIconUpload) (string, error) {
	if s.uploads == nil {
		return "", apperrors.ErrInternal("Storage is not configured; cannot upload skill icon")
	}
	if len(icon.Content) == 0 {
		return "", apperrors.ErrBadRequest("Icon file is empty")
	}
	if len(icon.Content) > skillIconMaxBytes {
		return "", apperrors.ErrBadRequest("Icon must not exceed 5MB")
	}
	sniffLen := len(icon.Content)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffed := http.DetectContentType(icon.Content[:sniffLen])
	// DetectContentType returns "image/svg+xml" only when the file starts
	// with the literal "<svg" or XML preamble, so any spoofed PNG that's
	// actually a script will fail this check.
	if !allowedSkillIconMIME[sniffed] {
		return "", apperrors.ErrBadRequest("Icon must be a valid image (PNG, JPEG, GIF, WEBP, or SVG)")
	}
	url, err := s.uploads.Upload(ctx, skillIconSubdir, icon.Ext, icon.Content, sniffed)
	if err != nil {
		return "", err
	}
	return url, nil
}

// Create inserts a new skill. The icon (if any) is uploaded BEFORE the row
// is persisted, so a Postgres-side failure won't leave an orphaned object.
func (s *SkillService) Create(ctx context.Context, in dto.SkillCreate, icon *SkillIconUpload) (*dto.SkillRead, error) {
	name, err := validateSkillName(in.Name)
	if err != nil {
		return nil, err
	}
	desc, err := validateSkillDescription(in.Description)
	if err != nil {
		return nil, err
	}
	if err := s.checkNameUnique(ctx, name, nil); err != nil {
		return nil, err
	}
	var iconURL *string
	if icon != nil {
		url, err := s.uploadIcon(ctx, *icon)
		if err != nil {
			return nil, err
		}
		iconURL = &url
	}
	row := &models.Skill{
		Name:        name,
		Description: desc,
		IconURL:     iconURL,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		// If we uploaded an icon, try to clean it up so we don't leak.
		if iconURL != nil && s.uploads != nil {
			_ = s.uploads.Delete(ctx, *iconURL)
		}
		return nil, err
	}
	out := skillToRead(row)
	return &out, nil
}

// Update patches a skill. A new icon (if provided) replaces the previous
// one; the previous object is best-effort deleted from storage AFTER the
// DB row is updated successfully.
func (s *SkillService) Update(ctx context.Context, id uuid.UUID, in dto.SkillUpdate, icon *SkillIconUpload) (*dto.SkillRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Skill")
		}
		return nil, err
	}
	if in.Name != nil {
		name, err := validateSkillName(*in.Name)
		if err != nil {
			return nil, err
		}
		if err := s.checkNameUnique(ctx, name, &row.ID); err != nil {
			return nil, err
		}
		row.Name = name
	}
	if in.Description != nil {
		desc, err := validateSkillDescription(*in.Description)
		if err != nil {
			return nil, err
		}
		row.Description = desc
	}

	prevIcon := row.IconURL
	if icon != nil {
		url, err := s.uploadIcon(ctx, *icon)
		if err != nil {
			return nil, err
		}
		row.IconURL = &url
	}

	if err := s.repo.Update(ctx, row); err != nil {
		// New icon was uploaded but DB write failed — clean it up.
		if icon != nil && row.IconURL != nil && s.uploads != nil {
			_ = s.uploads.Delete(ctx, *row.IconURL)
		}
		return nil, err
	}
	// Best-effort delete of the prior icon object after the DB commit.
	if icon != nil && prevIcon != nil && *prevIcon != "" && s.uploads != nil {
		_ = s.uploads.Delete(ctx, *prevIcon)
	}
	out := skillToRead(row)
	return &out, nil
}

// Delete soft-deletes a skill after verifying no live employee_skills row
// references it. The 409 body carries the conflict details so the FE can
// show "still assigned to N employees" without a follow-up call.
func (s *SkillService) Delete(ctx context.Context, id uuid.UUID) error {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Skill")
		}
		return err
	}
	count, err := s.empSkills.CountEmployeesBySkill(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		word := "employee is"
		if count > 1 {
			word = "employees are"
		}
		return apperrors.ErrConflict(fmt.Sprintf(
			"Cannot delete — %d %s still assigned to this skill. Unassign them first.", count, word)).
			WithDetails(map[string]any{
				"skill_id":       row.ID,
				"skill_name":     row.Name,
				"employee_count": count,
			})
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *SkillService) Get(ctx context.Context, id uuid.UUID) (*dto.SkillRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Skill")
		}
		return nil, err
	}
	out := skillToRead(row)
	return &out, nil
}

func (s *SkillService) List(ctx context.Context, q dto.SkillListQuery) (*dto.PaginatedData[dto.SkillRead], error) {
	items, total, err := s.repo.List(ctx, repositories.SkillFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
	})
	if err != nil {
		return nil, err
	}
	reads := make([]dto.SkillRead, 0, len(items))
	for i := range items {
		reads = append(reads, skillToRead(&items[i]))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return &dto.PaginatedData[dto.SkillRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}

// ---- Employee ↔ Skill assignment ----

// ListForEmployee returns the live skills assigned to an employee.
// The 404-on-missing-employee check happens here so the handler doesn't
// need to inline an employee lookup.
func (s *SkillService) ListForEmployee(ctx context.Context, employeeID uuid.UUID) ([]dto.SkillRead, error) {
	if _, err := s.emps.FindByID(ctx, employeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	rows, err := s.empSkills.ListByEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SkillRead, 0, len(rows))
	for _, row := range rows {
		// Defensive: a join row may reference a soft-deleted skill if the
		// preload's NotDeleted scope returned nil. Skip those.
		if row.Skill == nil {
			continue
		}
		out = append(out, skillToRead(row.Skill))
	}
	return out, nil
}

// ValidateSkillIDs de-duplicates the requested set and verifies every id
// references a live skill. Returns the cleaned slice or a 400 with the
// offending id — used both by ReplaceForEmployee and by inline assignment on
// employee create/update.
func (s *SkillService) ValidateSkillIDs(ctx context.Context, skillIDs []uuid.UUID) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{}, len(skillIDs))
	cleaned := make([]uuid.UUID, 0, len(skillIDs))
	for _, sid := range skillIDs {
		if sid == uuid.Nil {
			return nil, apperrors.ErrBadRequest("skill_ids contains an empty UUID")
		}
		if _, dup := seen[sid]; dup {
			continue
		}
		seen[sid] = struct{}{}
		if _, err := s.repo.FindByID(ctx, sid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.ErrBadRequest(fmt.Sprintf("Skill %s not found", sid))
			}
			return nil, err
		}
		cleaned = append(cleaned, sid)
	}
	return cleaned, nil
}

// ReplaceForEmployee atomically replaces the employee's skill set. Every
// requested skill_id must reference a live skill row; the call fails
// 400 BadRequest otherwise (no partial application).
func (s *SkillService) ReplaceForEmployee(ctx context.Context, employeeID uuid.UUID, skillIDs []uuid.UUID) ([]dto.SkillRead, error) {
	if _, err := s.emps.FindByID(ctx, employeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	cleaned, err := s.ValidateSkillIDs(ctx, skillIDs)
	if err != nil {
		return nil, err
	}
	if err := s.empSkills.ReplaceForEmployee(ctx, employeeID, cleaned); err != nil {
		return nil, err
	}
	return s.ListForEmployee(ctx, employeeID)
}
