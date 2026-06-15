package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// HolidayService manages company holiday records and triggers leave recalculation.
type HolidayService struct {
	repo         repositories.HolidayRepository
	templateRepo repositories.HolidayTemplateRepository
	leaveRepo    repositories.LeaveRequestRepository
}

// NewHolidayService constructs a HolidayService.
func NewHolidayService(
	repo repositories.HolidayRepository,
	templateRepo repositories.HolidayTemplateRepository,
	leaveRepo repositories.LeaveRequestRepository,
) *HolidayService {
	return &HolidayService{repo: repo, templateRepo: templateRepo, leaveRepo: leaveRepo}
}

func toHolidayRead(h models.Holiday) dto.HolidayRead {
	return dto.HolidayRead{
		ID:        h.ID,
		Year:      h.Year,
		Name:      h.Name,
		FromDate:  h.FromDate,
		ToDate:    h.ToDate,
		TotalDays: int(h.ToDate.Sub(h.FromDate).Hours()/24) + 1,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

func toHolidayTemplateRead(t models.HolidayTemplate) dto.HolidayTemplateRead {
	return dto.HolidayTemplateRead{
		ID:        t.ID,
		Year:      t.Year,
		Name:      t.Name,
		FromDate:  t.FromDate,
		ToDate:    t.ToDate,
		TotalDays: int(t.ToDate.Sub(t.FromDate).Hours()/24) + 1,
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// recalculateAffectedLeaves recomputes total_days for all Approved leave requests
// that overlap [from, to]. Returns the count of rows updated.
func (s *HolidayService) recalculateAffectedLeaves(ctx context.Context, from, to time.Time) (int, error) {
	leaves, err := s.leaveRepo.FindApprovedOverlapping(ctx, from, to)
	if err != nil {
		return 0, err
	}
	if len(leaves) == 0 {
		return 0, nil
	}
	updates := make([]repositories.TotalDaysUpdate, 0, len(leaves))
	for _, lr := range leaves {
		holidays, err := s.repo.FindInRange(ctx, lr.FromDate, lr.ToDate)
		if err != nil {
			return 0, err
		}
		ranges := make([]utils.DateRange, len(holidays))
		for i, h := range holidays {
			ranges[i] = utils.DateRange{From: h.FromDate, To: h.ToDate}
		}
		td := utils.CalcLeaveDays(lr.FromDate, lr.ToDate, lr.LeavePeriod, ranges)
		updates = append(updates, repositories.TotalDaysUpdate{ID: lr.ID, TotalDays: td})
	}
	return s.leaveRepo.BulkUpdateTotalDays(ctx, updates)
}

// Create inserts a new holiday and recalculates affected approved leaves.
func (s *HolidayService) Create(ctx context.Context, req dto.HolidayCreate) (*dto.HolidayRead, int, *apperrors.AppError) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, 0, apperrors.ErrBadRequest("name is required")
	}
	from := truncateToDate(req.FromDate)
	to := truncateToDate(req.ToDate)
	if to.Before(from) {
		return nil, 0, apperrors.ErrBadRequest("to_date must be on or after from_date")
	}
	exists, err := s.repo.ExistsByNameAndYear(ctx, req.Name, req.Year, nil)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	if exists {
		return nil, 0, apperrors.ErrConflict(fmt.Sprintf("a holiday named %q already exists in %d", req.Name, req.Year))
	}
	h := &models.Holiday{
		Year:     req.Year,
		Name:     strings.TrimSpace(req.Name),
		FromDate: from,
		ToDate:   to,
	}
	if err := s.repo.Create(ctx, h); err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	affected, err := s.recalculateAffectedLeaves(ctx, from, to)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	out := toHolidayRead(*h)
	return &out, affected, nil
}

// Update patches an existing holiday and recalculates leaves for the union of old and new ranges.
func (s *HolidayService) Update(ctx context.Context, id uuid.UUID, req dto.HolidayUpdate) (*dto.HolidayRead, int, *apperrors.AppError) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperrors.ErrNotFound("holiday")
		}
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	oldFrom := h.FromDate
	oldTo := h.ToDate

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, 0, apperrors.ErrBadRequest("name cannot be blank")
		}
		exists, err := s.repo.ExistsByNameAndYear(ctx, name, h.Year, &h.ID)
		if err != nil {
			return nil, 0, apperrors.ErrInternal(err.Error())
		}
		if exists {
			return nil, 0, apperrors.ErrConflict(fmt.Sprintf("a holiday named %q already exists in %d", name, h.Year))
		}
		h.Name = name
	}
	if req.FromDate != nil {
		h.FromDate = truncateToDate(*req.FromDate)
	}
	if req.ToDate != nil {
		h.ToDate = truncateToDate(*req.ToDate)
	}
	if h.ToDate.Before(h.FromDate) {
		return nil, 0, apperrors.ErrBadRequest("to_date must be on or after from_date")
	}
	if err := s.repo.Update(ctx, h); err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}

	recalcFrom := minTime(oldFrom, h.FromDate)
	recalcTo := maxTime(oldTo, h.ToDate)
	affected, err := s.recalculateAffectedLeaves(ctx, recalcFrom, recalcTo)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	out := toHolidayRead(*h)
	return &out, affected, nil
}

