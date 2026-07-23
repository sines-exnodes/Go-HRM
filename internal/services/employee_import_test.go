package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

func csvBytes(header string, rows ...string) []byte {
	var b strings.Builder
	b.WriteString(header)
	b.WriteByte('\n')
	for _, r := range rows {
		b.WriteString(r)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// TestImportCSV_CreatesValidRows: two well-formed rows must both land as
// employees. Import is create-only; partial success depends on per-row Create.
func TestImportCSV_CreatesValidRows(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	file := csvBytes(
		"email,first_name,last_name",
		"alice@example.com,Alice,Smith",
		"bob@example.com,Bob,Jones",
	)
	out, err := svc.ImportCSV(ctx, file, services.AllEmployeeFieldPerms)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 2, out.TotalRows)
	assert.Equal(t, 2, out.Created)
	assert.Equal(t, 0, out.Failed)
	require.Len(t, out.Results, 2)
	assert.True(t, out.Results[0].OK)
	assert.Equal(t, 2, out.Results[0].Row)
	assert.Equal(t, "alice@example.com", out.Results[0].Email)
	assert.NotNil(t, out.Results[0].EmployeeID)
	assert.NotNil(t, out.Results[0].UserID)
	assert.True(t, out.Results[1].OK)
	assert.Equal(t, 3, out.Results[1].Row)

	var n int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM employees").Scan(&n).Error)
	assert.Equal(t, int64(2), n)
}

// TestImportCSV_DuplicateEmail_IsRowError: a pre-existing email must not abort
// the rest of the file — partial success is the product contract (D1/D12).
func TestImportCSV_DuplicateEmail_IsRowError(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	u := makeUser(t, "dup@example.com", "Pass12345")
	_ = makeEmployee(t, u, "Dup User")

	file := csvBytes(
		"email,first_name,last_name",
		"dup@example.com,Dup,Again",
		"fresh@example.com,Fresh,Person",
	)
	out, err := svc.ImportCSV(ctx, file, services.AllEmployeeFieldPerms)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 2, out.TotalRows)
	assert.Equal(t, 1, out.Created)
	assert.Equal(t, 1, out.Failed)
	require.Len(t, out.Results, 2)
	assert.False(t, out.Results[0].OK)
	assert.Contains(t, out.Results[0].Error, "already exists")
	assert.True(t, out.Results[1].OK)
	assert.Equal(t, "fresh@example.com", out.Results[1].Email)
}

// TestImportCSV_MissingRequiredColumn_WholeFile400: missing email header is a
// whole-file problem, not a per-row problem — fail loud so typos aren't silent.
func TestImportCSV_MissingRequiredColumn_WholeFile400(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)

	file := csvBytes(
		"first_name,last_name",
		"Alice,Smith",
	)
	out, err := svc.ImportCSV(context.Background(), file, services.AllEmployeeFieldPerms)
	require.Error(t, err)
	assert.Nil(t, out)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
	assert.Contains(t, strings.ToLower(ae.Message), "email")
}

// TestImportCSV_UnknownHeader_WholeFile400: unknown columns must reject the
// whole file so a typo never silently drops data.
func TestImportCSV_UnknownHeader_WholeFile400(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)

	file := csvBytes(
		"email,first_name,last_name,not_a_real_column",
		"a@example.com,A,B,x",
	)
	out, err := svc.ImportCSV(context.Background(), file, services.AllEmployeeFieldPerms)
	require.Error(t, err)
	assert.Nil(t, out)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
	assert.Contains(t, strings.ToLower(ae.Message), "unknown")
}

// TestImportCSV_UnknownDepartment_RowError: a bad department name is a row
// error; a sibling good row with a real department still creates.
func TestImportCSV_UnknownDepartment_RowError(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	deptID, _ := makeOrg(t, "Engineering", "Engineer")
	_ = deptID

	file := csvBytes(
		"email,first_name,last_name,department",
		"good@example.com,Good,Emp,Engineering",
		"bad@example.com,Bad,Emp,No Such Dept",
	)
	out, err := svc.ImportCSV(ctx, file, services.AllEmployeeFieldPerms)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 2, out.TotalRows)
	assert.Equal(t, 1, out.Created)
	assert.Equal(t, 1, out.Failed)

	var goodOK, badOK bool
	for _, r := range out.Results {
		switch r.Email {
		case "good@example.com":
			goodOK = r.OK
			assert.True(t, r.OK)
		case "bad@example.com":
			badOK = !r.OK
			assert.False(t, r.OK)
			assert.Contains(t, r.Error, "unknown department")
		}
	}
	assert.True(t, goodOK)
	assert.True(t, badOK)
}

