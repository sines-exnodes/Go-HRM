package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// AttendanceListFilter is the admin listing query. EmployeeID and
// DepartmentID are both nullable; nil means "no filter at that axis".
// Status is the row-level on_time/late string ("" means no filter).
type AttendanceListFilter struct {
	EmployeeID   *uuid.UUID
	DepartmentID *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	Status       string
	Page         int
	PageSize     int
}

// AttendanceRepository defines data access for the attendance aggregate
// (attendance + attendance_sessions). Every list/find method scopes through
// the NotDeleted scope.
type AttendanceRepository interface {
	Create(ctx context.Context, a *models.Attendance) error
	Update(ctx context.Context, a *models.Attendance) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*models.Attendance, error)

	List(ctx context.Context, f AttendanceListFilter) ([]models.Attendance, int64, error)
	MonthlyCheckInCount(ctx context.Context, employeeID uuid.UUID, year, month int) (int64, error)
	DatesWithCheckIn(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]time.Time, error)
	ListForEmployeesInRange(ctx context.Context, employeeIDs []uuid.UUID, from, to time.Time) ([]models.Attendance, error)

	CreateSession(ctx context.Context, s *models.AttendanceSession) error
	UpdateSession(ctx context.Context, s *models.AttendanceSession) error
	FindOpenSession(ctx context.Context, attendanceID uuid.UUID) (*models.AttendanceSession, error)
	OpenSessionsBefore(ctx context.Context, cutoff time.Time) ([]models.AttendanceSession, error)
}

type attendanceRepo struct{ db *gorm.DB }

// NewAttendanceRepository constructs a Postgres-backed AttendanceRepository.
func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepo{db: db}
}

// base scopes to non-soft-deleted attendance rows. Callers Model() the
// target type when issuing Count or aggregation queries.
func (r *attendanceRepo) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

// preloadSessions inflates the Sessions association in deterministic order
// (check_in ASC) so service code can read row.Sessions[0]/[len-1] as
// first/last without an extra sort.
func preloadSessions(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false).Order("check_in ASC")
}

func (r *attendanceRepo) Create(ctx context.Context, a *models.Attendance) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *attendanceRepo) Update(ctx context.Context, a *models.Attendance) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *attendanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *attendanceRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error) {
	var a models.Attendance
	err := r.base(ctx).
		Preload("Sessions", preloadSessions).
		Preload("Employee").
		Where("id = ?", id).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepo) FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*models.Attendance, error) {
	var a models.Attendance
	err := r.base(ctx).
		Preload("Sessions", preloadSessions).
		Where("employee_id = ? AND date = ?", employeeID, date.Format("2006-01-02")).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepo) List(ctx context.Context, f AttendanceListFilter) ([]models.Attendance, int64, error) {
	// Qualify the soft-delete predicate to attendance.is_deleted — the
	// optional employees join below introduces a second is_deleted column
	// that would otherwise make the unqualified NotDeleted scope
	// ambiguous in Postgres.
	q := r.db.WithContext(ctx).Model(&models.Attendance{}).Where("attendance.is_deleted = ?", false)
	if f.EmployeeID != nil {
		q = q.Where("attendance.employee_id = ?", *f.EmployeeID)
	}
	if f.DepartmentID != nil {
		// Department lives on employees(department_id) per the Go schema
		// split — NOT on users. Join via the employee row.
		q = q.Joins("JOIN employees e ON e.id = attendance.employee_id AND e.is_deleted = false").
			Where("e.department_id = ?", *f.DepartmentID)
	}
	if f.StartDate != nil {
		q = q.Where("attendance.date >= ?", f.StartDate.Format("2006-01-02"))
	}
	if f.EndDate != nil {
		q = q.Where("attendance.date <= ?", f.EndDate.Format("2006-01-02"))
	}
	switch strings.ToLower(f.Status) {
	case "late":
		q = q.Where("attendance.is_late = ?", true)
	case "on_time":
		q = q.Where("attendance.is_late = ?", false)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
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

	var rows []models.Attendance
	err := q.
		Preload("Sessions", preloadSessions).
		Preload("Employee").
		Order("attendance.date DESC, attendance.created_at DESC").
		Limit(size).Offset((page - 1) * size).
		Find(&rows).Error
	return rows, total, err
}

func (r *attendanceRepo) MonthlyCheckInCount(ctx context.Context, employeeID uuid.UUID, year, month int) (int64, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	var n int64
	// Qualify is_deleted with the attendance table — joining
	// attendance_sessions introduces a second is_deleted column and the
	// unqualified NotDeleted scope would raise "ambiguous column reference".
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Joins("JOIN attendance_sessions s ON s.attendance_id = attendance.id AND s.is_deleted = false").
		Where("attendance.is_deleted = ? AND attendance.employee_id = ? AND attendance.date BETWEEN ? AND ?",
			false, employeeID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Distinct("attendance.id").
		Count(&n).Error
	return n, err
}

func (r *attendanceRepo) DatesWithCheckIn(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]time.Time, error) {
	var dates []time.Time
	// Same qualification as MonthlyCheckInCount — see comment there.
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Joins("JOIN attendance_sessions s ON s.attendance_id = attendance.id AND s.is_deleted = false").
		Where("attendance.is_deleted = ? AND attendance.employee_id = ? AND attendance.date BETWEEN ? AND ?",
			false, employeeID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Distinct("attendance.date").
		Order("attendance.date").
		Pluck("attendance.date", &dates).Error
	return dates, err
}

func (r *attendanceRepo) ListForEmployeesInRange(ctx context.Context, employeeIDs []uuid.UUID, from, to time.Time) ([]models.Attendance, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	var rows []models.Attendance
	err := r.base(ctx).
		Preload("Sessions", preloadSessions).
		Where("employee_id IN ? AND date BETWEEN ? AND ?",
			employeeIDs, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&rows).Error
	return rows, err
}

func (r *attendanceRepo) CreateSession(ctx context.Context, s *models.AttendanceSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *attendanceRepo) UpdateSession(ctx context.Context, s *models.AttendanceSession) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// FindOpenSession returns the (at most one) open session for the given
// attendance row, or nil. Defends the partial unique index
// uq_attendance_sessions_one_open by surfacing a clean 409 at the service
// layer rather than letting Postgres raise a unique-violation 500.
func (r *attendanceRepo) FindOpenSession(ctx context.Context, attendanceID uuid.UUID) (*models.AttendanceSession, error) {
	var s models.AttendanceSession
	err := r.base(ctx).
		Where("attendance_id = ? AND check_out IS NULL", attendanceID).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// OpenSessionsBefore returns every open session whose check_in is before
// the cutoff. Used by the (deferred) auto-checkout cron — stubbed here so
// the cron has a wiring point when it's introduced.
func (r *attendanceRepo) OpenSessionsBefore(ctx context.Context, cutoff time.Time) ([]models.AttendanceSession, error) {
	var rows []models.AttendanceSession
	err := r.base(ctx).
		Where("check_out IS NULL AND check_in < ?", cutoff).
		Find(&rows).Error
	return rows, err
}