// Delete soft-deletes a holiday then recalculates affected approved leaves (restoring days).
func (s *HolidayService) Delete(ctx context.Context, id uuid.UUID) (int, *apperrors.AppError) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.ErrNotFound("holiday")
		}
		return 0, apperrors.ErrInternal(err.Error())
	}
	from := h.FromDate
	to := h.ToDate

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.ErrNotFound("holiday")
		}
		return 0, apperrors.ErrInternal(err.Error())
	}
	// Recalculate AFTER soft-delete so FindInRange excludes the deleted holiday.
	affected, err := s.recalculateAffectedLeaves(ctx, from, to)
	if err != nil {
		return 0, apperrors.ErrInternal(err.Error())
	}
	return affected, nil
}

// List returns a paginated, year-scoped list of holidays ordered by from_date.
func (s *HolidayService) List(ctx context.Context, q dto.HolidayListQuery) (dto.PaginatedData[dto.HolidayRead], *apperrors.AppError) {
	if q.Year < 2000 || q.Year > 2100 {
		return dto.PaginatedData[dto.HolidayRead]{}, apperrors.ErrBadRequest("year must be between 2000 and 2100")
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	rows, total, err := s.repo.List(ctx, repositories.HolidayListQuery{
		Year:     q.Year,
		Search:   q.Search,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		return dto.PaginatedData[dto.HolidayRead]{}, apperrors.ErrInternal(err.Error())
	}
	items := make([]dto.HolidayRead, len(rows))
	for i, h := range rows {
		items[i] = toHolidayRead(h)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(q.PageSize)))
	}
	return dto.PaginatedData[dto.HolidayRead]{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetYears returns distinct years with at least one holiday, always including current year.
func (s *HolidayService) GetYears(ctx context.Context) ([]int, *apperrors.AppError) {
	years, err := s.repo.YearsWithHolidays(ctx)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	currentYear := time.Now().UTC().Year()
	for _, y := range years {
		if y == currentYear {
			return years, nil
		}
	}
	result := make([]int, 0, len(years)+1)
	inserted := false
	for _, y := range years {
		if !inserted && currentYear < y {
			result = append(result, currentYear)
			inserted = true
		}
		result = append(result, y)
	}
	if !inserted {
		result = append(result, currentYear)
	}
	return result, nil
}

// ListTemplates returns preset templates for the given year.
func (s *HolidayService) ListTemplates(ctx context.Context, year int) ([]dto.HolidayTemplateRead, *apperrors.AppError) {
	rows, err := s.templateRepo.ListByYear(ctx, year)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := make([]dto.HolidayTemplateRead, len(rows))
	for i, t := range rows {
		out[i] = toHolidayTemplateRead(t)
	}
	return out, nil
}

// Import bulk-inserts selected templates into the target year's holiday list.
// Duplicates (same name already exists) are skipped.
func (s *HolidayService) Import(ctx context.Context, req dto.HolidayImportRequest) (*dto.HolidayImportResult, *apperrors.AppError) {
	if len(req.TemplateIDs) == 0 {
		return nil, apperrors.ErrBadRequest("template_ids must not be empty")
	}
	templates, err := s.templateRepo.ListByYear(ctx, req.Year)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	wantSet := make(map[uuid.UUID]bool, len(req.TemplateIDs))
	for _, id := range req.TemplateIDs {
		wantSet[id] = true
	}

	imported := 0
	skipped := 0
	var importedRanges []utils.DateRange

	for _, tmpl := range templates {
		if !wantSet[tmpl.ID] {
			continue
		}
		exists, err := s.repo.ExistsByNameAndYear(ctx, tmpl.Name, req.Year, nil)
		if err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
		if exists {
			skipped++
			continue
		}
		h := &models.Holiday{
			Year:     req.Year,
			Name:     tmpl.Name,
			FromDate: tmpl.FromDate,
			ToDate:   tmpl.ToDate,
		}
		if err := s.repo.Create(ctx, h); err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
		imported++
		importedRanges = append(importedRanges, utils.DateRange{From: h.FromDate, To: h.ToDate})
	}

	for _, r := range importedRanges {
		if _, err := s.recalculateAffectedLeaves(ctx, r.From, r.To); err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
	}

	return &dto.HolidayImportResult{Imported: imported, Skipped: skipped}, nil
}
