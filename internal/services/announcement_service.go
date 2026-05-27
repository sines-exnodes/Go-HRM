package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// HubBroadcaster abstracts the SSE hub so the service can broadcast
// without importing the concrete sse package — keeps the service unit-
// testable with a tiny mock. The signature mirrors sse.Hub.Broadcast,
// using a `string` event type for testability (Phase 7 plan §16).
type HubBroadcaster interface {
	Broadcast(eventType string, data any, filter func(userID uuid.UUID) bool)
}

// AnnouncementService owns the announcement aggregate. Permission gating
// is two-layered: route-level RequirePerms upstream; service-level
// ownership branch via the asAdmin bool precomputed by the handler from
// the JWT-preloaded user.Roles. Same shape as LeaveService /
// AttendanceService.
type AnnouncementService struct {
	repo   repositories.AnnouncementRepository
	emps   repositories.EmployeeRepository
	depts  repositories.DepartmentRepository
	labels repositories.LabelRepository
	hub    HubBroadcaster // optional — nil disables SSE broadcasts
}

// NewAnnouncementService constructs an AnnouncementService. Pass nil
// for hub in tests that don't need to assert broadcasts.
func NewAnnouncementService(
	repo repositories.AnnouncementRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	labels repositories.LabelRepository,
	hub HubBroadcaster,
) *AnnouncementService {
	return &AnnouncementService{repo: repo, emps: emps, depts: depts, labels: labels, hub: hub}
}

// ---- Shared helpers ----

// resolveCurrentEmployee returns the employee row for the authenticated
// user. Missing record yields 403 — same pattern as LeaveService.
func (s *AnnouncementService) resolveCurrentEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	emp, err := s.emps.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden("No employee record for current user")
		}
		return nil, err
	}
	return emp, nil
}

// validateLabelIDs ensures every label_id references a live label row.
// Returns 400 with the missing IDs on failure.
func (s *AnnouncementService) validateLabelIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	labels, err := s.labels.List(ctx) // small catalog — full scan is fine
	if err != nil {
		return err
	}
	known := make(map[uuid.UUID]struct{}, len(labels))
	for _, l := range labels {
		known[l.ID] = struct{}{}
	}
	missing := make([]string, 0)
	for _, id := range ids {
		if _, ok := known[id]; !ok {
			missing = append(missing, id.String())
		}
	}
	if len(missing) > 0 {
		return apperrors.ErrBadRequest("unknown label_id(s): " + strings.Join(missing, ","))
	}
	return nil
}

// validateDepartmentIDs ensures every department_id references a live
// department row.
func (s *AnnouncementService) validateDepartmentIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	missing := make([]string, 0)
	for _, id := range ids {
		d, err := s.depts.FindByID(ctx, id, false)
		if err != nil || d == nil {
			missing = append(missing, id.String())
		}
	}
	if len(missing) > 0 {
		return apperrors.ErrBadRequest("unknown department_id(s): " + strings.Join(missing, ","))
	}
	return nil
}

