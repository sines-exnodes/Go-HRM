package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// ListByEmployeeFilter is the per-employee history query used by the
// /history/me endpoint. StartDate/EndDate are inclusive; nil means
// "no bound". Statuses is OR-of-string; empty means "no status filter".
type ListByEmployeeFilter struct {
	Statuses  []string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}

// LeaveRequestRepository defines data access for the leave_requests table.
// Every list/find method returns only non-soft-deleted rows (NotDeleted
// scope). Aggregations (SumApprovedDays, Overlapping) likewise scope to
// is_deleted=false.
type LeaveRequestRepository interface {
	Create(ctx context.Context, lr *models.LeaveRequest) error
	Update(ctx context.Context, lr *models.LeaveRequest) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error)

	// List is the admin/manager listing: paginated, filterable by an
	// optional employee-id allowlist and a status enum list. employeeIDs
	// is nil → no employee filter; empty slice → no matches (caller intent).
	List(ctx context.Context, employeeIDs []uuid.UUID, statuses []string, page, pageSize int) ([]models.LeaveRequest, int64, error)

	// ListByEmployee is the "/history/me" query: rows where to_date is in
	// the past OR status is terminal (rejected/cancelled).
	ListByEmployee(ctx context.Context, employeeID uuid.UUID, filter ListByEmployeeFilter) ([]models.LeaveRequest, int64, error)

	// Upcoming returns pending/approved requests with from_date on or after
	// `today`, capped to `limit`. Used by the dashboard endpoint.
	Upcoming(ctx context.Context, employeeID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error)

	// History returns requests whose to_date is strictly before `today`,
	// capped to `limit`. Used by the dashboard endpoint.
	History(ctx context.Context, employeeID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error)

	// SumApprovedDays groups approved (status='approved') live rows for
	// the given employee whose from_date falls in the calendar year, by
	// leave_type, returning days summed and a row count per type. Used
	// for the balance endpoint.
	SumApprovedDays(ctx context.Context, employeeID uuid.UUID, year int) (map[models.LeaveType]LeaveDaysCount, error)

	// Overlapping returns live pending/approved rows for the employee
	// whose date range intersects [from, to]. excludeID, if non-nil,
	// excludes that row (used when updating an existing request).
	Overlapping(ctx context.Context, employeeID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]models.LeaveRequest, error)
}

// LeaveDaysCount carries the per-type aggregate result of SumApprovedDays.
// Kept as a named type so callers don't need to spell an anonymous struct.
type LeaveDaysCount struct {
	Days  float64
	Count int64
}

type leaveRequestRepo struct{ db *gorm.DB }

// NewLeaveRequestRepository constructs a Postgres-backed
// LeaveRequestRepository.
func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &leaveRequestRepo{db: db}
}

// base applies the NotDeleted scope. Callers Model() the target type
// when issuing Count or aggregation queries.
func (r *leaveRequestRepo) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *leaveRequestRepo) Create(ctx context.Context, lr *models.LeaveRequest) error {
	return r.db.WithContext(ctx).Create(lr).Error
}

func (r *leaveRequestRepo) Update(ctx context.Context, lr *models.LeaveRequest) error {
	return r.db.WithContext(ctx).Save(lr).Error
}

func (r *leaveRequestRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.LeaveRequest{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *leaveRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error) {
	var lr models.LeaveRequest
	if err := r.base(ctx).Where("id = ?", id).First(&lr).Error; err != nil {
		return nil, err
	}
	return &lr, nil
}

func (r *leaveRequestRepo) List(ctx context.Context, employeeIDs []uuid.UUID, statuses []string, page, pageSize int) ([]models.LeaveRequest, int64, error) {
	q := r.base(ctx).Model(&models.LeaveRequest{})
	if employeeIDs != nil {
		if len(employeeIDs) == 0 {
			return []models.LeaveRequest{}, 0, nil
		}
		q = q.Where("employee_id IN ?", employeeIDs)
	}
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var items []models.LeaveRequest
	err := q.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

func (r *leaveRequestRepo) ListByEmployee(ctx context.Context, employeeID uuid.UUID, filter ListByEmployeeFilter) ([]models.LeaveRequest, int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	q := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where(
			"(to_date < ? OR status IN ?)",
			today,
			[]string{string(models.LeaveStatusRejected), string(models.LeaveStatusCancelled)},
		)
	if len(filter.Statuses) > 0 {
		q = q.Where("status IN ?", filter.Statuses)
	}
	if filter.StartDate != nil {
		q = q.Where("from_date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		q = q.Where("to_date <= ?", *filter.EndDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.PageSize
	if size < 1 {
		size = 10
	}
	var items []models.LeaveRequest
	err := q.
		Order("to_date DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *leaveRequestRepo) Upcoming(ctx context.Context, employeeID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error) {
	if limit < 1 {
		limit = 5
	}
	var items []models.LeaveRequest
	err := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where("status IN ?", []string{string(models.LeaveStatusPending), string(models.LeaveStatusApproved)}).
		Where("from_date >= ?", today).
		Order("from_date ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *leaveRequestRepo) History(ctx context.Context, employeeID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error) {
	if limit < 1 {
		limit = 5
	}
	var items []models.LeaveRequest
	err := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where("to_date < ?", today).
		Order("to_date DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// aggRow is the scan target for the per-type aggregate query.
type aggRow struct {
	LeaveType string
	Days      float64
	Count     int64
}

func (r *leaveRequestRepo) SumApprovedDays(ctx context.Context, employeeID uuid.UUID, year int) (map[models.LeaveType]LeaveDaysCount, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	var rows []aggRow
	err := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Select("leave_type, COALESCE(SUM(total_days), 0) AS days, COUNT(*) AS count").
		Where(
			"employee_id = ? AND status = ? AND from_date >= ? AND from_date <= ?",
			employeeID, string(models.LeaveStatusApproved), start, end,
		).
		Group("leave_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make(map[models.LeaveType]LeaveDaysCount, len(rows))
	for _, row := range rows {
		out[models.LeaveType(row.LeaveType)] = LeaveDaysCount{Days: row.Days, Count: row.Count}
	}
	return out, nil
}

func (r *leaveRequestRepo) Overlapping(ctx context.Context, employeeID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]models.LeaveRequest, error) {
	q := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where("status IN ?", []string{string(models.LeaveStatusPending), string(models.LeaveStatusApproved)}).
		Where("from_date <= ? AND to_date >= ?", to, from)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var items []models.LeaveRequest
	err := q.Find(&items).Error
	return items, err
}
