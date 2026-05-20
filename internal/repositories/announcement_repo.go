package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// AnnouncementListFilter is the canonical list-query input. Visibility
// predicate (REVISION NOTES #14) is encoded by the {AsAdmin,
// CurrentEmployeeID, CurrentDepartmentID, Scope} quartet — the repo
// applies it inline so the service doesn't have to compose SQL.
type AnnouncementListFilter struct {
	// Authorization.
	AsAdmin             bool       // when true, skip the visibility WHERE
	CurrentEmployeeID   uuid.UUID  // required when !AsAdmin
	CurrentDepartmentID *uuid.UUID // optional (employee may have no dept)
	// Visibility flavor (only meaningful when !AsAdmin).
	//   "all"             — visibility predicate above (default).
	//   "mine"            — rows authored by CurrentEmployeeID.
	//   "targeted-at-me"  — same as "all" but doesn't add the author OR
	//                       branch (only published + audience match).
	Scope string
	// Generic filters.
	Search       string
	Statuses     []models.AnnouncementStatus
	LabelID      *uuid.UUID
	Pinned       *bool
	DepartmentID *uuid.UUID
	// Paging.
	Page     int
	PageSize int
}

// AnnouncementRepository owns the full announcement aggregate
// (announcements + labels-join + target-departments + attachments +
// views).
type AnnouncementRepository interface {
	// Core CRUD.
	Create(ctx context.Context, a *models.Announcement) error
	Update(ctx context.Context, a *models.Announcement) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Announcement, error)
	List(ctx context.Context, f AnnouncementListFilter) ([]models.Announcement, int64, error)

	// Many-to-many composition. Each Replace* writes inside a transaction:
	// soft-deletes existing rows first, then inserts the new set. Same
	// shape as Phase 4 EmployeeSkill.ReplaceForEmployee.
	ReplaceLabels(ctx context.Context, announcementID uuid.UUID, labelIDs []uuid.UUID) error
	ReplaceTargetDepartments(ctx context.Context, announcementID uuid.UUID, departmentIDs []uuid.UUID) error

	// Attachment lifecycle.
	CreateAttachment(ctx context.Context, att *models.AnnouncementAttachment) error
	SoftDeleteAttachment(ctx context.Context, id uuid.UUID) error
	FindAttachmentByID(ctx context.Context, id uuid.UUID) (*models.AnnouncementAttachment, error)

	// View tracking.
	UpsertView(ctx context.Context, announcementID, userID uuid.UUID) error
	HasViewed(ctx context.Context, announcementID, userID uuid.UUID) (bool, error)
}

type announcementRepo struct{ db *gorm.DB }

// NewAnnouncementRepository constructs a Postgres-backed
// AnnouncementRepository.
func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepo{db: db}
}

// preloadAttachments preloads non-soft-deleted attachments, ordered by
// created_at ASC for stable rendering.
func preloadAttachments(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false).Order("created_at ASC")
}

// preloadAnnouncementLabels preloads live join rows AND their Label.
func preloadAnnouncementLabels(db *gorm.DB) *gorm.DB {
	return db.Where("announcement_labels.is_deleted = ?", false)
}

// preloadTargetDepartments preloads live target-department joins.
func preloadTargetDepartments(db *gorm.DB) *gorm.DB {
	return db.Where("announcement_target_departments.is_deleted = ?", false)
}

