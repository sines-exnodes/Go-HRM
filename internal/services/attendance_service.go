package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// AttendanceService owns the attendance aggregate (attendance +
// attendance_sessions). Permission resolution lives at the route layer
// (RequirePerms) and the handler precomputes an `asAdmin` bool from the
// JWT-preloaded user.Roles — the service enforces ownership using that
// flag plus the resolved current-employee ID. Same shape as LeaveService.
type AttendanceService struct {
	cfg   *config.Config
	repo  repositories.AttendanceRepository
	emps  repositories.EmployeeRepository
	depts repositories.DepartmentRepository
	pos   repositories.PositionRepository
}

// NewAttendanceService constructs an AttendanceService.
func NewAttendanceService(
	cfg *config.Config,
	repo repositories.AttendanceRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
) *AttendanceService {
	return &AttendanceService{cfg: cfg, repo: repo, emps: emps, depts: depts, pos: pos}
}

// ---- shared helpers ----

func (s *AttendanceService) tz() *time.Location {
	return loadTZ(s.cfg.CompanyTimezone)
}

// resolveCurrentEmployee returns the employee row for the authenticated
// user. Missing record (user without an HR profile) yields a 403 — the
// caller cannot record attendance without an employee profile.
func (s *AttendanceService) resolveCurrentEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	emp, err := s.emps.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden("No employee record for current user")
		}
		return nil, err
	}
	return emp, nil
}

