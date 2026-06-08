package services_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
)

func TestAttendance_Export_BulkProducesXlsx(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "exp-mgr@example.com", "Mgr")
	makeEmpUser(t, "exp-e1@example.com", "E1")

	data, err := svc.ExportMatrix(ctx, mgr.ID, true, dto.AttendanceMatrixQuery{Month: 5, Year: 2026}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetName(0))
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	header := rows[0]
	assert.Equal(t, "Employee", header[0])
	assert.Equal(t, "Department", header[1])
	assert.Contains(t, header, "Total Late Time")
	assert.Contains(t, header, "Total Early Time")
}

func TestAttendance_Export_NonManagerSingle_DeniesOther(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newAttendanceSvc(t)
	alice, _ := makeEmpUser(t, "exp-alice@example.com", "Alice")
	_, bobEmp := makeEmpUser(t, "exp-bob@example.com", "Bob")

	_, err := svc.ExportMatrix(ctx, alice.ID, false, dto.AttendanceMatrixQuery{Month: 5, Year: 2026}, &bobEmp.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}
