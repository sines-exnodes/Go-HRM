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
)

// ---------- Admin: create ----------

// AdminCreate inserts an attendance row on behalf of an employee. Conflict
// (409) when (employee_id, date) already exists — the partial unique
// constraint is defended at the service level by FindByEmployeeAndDate.
// When CheckIn is provided and IsLate is nil, is_late is auto-derived
// from CheckIn vs the configured threshold.
func (s *AttendanceService) AdminCreate(ctx context.Context, in dto.AttendanceAdminCreateReq) (dto.AttendanceRead, error) {
	loc := s.tz()
	day, err := parseDateYMD(in.Date, loc)
	if err != nil {
		return dto.AttendanceRead{}, apperrors.ErrBadRequest("invalid date (expected YYYY-MM-DD)")
	}

	// Verify the subject employee exists.
	if _, err := s.emps.FindByID(ctx, in.EmployeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AttendanceRead{}, apperrors.ErrNotFound("Employee")
		}
		return dto.AttendanceRead{}, err
	}

	existing, err := s.repo.FindByEmployeeAndDate(ctx, in.EmployeeID, day)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	if existing != nil {
		return dto.AttendanceRead{}, apperrors.ErrConflict("Attendance for this employee/date already exists")
	}

	row := &models.Attendance{
		EmployeeID:   in.EmployeeID,
		Date:         day,
		WorkLocation: in.WorkLocation,
		Notes:        in.Notes,
	}
	if in.IsLate != nil {
		row.IsLate = *in.IsLate
	}
	if in.IsHalfDay != nil {
		row.IsHalfDay = *in.IsHalfDay
	}
	// Auto-derive is_late when CheckIn provided and not explicitly set.
	if in.CheckIn != nil && in.IsLate == nil {
		ci := in.CheckIn.In(loc)
		lateAt := thresholdAt(ci, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
		row.IsLate = ci.After(lateAt)
	}

	if err := s.repo.Create(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	if in.CheckIn != nil {
		sess := &models.AttendanceSession{
			AttendanceID: row.ID,
			CheckIn:      in.CheckIn.UTC(),
		}
		if in.CheckOut != nil {
			co := in.CheckOut.UTC()
			sess.CheckOut = &co
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
	}

	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(ctx, final), nil
}

// ---------- Admin: update ----------

// AdminUpdate patches an existing attendance row. Pointer fields preserve
// "not provided" semantics. CheckIn/CheckOut adjust the FIRST session's
// times when one exists; otherwise (CheckIn provided, no sessions yet) a
// new session is appended.
func (s *AttendanceService) AdminUpdate(ctx context.Context, id uuid.UUID, in dto.AttendanceAdminUpdateReq) (dto.AttendanceRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AttendanceRead{}, apperrors.ErrNotFound("Attendance")
		}
		return dto.AttendanceRead{}, err
	}

	if in.IsLate != nil {
		row.IsLate = *in.IsLate
	}
	if in.IsHalfDay != nil {
		row.IsHalfDay = *in.IsHalfDay
	}
	if in.WorkLocation != nil {
		row.WorkLocation = in.WorkLocation
	}
	if in.Notes != nil {
		row.Notes = in.Notes
	}

	if (in.CheckIn != nil || in.CheckOut != nil) && len(row.Sessions) > 0 {
		first := row.Sessions[0]
		if in.CheckIn != nil {
			first.CheckIn = in.CheckIn.UTC()
		}
		if in.CheckOut != nil {
			co := in.CheckOut.UTC()
			first.CheckOut = &co
		}
		if err := s.repo.UpdateSession(ctx, &first); err != nil {
			return dto.AttendanceRead{}, err
		}
	} else if in.CheckIn != nil {
		sess := &models.AttendanceSession{
			AttendanceID: row.ID,
			CheckIn:      in.CheckIn.UTC(),
		}
		if in.CheckOut != nil {
			co := in.CheckOut.UTC()
			sess.CheckOut = &co
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
	}

	if err := s.repo.Update(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(ctx, final), nil
}

// ---------- Admin: delete ----------

// AdminDelete soft-deletes an attendance row. Sessions hard-delete via
// ON DELETE CASCADE when the row is eventually purged; for now they
// remain in the table but are excluded from every read via the
// NotDeleted scope applied at the parent row.
func (s *AttendanceService) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Attendance")
		}
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

// ---------- Matrix ----------

// Matrix cell status enum.
const (
	matrixOnTime         = "on_time"
	matrixLate           = "late"
	matrixAbsent         = "absent"
	matrixWeekend        = "weekend"
	matrixNoData         = "no_data"
	matrixAnnualLeave    = "annual_leave"
	matrixSickLeave      = "sick_leave"
	matrixPersonalLeave  = "personal_leave"
	matrixMaternityLeave = "maternity_leave"
	matrixUnpaidLeave    = "unpaid_leave"
	matrixHalfDayLeave   = "half_day_leave"
)

// Half-day boundary constants — BA-fixed (DR-004-001-01 SR-002/SR-011 v1.2),
// mirror Python's _AM_END / _PM_LATE. AM late + workday-end thresholds come
// from config (LateThreshold / CheckoutThreshold).
const (
	amHalfEndHour = 12 // end of the AM half (early-leave boundary when AM worked + PM on leave)
	amHalfEndMin  = 0
	pmLateHour    = 13 // PM-half late threshold (when AM on leave + PM worked)
	pmLateMin     = 15
)

// leaveTypeToStatus maps an approved full-day leave type to its cell status.
var leaveTypeToStatus = map[models.LeaveType]string{
	models.LeaveTypeAnnual:    matrixAnnualLeave,
	models.LeaveTypeSick:      matrixSickLeave,
	models.LeaveTypePersonal:  matrixPersonalLeave,
	models.LeaveTypeMaternity: matrixMaternityLeave,
	models.LeaveTypeUnpaid:    matrixUnpaidLeave,
}

// onLeaveStatuses is the set used by the on_leave status filter.
var onLeaveStatuses = map[string]struct{}{
	matrixAnnualLeave: {}, matrixSickLeave: {}, matrixPersonalLeave: {},
	matrixMaternityLeave: {}, matrixUnpaidLeave: {}, matrixHalfDayLeave: {},
}

// applyLeaveCell overlays an approved leave onto a cell. Full-day leave maps
// to its type's status; half-day leave maps to half_day_leave and the caller
// is responsible for attaching any worked-half record + WorkedHalfStatus.
func applyLeaveCell(cell *dto.AttendanceCellRead, lv models.LeaveRequest) {
	lt := string(lv.LeaveType)
	lp := string(lv.LeavePeriod)
	cell.LeaveType = &lt
	cell.LeavePeriod = &lp
	if models.IsHalfDayPeriod(lv.LeavePeriod) {
		cell.Status = matrixHalfDayLeave
		return
	}
	if st, ok := leaveTypeToStatus[lv.LeaveType]; ok {
		cell.Status = st
	} else {
		cell.Status = matrixAnnualLeave
	}
}

// attachRecordToCell projects a worked attendance record (check-in/out, hours,
// sessions) onto a cell. Used both for plain worked days and for the worked
// half of a combined half-day-leave cell.
func attachRecordToCell(cell *dto.AttendanceCellRead, rec models.Attendance) {
	f := rec.Sessions[0]
	l := rec.Sessions[len(rec.Sessions)-1]
	ci := f.CheckIn
	cell.CheckIn = &ci
	cell.CheckOut = l.CheckOut
	cell.IsLate = rec.IsLate
	var total float64
	sessions := make([]dto.AttendanceSessionRead, 0, len(rec.Sessions))
	for _, sess := range rec.Sessions {
		hw := hoursBetween(sess.CheckIn, sess.CheckOut)
		sessions = append(sessions, dto.AttendanceSessionRead{
			ID: sess.ID, CheckIn: sess.CheckIn, CheckOut: sess.CheckOut,
			IsAutoCheckout: sess.IsAutoCheckout, HoursWorked: hw,
		})
		if hw != nil {
			total += *hw
		}
	}
	if total > 0 {
		cell.HoursWorked = &total
	}
	cell.Sessions = sessions
}

// computeWorkedHalfStatus derives on_time | late | absent for the worked half
// of a combined half-day-leave cell. When AM is on leave (worked PM) the PM
// late boundary (pmLateHour:pmLateMin) applies; otherwise the configured AM
// late threshold applies.
func (s *AttendanceService) computeWorkedHalfStatus(cell *dto.AttendanceCellRead, lv models.LeaveRequest, loc *time.Location) string {
	if cell.CheckIn == nil {
		return matrixAbsent
	}
	firstLocal := cell.CheckIn.In(loc)
	var threshold time.Time
	if lv.LeavePeriod == models.LeavePeriodMorningHalf {
		threshold = thresholdAt(firstLocal, pmLateHour, pmLateMin)
	} else {
		threshold = thresholdAt(firstLocal, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
	}
	if firstLocal.After(threshold) {
		return matrixLate
	}
	return matrixOnTime
}

// accumulateSummary returns the (late, early) minute contribution of a single
// cell to the row summary. Leave statuses contribute nothing; half-day-leave
// cells contribute only for their worked half using the SR-011 boundaries.
func (s *AttendanceService) accumulateSummary(cell dto.AttendanceCellRead, loc *time.Location) (int, int) {
	switch cell.Status {
	case matrixWeekend, matrixAbsent, matrixNoData,
		matrixAnnualLeave, matrixSickLeave, matrixPersonalLeave,
		matrixMaternityLeave, matrixUnpaidLeave:
		return 0, 0
	}
	if cell.Status == matrixHalfDayLeave && cell.CheckIn == nil {
		return 0, 0
	}
	lateHour, lateMin := s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute
	earlyHour, earlyMin := s.cfg.CheckoutThresholdHour, s.cfg.CheckoutThresholdMinute
	if cell.Status == matrixHalfDayLeave && cell.LeavePeriod != nil {
		switch *cell.LeavePeriod {
		case string(models.LeavePeriodMorningHalf):
			lateHour, lateMin = pmLateHour, pmLateMin
		case string(models.LeavePeriodAfternoonHalf):
			earlyHour, earlyMin = amHalfEndHour, amHalfEndMin
		}
	}
	var lateAdd, earlyAdd int
	if cell.CheckIn != nil {
		ci := cell.CheckIn.In(loc)
		ref := thresholdAt(ci, lateHour, lateMin)
		if ci.After(ref) {
			lateAdd = int(ci.Sub(ref).Minutes())
		}
	}
	if cell.CheckOut != nil {
		co := cell.CheckOut.In(loc)
		ref := thresholdAt(co, earlyHour, earlyMin)
		if co.Before(ref) {
			earlyAdd = int(ref.Sub(co).Minutes())
		}
	}
	return lateAdd, earlyAdd
}

// Matrix returns the monthly attendance matrix. Managers (asAdmin) see
// every employee filtered by Search + DepartmentID; non-managers see only
// their own row. Cells are keyed by day-of-month. Weekends, no-data
// future days, and absent past days are distinguished.
func (s *AttendanceService) Matrix(ctx context.Context, currentUserID uuid.UUID, asAdmin bool, q dto.AttendanceMatrixQuery) (dto.AttendanceMatrixRead, error) {
	loc := s.tz()
	now, _ := todayInTZ(loc)

	year := q.Year
	month := q.Month
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 20
	}

	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, loc).Day()

	var employees []models.Employee
	if asAdmin {
		empQuery := dto.EmployeeListQuery{
			Page:     1,
			PageSize: 1000,
			Search:   strings.TrimSpace(q.Search),
		}
		if q.DepartmentID != "" {
			did, err := uuid.Parse(q.DepartmentID)
			if err != nil {
				return dto.AttendanceMatrixRead{}, apperrors.ErrBadRequest("invalid department_id")
			}
			empQuery.DepartmentIDs = []uuid.UUID{did}
		}
		rows, _, err := s.emps.List(ctx, empQuery)
		if err != nil {
			return dto.AttendanceMatrixRead{}, err
		}
		employees = rows
	} else {
		// Own row only. Reject the request when the user has no HR profile
		// — there's nothing to render.
		me, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return dto.AttendanceMatrixRead{}, err
		}
		emp, err := s.emps.FindByID(ctx, me.ID)
		if err != nil {
			return dto.AttendanceMatrixRead{}, err
		}
		employees = []models.Employee{*emp}
	}

	rows, err := s.buildAllRows(ctx, employees, year, month, loc, parseCSVSet(q.Status))
	if err != nil {
		return dto.AttendanceMatrixRead{}, err
	}

	total := len(rows)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	pageRows := rows[start:end]
	totalPages := 0
	if total > 0 {
		totalPages = (total + size - 1) / size
	}

	return dto.AttendanceMatrixRead{
		Year:        year,
		Month:       month,
		DaysInMonth: daysInMonth,
		Items:       pageRows,
		Total:       total,
		Page:        page,
		PageSize:    size,
		TotalPages:  totalPages,
	}, nil
}