// toRead builds the canonical wire shape from a model. Inflates the
// employee/department/position projection when the FKs are live; missing
// refs silently drop to nil (matches LeaveService.populateRead).
func (s *AttendanceService) toRead(ctx context.Context, a *models.Attendance) dto.AttendanceRead {
	out := dto.AttendanceRead{
		ID:           a.ID,
		EmployeeID:   a.EmployeeID,
		Date:         a.Date.Format("2006-01-02"),
		IsLate:       a.IsLate,
		IsHalfDay:    a.IsHalfDay,
		WorkLocation: a.WorkLocation,
		Notes:        a.Notes,
		Sessions:     make([]dto.AttendanceSessionRead, 0, len(a.Sessions)),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
	var totalHours float64
	for _, sess := range a.Sessions {
		hw := hoursBetween(sess.CheckIn, sess.CheckOut)
		out.Sessions = append(out.Sessions, dto.AttendanceSessionRead{
			ID:             sess.ID,
			CheckIn:        sess.CheckIn,
			CheckOut:       sess.CheckOut,
			IsAutoCheckout: sess.IsAutoCheckout,
			HoursWorked:    hw,
		})
		if hw != nil {
			totalHours += *hw
		}
	}
	if len(a.Sessions) > 0 {
		first := a.Sessions[0]
		last := a.Sessions[len(a.Sessions)-1]
		ci := first.CheckIn
		out.CheckIn = &ci
		out.CheckOut = last.CheckOut
		if totalHours > 0 {
			out.HoursWorked = &totalHours
		}
	}
	if a.Employee != nil {
		brief := &dto.AttendanceEmployeeBrief{
			ID:        a.Employee.ID,
			FullName:  a.Employee.FullName,
			AvatarURL: a.Employee.AvatarURL,
		}
		if a.Employee.DepartmentID != nil {
			if d, err := s.depts.FindByID(ctx, *a.Employee.DepartmentID, false); err == nil && d != nil {
				brief.Department = &dto.AttendanceRefRead{ID: d.ID, Name: d.Name}
			}
		}
		if a.Employee.PositionID != nil {
			if p, err := s.pos.FindByID(ctx, *a.Employee.PositionID); err == nil && p != nil {
				brief.Position = &dto.AttendanceRefRead{ID: p.ID, Name: p.Name}
			}
		}
		out.Employee = brief
	}
	return out
}

// ---- Check-in ----

// CheckIn records a check-in for the authenticated employee. First check-in
// of the day creates the attendance row with is_late computed from the
// first check-in vs the configured threshold; subsequent check-ins (after
// a check-out) append a new session WITHOUT re-evaluating is_late
// (REVISION NOTES item #5). Returns 409 if an open session already exists.
func (s *AttendanceService) CheckIn(ctx context.Context, currentUserID uuid.UUID, in dto.AttendanceCheckInReq) (dto.AttendanceRead, error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}

	loc := s.tz()
	nowLocal, todayLocal := todayInTZ(loc)
	when := nowLocal
	if in.CheckIn != nil {
		when = in.CheckIn.In(loc)
		todayLocal = time.Date(when.Year(), when.Month(), when.Day(), 0, 0, 0, 0, loc)
	}

	// GPS validation only when explicitly enabled.
	if s.cfg.OfficeGPSEnabled {
		if in.Latitude == nil || in.Longitude == nil {
			return dto.AttendanceRead{}, apperrors.ErrBadRequest("GPS coordinates required")
		}
		if in.Accuracy != nil && *in.Accuracy > s.cfg.OfficeRadiusMeters {
			return dto.AttendanceRead{}, apperrors.ErrBadRequest(
				fmt.Sprintf("GPS accuracy (%dm) is too low — must be within %dm", int(*in.Accuracy), int(s.cfg.OfficeRadiusMeters)))
		}
		dist := haversineMeters(*in.Latitude, *in.Longitude, s.cfg.OfficeLatitude, s.cfg.OfficeLongitude)
		if dist > s.cfg.OfficeRadiusMeters {
			return dto.AttendanceRead{}, apperrors.ErrBadRequest(
				fmt.Sprintf("You are %dm from the office (limit %dm)", int(dist), int(s.cfg.OfficeRadiusMeters)))
		}
	}

	existing, err := s.repo.FindByEmployeeAndDate(ctx, currentEmp.ID, todayLocal)
	if err != nil {
		return dto.AttendanceRead{}, err
	}

	if existing != nil {
		// Defend the partial unique index — surface a clean 409 instead of
		// letting Postgres raise a unique-violation 500.
		open, err := s.repo.FindOpenSession(ctx, existing.ID)
		if err != nil {
			return dto.AttendanceRead{}, err
		}
		if open != nil {
			return dto.AttendanceRead{}, apperrors.ErrConflict("You are already checked in")
		}
		// Append a new session without recomputing is_late.
		sess := &models.AttendanceSession{
			AttendanceID: existing.ID,
			CheckIn:      when.UTC(),
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
		reloaded, err := s.repo.FindByID(ctx, existing.ID)
		if err != nil {
			return dto.AttendanceRead{}, err
		}
		return s.toRead(ctx, reloaded), nil
	}

	// First check-in of the day → create the attendance row + first session.
	lateAt := thresholdAt(when, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
	isLate := when.After(lateAt)

	row := &models.Attendance{
		EmployeeID:   currentEmp.ID,
		Date:         truncateToDateInTZ(when, loc),
		IsLate:       isLate,
		WorkLocation: in.WorkLocation,
		Notes:        in.Notes,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	sess := &models.AttendanceSession{
		AttendanceID: row.ID,
		CheckIn:      when.UTC(),
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return dto.AttendanceRead{}, err
	}
	reloaded, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(ctx, reloaded), nil
}

// ---- Check-out ----

// CheckOut closes the (single) open session for today. Returns 409 when
// no open session exists. After close, total hours-worked is summed across
// all sessions and is_half_day is flipped when the day total is below the
// configured threshold.
func (s *AttendanceService) CheckOut(ctx context.Context, currentUserID uuid.UUID, in dto.AttendanceCheckOutReq) (dto.AttendanceRead, error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}

	loc := s.tz()
	nowLocal, todayLocal := todayInTZ(loc)
	when := nowLocal
	if in.CheckOut != nil {
		when = in.CheckOut.In(loc)
		todayLocal = time.Date(when.Year(), when.Month(), when.Day(), 0, 0, 0, 0, loc)
	}

	row, err := s.repo.FindByEmployeeAndDate(ctx, currentEmp.ID, todayLocal)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	if row == nil {
		return dto.AttendanceRead{}, apperrors.ErrBadRequest("No check-in found for today")
	}

	open, err := s.repo.FindOpenSession(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	if open == nil {
		return dto.AttendanceRead{}, apperrors.ErrConflict("You are not currently checked in")
	}

	whenUTC := when.UTC()
	if whenUTC.Before(open.CheckIn) {
		return dto.AttendanceRead{}, apperrors.ErrBadRequest("Check-out time cannot be before check-in")
	}
	open.CheckOut = &whenUTC
	if err := s.repo.UpdateSession(ctx, open); err != nil {
		return dto.AttendanceRead{}, err
	}

	reloaded, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	var total float64
	for _, sess := range reloaded.Sessions {
		if hw := hoursBetween(sess.CheckIn, sess.CheckOut); hw != nil {
			total += *hw
		}
	}
	if total > 0 && total < s.cfg.HalfDayHoursThreshold {
		reloaded.IsHalfDay = true
	}
	if in.Notes != nil {
		reloaded.Notes = in.Notes
	}
	if err := s.repo.Update(ctx, reloaded); err != nil {
		return dto.AttendanceRead{}, err
	}

	final, err := s.repo.FindByID(ctx, reloaded.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(ctx, final), nil
}

// ---- Today status ----

// Today returns the current employee's attendance status for today plus a
// monthly check-in count and consecutive-workday streak. Status enum:
// not_checked_in | checked_in | checked_out.
func (s *AttendanceService) Today(ctx context.Context, currentUserID uuid.UUID) (dto.TodayStatusRead, error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return dto.TodayStatusRead{}, err
	}

	loc := s.tz()
	_, todayLocal := todayInTZ(loc)

	row, err := s.repo.FindByEmployeeAndDate(ctx, currentEmp.ID, todayLocal)
	if err != nil {
		return dto.TodayStatusRead{}, err
	}

	out := dto.TodayStatusRead{Status: "not_checked_in", Sessions: []dto.AttendanceSessionRead{}}
	if row != nil && len(row.Sessions) > 0 {
		out.IsLate = row.IsLate
		for _, sess := range row.Sessions {
			out.Sessions = append(out.Sessions, dto.AttendanceSessionRead{
				ID:             sess.ID,
				CheckIn:        sess.CheckIn,
				CheckOut:       sess.CheckOut,
				IsAutoCheckout: sess.IsAutoCheckout,
				HoursWorked:    hoursBetween(sess.CheckIn, sess.CheckOut),
			})
		}
		last := row.Sessions[len(row.Sessions)-1]
		if last.CheckOut == nil {
			out.Status = "checked_in"
			ci := last.CheckIn
			out.CurrentCheckIn = &ci
		} else {
			out.Status = "checked_out"
		}
	}

	cnt, err := s.repo.MonthlyCheckInCount(ctx, currentEmp.ID, todayLocal.Year(), int(todayLocal.Month()))
	if err != nil {
		return dto.TodayStatusRead{}, err
	}
	out.MonthlyCount = int(cnt)

	from := todayLocal.AddDate(-1, 0, 0)
	dates, err := s.repo.DatesWithCheckIn(ctx, currentEmp.ID, from, todayLocal)
	if err != nil {
		return dto.TodayStatusRead{}, err
	}
	hit := make(map[string]struct{}, len(dates))
	for _, d := range dates {
		hit[d.Format("2006-01-02")] = struct{}{}
	}
	// Walk backward from today over workdays. If today is itself a workday
	// with no check-in, skip it (don't penalize before end-of-day) and
	// count consecutive prior workdays.
	streak := 0
	d := todayLocal
	if isWorkday(d) {
		if _, ok := hit[d.Format("2006-01-02")]; !ok {
			d = d.AddDate(0, 0, -1)
		}
	}
	for d.After(from) || d.Equal(from) {
		if isWorkday(d) {
			if _, ok := hit[d.Format("2006-01-02")]; ok {
				streak++
			} else {
				break
			}
		}
		d = d.AddDate(0, 0, -1)
	}
	out.Streak = streak
	return out, nil
}

// ---- Get by ID ----

// Get returns a single attendance row. Owner or admin only; everyone else
// gets a 403. Route-level PermAttendanceRead is applied upstream.
func (s *AttendanceService) Get(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (dto.AttendanceRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AttendanceRead{}, apperrors.ErrNotFound("Attendance")
		}
		return dto.AttendanceRead{}, err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return dto.AttendanceRead{}, err
		}
		if row.EmployeeID != currentEmp.ID {
			return dto.AttendanceRead{}, apperrors.ErrForbidden("You do not own this attendance row")
		}
	}
	return s.toRead(ctx, row), nil
}

// ---- List ----

// List returns paginated attendance rows. Route-level PermAttendanceRead
// is applied upstream. Non-admin (no PermAttendanceManage) is silently
// scoped to own employee_id (matches the Python contract: managers see
// all; non-managers see only their own row).
func (s *AttendanceService) List(ctx context.Context, currentUserID uuid.UUID, asAdmin bool, q dto.AttendanceListQuery) (dto.PaginatedData[dto.AttendanceRead], error) {
	loc := s.tz()
	f := repositories.AttendanceListFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Status:   q.Status,
	}
	if q.StartDate != "" {
		t, err := parseDateYMD(q.StartDate, loc)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperrors.ErrBadRequest("invalid start_date")
		}
		f.StartDate = &t
	}
	if q.EndDate != "" {
		t, err := parseDateYMD(q.EndDate, loc)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperrors.ErrBadRequest("invalid end_date")
		}
		f.EndDate = &t
	}
	if q.EmployeeID != "" {
		eid, err := uuid.Parse(q.EmployeeID)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperrors.ErrBadRequest("invalid employee_id")
		}
		f.EmployeeID = &eid
	}
	if q.DepartmentID != "" {
		did, err := uuid.Parse(q.DepartmentID)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperrors.ErrBadRequest("invalid department_id")
		}
		f.DepartmentID = &did
	}

	// Non-admin: force scope to own employee. An explicit employee_id
	// query param from a non-admin is silently overridden (don't 403 —
	// matches the Python contract).
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, err
		}
		me := currentEmp.ID
		f.EmployeeID = &me
	}

	rows, total, err := s.repo.List(ctx, f)
	if err != nil {
		return dto.PaginatedData[dto.AttendanceRead]{}, err
	}
	items := make([]dto.AttendanceRead, 0, len(rows))
	for i := range rows {
		items = append(items, s.toRead(ctx, &rows[i]))
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
	return dto.PaginatedData[dto.AttendanceRead]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