// populateRead inflates the model into the canonical wire shape.
// hasViewed is computed by the caller and passed in.
func (s *AnnouncementService) populateRead(a *models.Announcement, hasViewed bool) dto.AnnouncementRead {
	out := dto.AnnouncementRead{
		ID:                a.ID,
		Title:             a.Title,
		Description:       a.Description,
		Summary:           a.Summary,
		Status:            a.Status,
		ScheduledAt:       a.ScheduledAt,
		PublishedAt:       a.PublishedAt,
		TargetAudience:    a.TargetAudience,
		Pinned:            a.Pinned,
		CoverImageURL:     a.CoverImageURL,
		Labels:            make([]dto.AnnouncementLabelBrief, 0, len(a.AnnouncementLabels)),
		TargetDepartments: make([]dto.AnnouncementDepartmentBrief, 0, len(a.TargetDepartments)),
		Attachments:       make([]dto.AnnouncementAttachmentRead, 0, len(a.Attachments)),
		HasViewed:         hasViewed,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
	if a.Author != nil {
		out.Author = &dto.AnnouncementAuthorBrief{
			ID:        a.Author.ID,
			FullName:  a.Author.FullName,
			AvatarURL: a.Author.AvatarURL,
		}
	}
	for _, al := range a.AnnouncementLabels {
		if al.Label == nil {
			continue
		}
		out.Labels = append(out.Labels, dto.AnnouncementLabelBrief{
			ID:   al.Label.ID,
			Name: al.Label.Name,
		})
	}
	for _, td := range a.TargetDepartments {
		if td.Department == nil {
			continue
		}
		out.TargetDepartments = append(out.TargetDepartments, dto.AnnouncementDepartmentBrief{
			ID:   td.Department.ID,
			Name: td.Department.Name,
		})
	}
	for _, att := range a.Attachments {
		out.Attachments = append(out.Attachments, dto.AnnouncementAttachmentRead{
			ID:          att.ID,
			URL:         att.URL,
			Filename:    att.Filename,
			ContentType: att.ContentType,
			SizeBytes:   att.SizeBytes,
			CreatedAt:   att.CreatedAt,
		})
	}
	return out
}

// broadcastPublished is a thin wrapper around the hub. Nil-safe so
// service tests without a hub still work.
func (s *AnnouncementService) broadcastPublished(a *models.Announcement) {
	if s.hub == nil || a.PublishedAt == nil {
		return
	}
	deptIDs := make([]uuid.UUID, 0, len(a.TargetDepartments))
	for _, td := range a.TargetDepartments {
		deptIDs = append(deptIDs, td.DepartmentID)
	}
	payload := dto.SSEAnnouncementPublishedEvent{
		ID:             a.ID,
		Title:          a.Title,
		Summary:        a.Summary,
		TargetAudience: a.TargetAudience,
		DepartmentIDs:  deptIDs,
		Pinned:         a.Pinned,
		PublishedAt:    *a.PublishedAt,
	}
	// Broadcast to ALL — the FE refetches the visible list on receipt
	// and that refetch goes through the GET visibility filter. Avoids
	// duplicating the audience logic in the hub layer.
	s.hub.Broadcast("announcement_published", payload, nil)
}

// ---- Create ----

// Create inserts a new announcement. Status defaults to draft. When
// status='published' is supplied at create time, published_at is stamped
// and the SSE broadcast fires (same code path as Publish).
func (s *AnnouncementService) Create(ctx context.Context, currentUserID uuid.UUID, in dto.AnnouncementCreate) (*dto.AnnouncementRead, error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if err := s.validateLabelIDs(ctx, in.LabelIDs); err != nil {
		return nil, err
	}
	if err := s.validateDepartmentIDs(ctx, in.DepartmentIDs); err != nil {
		return nil, err
	}

	row := &models.Announcement{
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Summary:     in.Summary,
		AuthorID:    currentEmp.ID,
		Status:      models.AnnouncementStatusDraft,
	}
	// send_now is the Python-parity shortcut: when true and status is
	// not explicitly set, promote the draft to published immediately
	// (mirrors what POST /:id/publish would do post-create). Explicit
	// status wins — sending status=draft + send_now=true keeps it draft.
	if in.Status != nil {
		row.Status = *in.Status
	} else if in.SendNow {
		row.Status = models.AnnouncementStatusPublished
	}
	if in.ScheduledAt != nil {
		row.ScheduledAt = in.ScheduledAt
	}
	if in.TargetAudience != nil {
		row.TargetAudience = *in.TargetAudience
	} else {
		row.TargetAudience = models.AnnouncementAudienceAll
	}
	if in.Pinned != nil {
		row.Pinned = *in.Pinned
	}
	if in.CoverImageURL != nil {
		row.CoverImageURL = in.CoverImageURL
	}
	if row.Status == models.AnnouncementStatusPublished && row.PublishedAt == nil {
		now := time.Now().UTC()
		row.PublishedAt = &now
	}
	// department audience requires at least one target
	if row.TargetAudience == models.AnnouncementAudienceDepartment && len(in.DepartmentIDs) == 0 {
		return nil, apperrors.ErrBadRequest("target_audience=department requires at least one department_id")
	}

	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	if len(in.LabelIDs) > 0 {
		if err := s.repo.ReplaceLabels(ctx, row.ID, in.LabelIDs); err != nil {
			return nil, err
		}
	}
	if len(in.DepartmentIDs) > 0 {
		if err := s.repo.ReplaceTargetDepartments(ctx, row.ID, in.DepartmentIDs); err != nil {
			return nil, err
		}
	}

	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	if final.Status == models.AnnouncementStatusPublished {
		s.broadcastPublished(final)
	}
	read := s.populateRead(final, false)
	return &read, nil
}

// ---- Update ----

// Update patches an existing announcement. Owner can edit own; admin
// can edit any. Status transitions through the same gate. Replaces
// labels/departments only when the corresponding pointer-slice is non-
// nil (per DTO doc).
func (s *AnnouncementService) Update(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool, in dto.AnnouncementUpdate) (*dto.AnnouncementRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Announcement")
		}
		return nil, err
	}
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if !asAdmin && row.AuthorID != currentEmp.ID {
		return nil, apperrors.ErrForbidden("You do not own this announcement")
	}

	if in.LabelIDs != nil {
		if err := s.validateLabelIDs(ctx, *in.LabelIDs); err != nil {
			return nil, err
		}
	}
	if in.DepartmentIDs != nil {
		if err := s.validateDepartmentIDs(ctx, *in.DepartmentIDs); err != nil {
			return nil, err
		}
	}

	if in.Title != nil {
		row.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		row.Description = *in.Description
	}
	if in.Summary != nil {
		row.Summary = in.Summary
	}
	if in.ScheduledAt != nil {
		row.ScheduledAt = in.ScheduledAt
	}
	if in.TargetAudience != nil {
		row.TargetAudience = *in.TargetAudience
	}
	if in.Pinned != nil {
		row.Pinned = *in.Pinned
	}
	if in.CoverImageURL != nil {
		row.CoverImageURL = in.CoverImageURL
	}
	previouslyPublished := row.Status == models.AnnouncementStatusPublished
	if in.Status != nil {
		row.Status = *in.Status
		if row.Status == models.AnnouncementStatusPublished && row.PublishedAt == nil {
			now := time.Now().UTC()
			row.PublishedAt = &now
		}
	}

	if err := s.repo.Update(ctx, row); err != nil {
		return nil, err
	}
	if in.LabelIDs != nil {
		if err := s.repo.ReplaceLabels(ctx, row.ID, *in.LabelIDs); err != nil {
			return nil, err
		}
	}
	if in.DepartmentIDs != nil {
		if err := s.repo.ReplaceTargetDepartments(ctx, row.ID, *in.DepartmentIDs); err != nil {
			return nil, err
		}
	}

	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	// Broadcast only on draft->published transitions inside this Update
	// (not on edit-of-already-published — avoids spam).
	if !previouslyPublished && final.Status == models.AnnouncementStatusPublished {
		s.broadcastPublished(final)
	}
	read := s.populateRead(final, false)
	return &read, nil
}