// buildAllRows constructs every employee's matrix row (unpaginated) for the
// given month. statusSet filters rows (nil = no filter). Shared by Matrix
// (which paginates) and the Excel export (which writes all rows).
func (s *AttendanceService) buildAllRows(
	ctx context.Context,
	employees []models.Employee,
	year, month int,
	loc *time.Location,
	statusSet map[string]struct{},
) ([]dto.AttendanceRowRead, error) {
	now, _ := todayInTZ(loc)
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	last := first.AddDate(0, 1, 0).Add(-time.Second)
	daysInMonth := last.Day()

	ids := make([]uuid.UUID, 0, len(employees))
	for _, e := range employees {
		ids = append(ids, e.ID)
	}
	records, err := s.repo.ListForEmployeesInRange(ctx, ids, first, last)
	if err != nil {
		return nil, err
	}
	byEmp := make(map[uuid.UUID]map[string]models.Attendance, len(employees))
	for _, r := range records {
		m, ok := byEmp[r.EmployeeID]
		if !ok {
			m = make(map[string]models.Attendance)
			byEmp[r.EmployeeID] = m
		}
		m[r.Date.Format("2006-01-02")] = r
	}

	// Approved leave overlay: keyed by employee → "YYYY-MM-DD" → the leave row.
	// The earliest-inserted leave wins on overlap (first-write semantics).
	leaves, err := s.leaves.ApprovedForEmployeesInRange(ctx, ids, first, last)
	if err != nil {
		return nil, err
	}
	leaveByEmp := make(map[uuid.UUID]map[string]models.LeaveRequest, len(employees))
	for _, lv := range leaves {
		m, ok := leaveByEmp[lv.EmployeeID]
		if !ok {
			m = make(map[string]models.LeaveRequest)
			leaveByEmp[lv.EmployeeID] = m
		}
		for d := lv.FromDate; !d.After(lv.ToDate); d = d.AddDate(0, 0, 1) {
			key := d.Format("2006-01-02")
			if _, exists := m[key]; !exists {
				m[key] = lv
			}
		}
	}

	rows := make([]dto.AttendanceRowRead, 0, len(employees))

	for _, emp := range employees {
		cells := make(map[int]dto.AttendanceCellRead, daysInMonth)
		empRecs := byEmp[emp.ID]
		var totalLate, totalEarly int
		cellStatusUnion := make(map[string]struct{}, 8)
		workedHalfUnion := make(map[string]struct{}, 4)

		for d := 1; d <= daysInMonth; d++ {
			day := time.Date(year, time.Month(month), d, 0, 0, 0, 0, loc)
			cell := dto.AttendanceCellRead{
				Date: day.Format("2006-01-02"),
				Day:  d,
			}
			// Cell precedence: weekend → approved leave → attendance record →
			// absent/no_data.
			switch {
			case day.Weekday() == time.Saturday || day.Weekday() == time.Sunday:
				cell.Status = matrixWeekend
			default:
				key := day.Format("2006-01-02")
				if lv, onLeave := leaveByEmp[emp.ID][key]; onLeave {
					applyLeaveCell(&cell, lv)
					if models.IsHalfDayPeriod(lv.LeavePeriod) {
						if rec, ok := empRecs[key]; ok && len(rec.Sessions) > 0 {
							attachRecordToCell(&cell, rec)
						}
						whs := s.computeWorkedHalfStatus(&cell, lv, loc)
						cell.WorkedHalfStatus = &whs
					}
				} else if rec, ok := empRecs[key]; ok && len(rec.Sessions) > 0 {
					attachRecordToCell(&cell, rec)
					if rec.IsLate {
						cell.Status = matrixLate
					} else {
						cell.Status = matrixOnTime
					}
				} else if day.Before(now) {
					cell.Status = matrixAbsent
				} else {
					cell.Status = matrixNoData
				}
			}
			cells[d] = cell
			cellStatusUnion[cell.Status] = struct{}{}
			if cell.WorkedHalfStatus != nil {
				workedHalfUnion[*cell.WorkedHalfStatus] = struct{}{}
			}
			la, ea := s.accumulateSummary(cell, loc)
			totalLate += la
			totalEarly += ea
		}

		// Status CSV filter: drop the row when none of its cells match. The
		// special "on_leave" token matches any leave status; otherwise a token
		// matches either a cell status or a worked-half status (combined cells).
		if statusSet != nil {
			matched := false
			for sf := range statusSet {
				switch sf {
				case "on_leave":
					for k := range cellStatusUnion {
						if _, ok := onLeaveStatuses[k]; ok {
							matched = true
						}
					}
				default:
					if _, ok := cellStatusUnion[sf]; ok {
						matched = true
					}
					if _, ok := workedHalfUnion[sf]; ok {
						matched = true
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		row := dto.AttendanceRowRead{
			EmployeeID:        emp.ID,
			EmployeeName:      emp.FullName(),
			AvatarURL:         emp.AvatarURL,
			Cells:             cells,
			TotalLateMinutes:  totalLate,
			TotalEarlyMinutes: totalEarly,
		}
		if emp.DepartmentID != nil {
			if d, err := s.depts.FindByID(ctx, *emp.DepartmentID, false); err == nil && d != nil {
				row.DepartmentName = &d.Name
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// parseCSVSet splits a comma-separated string into a set. Returns nil when
// the input is empty so callers can short-circuit the "no filter" branch.
func parseCSVSet(s string) map[string]struct{} {
	if s == "" {
		return nil
	}
	out := make(map[string]struct{}, 4)
	for _, raw := range strings.Split(s, ",") {
		v := strings.TrimSpace(raw)
		if v != "" {
			out[v] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