func (r *announcementRepo) Create(ctx context.Context, a *models.Announcement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *announcementRepo) Update(ctx context.Context, a *models.Announcement) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *announcementRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Announcement{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *announcementRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Announcement, error) {
	var a models.Announcement
	err := r.db.WithContext(ctx).
		Where("announcements.is_deleted = ?", false).
		Preload("Author").
		Preload("AnnouncementLabels", preloadAnnouncementLabels).
		Preload("AnnouncementLabels.Label").
		Preload("TargetDepartments", preloadTargetDepartments).
		Preload("TargetDepartments.Department").
		Preload("Attachments", preloadAttachments).
		Where("announcements.id = ?", id).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// List applies the visibility predicate (REVISION NOTES #14) for
// non-admin callers and the generic filters for everyone. Pagination
// uses the same (Count + Offset/Limit) pattern as Phase 5 / 6 repos.
func (r *announcementRepo) List(ctx context.Context, f AnnouncementListFilter) ([]models.Announcement, int64, error) {
	// Qualify is_deleted with the table name (REVISION NOTES #11) —
	// joins later make the unqualified scope ambiguous.
	q := r.db.WithContext(ctx).
		Model(&models.Announcement{}).
		Where("announcements.is_deleted = ?", false)

	// ---- Visibility predicate ----
	if !f.AsAdmin {
		switch strings.ToLower(strings.TrimSpace(f.Scope)) {
		case "mine":
			q = q.Where("announcements.author_id = ?", f.CurrentEmployeeID)
		case "targeted-at-me":
			q = q.Where("announcements.status = ?", string(models.AnnouncementStatusPublished))
			q = applyAudienceFilter(q, f.CurrentEmployeeID, f.CurrentDepartmentID, false)
		default: // "" or "all"
			// published AND (author OR audience match).
			q = q.Where("announcements.status = ?", string(models.AnnouncementStatusPublished))
			q = applyAudienceFilter(q, f.CurrentEmployeeID, f.CurrentDepartmentID, true)
		}
	}

	// ---- Generic filters ----
	if s := strings.TrimSpace(f.Search); s != "" {
		pat := utils.BuildILIKEPattern(s)
		q = q.Where("announcements.title ILIKE ? OR announcements.body ILIKE ?", pat, pat)
	}
	if len(f.Statuses) > 0 {
		strs := make([]string, 0, len(f.Statuses))
		for _, st := range f.Statuses {
			strs = append(strs, string(st))
		}
		q = q.Where("announcements.status IN ?", strs)
	}
	if f.LabelID != nil {
		q = q.Joins(
			"JOIN announcement_labels al ON al.announcement_id = announcements.id "+
				"AND al.is_deleted = false AND al.label_id = ?",
			*f.LabelID,
		)
	}
	if f.Pinned != nil {
		q = q.Where("announcements.pinned = ?", *f.Pinned)
	}
	if f.DepartmentID != nil {
		// Explicit per-department filter (admin only). Uses EXISTS to
		// avoid duplicating rows when announcement targets multiple
		// departments. Soft-delete also qualified.
		q = q.Where(
			"EXISTS (SELECT 1 FROM announcement_target_departments td "+
				"WHERE td.announcement_id = announcements.id "+
				"AND td.is_deleted = false AND td.department_id = ?)",
			*f.DepartmentID,
		)
	}

	// ---- Count + paginate ----
	var total int64
	if err := q.Distinct("announcements.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 20
	}

	var rows []models.Announcement
	err := q.
		Preload("Author").
		Preload("AnnouncementLabels", preloadAnnouncementLabels).
		Preload("AnnouncementLabels.Label").
		Preload("TargetDepartments", preloadTargetDepartments).
		Preload("TargetDepartments.Department").
		Preload("Attachments", preloadAttachments).
		Order("announcements.pinned DESC, announcements.published_at DESC NULLS LAST, announcements.created_at DESC").
		Distinct("announcements.*").
		Limit(size).Offset((page - 1) * size).
		Find(&rows).Error
	return rows, total, err
}

// applyAudienceFilter encapsulates the published+(audience match) clause.
// When includeAuthor=true, rows where announcements.author_id = empID
// are also included regardless of audience match (the "all" scope —
// authors always see their own work). When false, only audience matches.
func applyAudienceFilter(q *gorm.DB, empID uuid.UUID, deptID *uuid.UUID, includeAuthor bool) *gorm.DB {
	deptMatch := "EXISTS (SELECT 1 FROM announcement_target_departments td " +
		"WHERE td.announcement_id = announcements.id " +
		"AND td.is_deleted = false AND td.department_id = ?)"
	allMatch := "announcements.target_audience = 'all'"

	if deptID == nil {
		// User has no department — can match only 'all'.
		if includeAuthor {
			return q.Where(
				"(announcements.author_id = ? OR "+allMatch+")",
				empID,
			)
		}
		return q.Where(allMatch)
	}

	if includeAuthor {
		return q.Where(
			"(announcements.author_id = ? OR "+allMatch+" "+
				"OR (announcements.target_audience = 'department' AND "+deptMatch+"))",
			empID, *deptID,
		)
	}
	return q.Where(
		"("+allMatch+" OR (announcements.target_audience = 'department' AND "+deptMatch+"))",
		*deptID,
	)
}

// ReplaceLabels writes the desired set in a transaction. Existing live
// rows for the announcement are soft-deleted first (audit trail), then
// the new set is inserted. Mirrors the Phase-4 employee-skills replace
// pattern.
func (r *announcementRepo) ReplaceLabels(ctx context.Context, announcementID uuid.UUID, labelIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&models.AnnouncementLabel{}).
			Where("announcement_id = ? AND is_deleted = ?", announcementID, false).
			Updates(map[string]any{"is_deleted": true, "deleted_at": &now}).Error; err != nil {
			return err
		}
		if len(labelIDs) == 0 {
			return nil
		}
		rows := make([]models.AnnouncementLabel, 0, len(labelIDs))
		seen := make(map[uuid.UUID]struct{}, len(labelIDs))
		for _, id := range labelIDs {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			rows = append(rows, models.AnnouncementLabel{AnnouncementID: announcementID, LabelID: id})
		}
		return tx.Create(&rows).Error
	})
}