// ---- Delete ----

// Delete soft-deletes an announcement. Owner can delete own; admin can
// delete any.
func (s *AnnouncementService) Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) error {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Announcement")
		}
		return err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return err
		}
		if row.AuthorID != currentEmp.ID {
			return apperrors.ErrForbidden("You do not own this announcement")
		}
	}
	return s.repo.SoftDelete(ctx, id)
}

// ---- Publish ----

// Publish transitions a non-published announcement to published, stamps
// published_at if absent, and broadcasts via the SSE hub. Already-
// published rows are a no-op (returned as-is, no rebroadcast).
func (s *AnnouncementService) Publish(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.AnnouncementRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Announcement")
		}
		return nil, err
	}
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if !asAdmin && row.AuthorID != currentEmp.ID {
		return nil, apperrors.ErrForbidden("You do not own this announcement")
	}

	wasPublished := row.Status == models.AnnouncementStatusPublished
	if !wasPublished {
		row.Status = models.AnnouncementStatusPublished
		if row.PublishedAt == nil {
			now := time.Now().UTC()
			row.PublishedAt = &now
		}
		if err := s.repo.Update(ctx, row); err != nil {
			return nil, err
		}
	}

	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	if !wasPublished {
		s.broadcastPublished(final)
	}
	read := s.populateRead(final, false)
	return &read, nil
}

