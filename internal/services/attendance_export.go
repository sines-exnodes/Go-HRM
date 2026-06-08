package services

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
)

// statusLabel maps a cell status to its short Excel glyph (mirrors Python
// _STATUS_LABEL).
var statusLabel = map[string]string{
	matrixOnTime:         "✓",
	matrixLate:           "L",
	matrixAbsent:         "A",
	matrixWeekend:        "—",
	matrixAnnualLeave:    "AL",
	matrixSickLeave:      "SL",
	matrixPersonalLeave:  "PL",
	matrixMaternityLeave: "ML",
	matrixUnpaidLeave:    "UL",
	matrixHalfDayLeave:   "½",
	matrixNoData:         "",
}

// formatHM renders integer minutes as "Xh Ym" (zero-padded minutes; negatives
// clipped to 0). Mirrors Python _format_hm (SR-011 display format).
func formatHM(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	return fmt.Sprintf("%dh %02dm", minutes/60, minutes%60)
}

// ExportMatrix builds an .xlsx of the monthly attendance matrix. When
// singleEmployeeID is non-nil, only that employee is exported (non-managers may
// export only themselves). Otherwise all visible employees are exported (admin →
// all; non-admin → self). Reuses buildAllRows so the summary columns match the
// on-screen matrix (SR-009 / AC-025).
func (s *AttendanceService) ExportMatrix(
	ctx context.Context,
	currentUserID uuid.UUID,
	asAdmin bool,
	q dto.AttendanceMatrixQuery,
	singleEmployeeID *uuid.UUID,
) ([]byte, error) {
	loc := s.tz()
	now, _ := todayInTZ(loc)
	year, month := q.Year, q.Month
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}

	var employees []models.Employee
	switch {
	case singleEmployeeID != nil:
		target, err := s.emps.FindByID(ctx, *singleEmployeeID)
		if err != nil {
			return nil, apperrors.ErrNotFound("Employee")
		}
		if !asAdmin {
			me, err := s.resolveCurrentEmployee(ctx, currentUserID)
			if err != nil {
				return nil, err
			}
			if me.ID != target.ID {
				return nil, apperrors.ErrForbidden("You do not have permission to export this employee's data")
			}
		}
		employees = []models.Employee{*target}
	case asAdmin:
		rows, _, err := s.emps.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 1000})
		if err != nil {
			return nil, err
		}
		employees = rows
	default:
		me, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		employees = []models.Employee{*me}
	}

	rows, err := s.buildAllRows(ctx, employees, year, month, loc, nil)
	if err != nil {
		return nil, err
	}

	daysInMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc).AddDate(0, 1, -1).Day()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)

	header := []any{"Employee", "Department"}
	for d := 1; d <= daysInMonth; d++ {
		header = append(header, fmt.Sprintf("%d", d))
	}
	header = append(header, "Total Late Time", "Total Early Time")
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}

	for i, row := range rows {
		rec := []any{row.EmployeeName, deref(row.DepartmentName)}
		for d := 1; d <= daysInMonth; d++ {
			rec = append(rec, exportCellLabel(row.Cells[d], loc))
		}
		rec = append(rec, formatHM(row.TotalLateMinutes), formatHM(row.TotalEarlyMinutes))
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &rec); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// exportCellLabel renders one cell as "GLYPH HH:MM-HH:MM" (times in company TZ
// when present), mirroring Python's data-row label build.
func exportCellLabel(cell dto.AttendanceCellRead, loc *time.Location) string {
	label := statusLabel[cell.Status]
	if cell.CheckIn != nil {
		ci := cell.CheckIn.In(loc).Format("15:04")
		if cell.CheckOut != nil {
			co := cell.CheckOut.In(loc).Format("15:04")
			return fmt.Sprintf("%s %s-%s", label, ci, co)
		}
		return fmt.Sprintf("%s %s", label, ci)
	}
	return label
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