// TestImportCSV_ResolvesRoleAndManager: role name and manager_email resolve to
// real FKs so CSV authors never need UUIDs (D5). Manager must already exist (D6).
func TestImportCSV_ResolvesRoleAndManager(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	role := makeRole(t, "Manager", []permissions.Permission{permissions.PermAuthLogin}, false)
	mgrUser := makeUser(t, "mgr@example.com", "Pass12345")
	mgrEmp := makeEmployee(t, mgrUser, "Mgr Boss")

	file := csvBytes(
		"email,first_name,last_name,role,manager_email",
		"report@example.com,Report,One,Manager,mgr@example.com",
	)
	out, err := svc.ImportCSV(ctx, file, services.AllEmployeeFieldPerms)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 1, out.Created)
	assert.Equal(t, 0, out.Failed)
	require.Len(t, out.Results, 1)
	require.True(t, out.Results[0].OK)
	require.NotNil(t, out.Results[0].EmployeeID)

	got, err := svc.Get(ctx, *out.Results[0].EmployeeID)
	require.NoError(t, err)
	require.NotNil(t, got.Manager)
	assert.Equal(t, mgrEmp.ID, got.Manager.ID)

	// Role resolved onto the user
	var roleCount int64
	require.NoError(t, testDB.Raw(
		`SELECT count(*) FROM user_roles ur
		 JOIN roles r ON r.id = ur.role_id
		 WHERE ur.user_id = ? AND r.id = ? AND ur.is_deleted = false`,
		got.UserID, role.ID,
	).Scan(&roleCount).Error)
	assert.Equal(t, int64(1), roleCount)
}

// TestImportCSV_EmptyFile_400: empty body is a whole-file 400, not a silent
// success with zero rows.
func TestImportCSV_EmptyFile_400(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)

	out, err := svc.ImportCSV(context.Background(), []byte{}, services.AllEmployeeFieldPerms)
	require.Error(t, err)
	assert.Nil(t, out)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Header-only is also empty of data rows.
	out, err = svc.ImportCSV(context.Background(), csvBytes("email,first_name,last_name"), services.AllEmployeeFieldPerms)
	require.Error(t, err)
	assert.Nil(t, out)
	ae, ok = apperrors.As(err)
	require.True(t, ok)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

// TestImportCSV_SalaryWithoutPerm_RowError: salary cells without salary_manage
// must fail the row (D9), not strip silently — admins need to know the write was refused.
func TestImportCSV_SalaryWithoutPerm_RowError(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)

	file := csvBytes(
		"email,first_name,last_name,basic_salary",
		"paid@example.com,Paid,Person,5000000",
	)
	out, err := svc.ImportCSV(context.Background(), file, services.EmployeeFieldPerms{})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 0, out.Created)
	assert.Equal(t, 1, out.Failed)
	require.Len(t, out.Results, 1)
	assert.False(t, out.Results[0].OK)
	assert.Contains(t, strings.ToLower(out.Results[0].Error), "salary")
}

// TestImportCSV_OverMaxRows_400: >500 data rows is a whole-file limit (D8).
func TestImportCSV_OverMaxRows_400(t *testing.T) {
	skipIfNoDB(t)
	// No truncate/DB writes needed — fail before Create. Still needs DB for svc wiring.
	svc, _ := newEmpSvc(testDB)

	rows := make([]string, 0, 501)
	for i := 0; i < 501; i++ {
		rows = append(rows, "u"+uuid.NewString()[:8]+"@example.com,First,Last")
	}
	file := csvBytes("email,first_name,last_name", rows...)
	out, err := svc.ImportCSV(context.Background(), file, services.AllEmployeeFieldPerms)
	require.Error(t, err)
	assert.Nil(t, out)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
	assert.Contains(t, strings.ToLower(ae.Message), "500")
}