// ---- Get + visibility ----

// Get returns one announcement. Admins see everything; non-admins must
// satisfy the visibility predicate (REVISION NOTES #14). View tracking
// is NOT auto-triggered on Get — clients call MarkViewed explicitly
// (matches the Python contract: read marker is a deliberate action).
func (s *AnnouncementService) Get(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.AnnouncementRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Announcement")
		}
		return nil, err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		if !s.canSee(row, currentEmp) {
			return nil, apperrors.ErrForbidden("You cannot view this announcement")
		}
	}
	hasViewed, err := s.repo.HasViewed(ctx, row.ID, currentUserID)
	if err != nil {
		return nil, err
	}
	read := s.populateRead(row, hasViewed)
	return &read, nil
}

// canSee evaluates the visibility predicate at the row level. Used by
// Get; the List/Mobile paths apply the SAME predicate at the SQL layer
// via AnnouncementListFilter for efficiency.
func (s *AnnouncementService) canSee(a *models.Announcement, emp *models.Employee) bool {
	if a.Status != models.AnnouncementStatusPublished {
		return a.AuthorID == emp.ID
	}
	if a.AuthorID == emp.ID {
		return true
	}
	if a.TargetAudience == models.AnnouncementAudienceAll {
		return true
	}
	if a.TargetAudience == models.AnnouncementAudienceDepartment && emp.DepartmentID != nil {
		for _, td := range a.TargetDepartments {
			if td.DepartmentID == *emp.DepartmentID {
				return true
			}
		}
	}
	return false
}

// ---- MarkViewed ----

// MarkViewed records that the caller read the announcement. Idempotent:
// second call from the same (announcement, user) pair is a no-op (the
// repo uses ON CONFLICT DO NOTHING — preserves the FIRST view time per
// Python contract).
func (s *AnnouncementService) MarkViewed(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) error {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Announcement")
		}
		return err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return err
		}
		if !s.canSee(row, currentEmp) {
			return apperrors.ErrForbidden("You cannot view this announcement")
		}
	}
	return s.repo.UpsertView(ctx, id, currentUserID)
}

// ---- List ----

// List returns paginated announcements. Non-admins see only rows that
// satisfy the visibility predicate (handled in the repo via
// AnnouncementListFilter). Admins see everything filtered by the
// non-visibility filters only.
func (s *AnnouncementService) List(ctx context.Context, currentUserID uuid.UUID, asAdmin bool, q dto.AnnouncementListQuery) (dto.PaginatedData[dto.AnnouncementRead], error) {
	f := repositories.AnnouncementListFilter{
		AsAdmin:  asAdmin,
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
		Scope:    q.Scope,
		Pinned:   q.Pinned,
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return dto.PaginatedData[dto.AnnouncementRead]{}, err
		}
		f.CurrentEmployeeID = currentEmp.ID
		f.CurrentDepartmentID = currentEmp.DepartmentID
	}
	if q.Status != "" {
		// CSV-ish? Accept single value for now.
		f.Statuses = []models.AnnouncementStatus{models.AnnouncementStatus(q.Status)}
	}
	if q.LabelID != "" {
		lid, err := uuid.Parse(q.LabelID)
		if err != nil {
			return dto.PaginatedData[dto.AnnouncementRead]{}, apperrors.ErrBadRequest("invalid label_id")
		}
		f.LabelID = &lid
	}
	if q.DepartmentID != "" {
		did, err := uuid.Parse(q.DepartmentID)
		if err != nil {
			return dto.PaginatedData[dto.AnnouncementRead]{}, apperrors.ErrBadRequest("invalid department_id")
		}
		f.DepartmentID = &did
	}

	rows, total, err := s.repo.List(ctx, f)
	if err != nil {
		return dto.PaginatedData[dto.AnnouncementRead]{}, err
	}

	// Batch-fetch viewed state for the current user across all rows.
	viewed := make(map[uuid.UUID]bool, len(rows))
	for i := range rows {
		v, err := s.repo.HasViewed(ctx, rows[i].ID, currentUserID)
		if err != nil {
			return dto.PaginatedData[dto.AnnouncementRead]{}, err
		}
		viewed[rows[i].ID] = v
	}

	items := make([]dto.AnnouncementRead, 0, len(rows))
	for i := range rows {
		items = append(items, s.populateRead(&rows[i], viewed[rows[i].ID]))
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
	return dto.PaginatedData[dto.AnnouncementRead]{
		Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
	}, nil
}

