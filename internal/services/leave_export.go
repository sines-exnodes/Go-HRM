package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
)

// ExportLeave builds an .xlsx of leave requests matching the query.
// All filters in q are applied (search, status, department, position).
// Mirrors Python leave export (G10 from 2026-06-08 parity audit).
func (s *LeaveService) ExportLeave(
	ctx context.Context,
	q dto.LeaveListQuery,
) ([]byte, error) {
	q.Page = 1
	q.PageSize = 5000

	result, err := s.List(ctx, q)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)

	header := []any{
		"Employee", "Department", "Position",
		"Leave Type", "From Date", "To Date", "Period", "Total Days",
		"Reason", "Status", "Created At",
	}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}

	for i, lr := range result.Items {
		empName, deptName, posName := "", "", ""
		if lr.Employee != nil {
			empName = lr.Employee.Name
		}
		if lr.Department != nil {
			deptName = lr.Department.Name
		}
		if lr.Position != nil {
			posName = lr.Position.Name
		}
		row := []any{
			empName,
			deptName,
			posName,
			string(lr.LeaveType),
			lr.FromDate.Format("2006-01-02"),
			lr.ToDate.Format("2006-01-02"),
			string(lr.LeavePeriod),
			fmt.Sprintf("%.1f", lr.TotalDays),
			lr.Reason,
			string(lr.Status),
			lr.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &row); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