// ReplaceTargetDepartments mirrors ReplaceLabels.
func (r *announcementRepo) ReplaceTargetDepartments(ctx context.Context, announcementID uuid.UUID, departmentIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&models.AnnouncementTargetDepartment{}).
			Where("announcement_id = ? AND is_deleted = ?", announcementID, false).
			Updates(map[string]any{"is_deleted": true, "deleted_at": &now}).Error; err != nil {
			return err
		}
		if len(departmentIDs) == 0 {
			return nil
		}
		rows := make([]models.AnnouncementTargetDepartment, 0, len(departmentIDs))
		seen := make(map[uuid.UUID]struct{}, len(departmentIDs))
		for _, id := range departmentIDs {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			rows = append(rows, models.AnnouncementTargetDepartment{AnnouncementID: announcementID, DepartmentID: id})
		}
		return tx.Create(&rows).Error
	})
}

func (r *announcementRepo) CreateAttachment(ctx context.Context, att *models.AnnouncementAttachment) error {
	return r.db.WithContext(ctx).Create(att).Error
}

func (r *announcementRepo) SoftDeleteAttachment(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.AnnouncementAttachment{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *announcementRepo) FindAttachmentByID(ctx context.Context, id uuid.UUID) (*models.AnnouncementAttachment, error) {
	var att models.AnnouncementAttachment
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("id = ?", id).
		First(&att).Error
	if err != nil {
		return nil, err
	}
	return &att, nil
}

// UpsertView is idempotent — second call from the same user is a no-op
// (the row already exists with the original viewed_at). The Python
// source preserves the FIRST view time, so we do the same: ON CONFLICT
// DO NOTHING.
func (r *announcementRepo) UpsertView(ctx context.Context, announcementID, userID uuid.UUID) error {
	view := &models.AnnouncementView{
		AnnouncementID: announcementID,
		UserID:         userID,
		ViewedAt:       time.Now().UTC(),
	}
	return r.db.WithContext(ctx).
		Set("gorm:insert_option", "ON CONFLICT (announcement_id, user_id) DO NOTHING").
		Create(view).Error
}

// HasViewed returns true when the (announcement, user) pair has a live
// view row.
func (r *announcementRepo) HasViewed(ctx context.Context, announcementID, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.AnnouncementView{}).
		Where("announcement_id = ? AND user_id = ? AND is_deleted = ?", announcementID, userID, false).
		Count(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return n > 0, nil
}