// ---- Mobile-specific ----

// MobileList returns the mobile-shaped projection. Always
// visibility-filtered (mobile never sees drafts/scheduled rows).
func (s *AnnouncementService) MobileList(ctx context.Context, currentUserID uuid.UUID, q dto.MobileAnnouncementListQuery) (dto.PaginatedData[dto.MobileAnnouncementBrief], error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return dto.PaginatedData[dto.MobileAnnouncementBrief]{}, err
	}
	f := repositories.AnnouncementListFilter{
		AsAdmin:             false,
		Scope:               "targeted-at-me",
		CurrentEmployeeID:   currentEmp.ID,
		CurrentDepartmentID: currentEmp.DepartmentID,
		Page:                q.Page,
		PageSize:            q.PageSize,
	}
	rows, total, err := s.repo.List(ctx, f)
	if err != nil {
		return dto.PaginatedData[dto.MobileAnnouncementBrief]{}, err
	}
	items := make([]dto.MobileAnnouncementBrief, 0, len(rows))
	for i := range rows {
		v, err := s.repo.HasViewed(ctx, rows[i].ID, currentUserID)
		if err != nil {
			return dto.PaginatedData[dto.MobileAnnouncementBrief]{}, err
		}
		items = append(items, s.toMobileBrief(&rows[i], v))
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
	return dto.PaginatedData[dto.MobileAnnouncementBrief]{
		Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
	}, nil
}

// MobileBrief returns the top 5 latest published announcements targeting
// the current user — the home-screen widget projection. Always
// visibility-filtered. No pagination metadata; the slice is the data.
// Mirrors Python's `GET /mobile/announcements/` contract.
func (s *AnnouncementService) MobileBrief(ctx context.Context, currentUserID uuid.UUID) ([]dto.MobileAnnouncementBrief, error) {
	const briefLimit = 5
	page, err := s.MobileList(ctx, currentUserID, dto.MobileAnnouncementListQuery{
		Page:     1,
		PageSize: briefLimit,
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// MobileGet returns the full detail (Description included) when visible
// to the current user. Same visibility rules as Get.
func (s *AnnouncementService) MobileGet(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) (*dto.AnnouncementRead, error) {
	// Mobile never has admin manage perm — always non-admin path.
	return s.Get(ctx, id, currentUserID, false)
}

// toMobileBrief is the projection from model → MobileAnnouncementBrief.
func (s *AnnouncementService) toMobileBrief(a *models.Announcement, viewed bool) dto.MobileAnnouncementBrief {
	out := dto.MobileAnnouncementBrief{
		ID:            a.ID,
		Title:         a.Title,
		Summary:       a.Summary,
		CoverImageURL: a.CoverImageURL,
		Status:        a.Status,
		Pinned:        a.Pinned,
		PublishedAt:   a.PublishedAt,
		Labels:        make([]dto.AnnouncementLabelBrief, 0, len(a.AnnouncementLabels)),
		HasViewed:     viewed,
	}
	for _, al := range a.AnnouncementLabels {
		if al.Label == nil {
			continue
		}
		out.Labels = append(out.Labels, dto.AnnouncementLabelBrief{
			ID:   al.Label.ID,
			Name: al.Label.Name,
		})
	}
	return out
}
