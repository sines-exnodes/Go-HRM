package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Helpers ----

// newAttendanceSvc builds an AttendanceService against the shared test DB
// with Phase-6 defaults (Asia/Ho_Chi_Minh, 9:00 late, 18:00 early-leave,
// 4h half-day). GPS disabled — tests don't exercise it.
func newAttendanceSvc(t *testing.T) *services.AttendanceService {
	t.Helper()
	cfg := &config.Config{
		CompanyTimezone:         "Asia/Ho_Chi_Minh",
		LateThresholdHour:       9,
		LateThresholdMinute:     0,
		CheckoutThresholdHour:   18,
		CheckoutThresholdMinute: 0,
		HalfDayHoursThreshold:   4.0,
		OfficeGPSEnabled:        false,
	}
	return services.NewAttendanceService(
		cfg,
		repositories.NewAttendanceRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewPositionRepository(testDB),
	)
}

// hcmTime returns a time in Asia/Ho_Chi_Minh — the configured TZ. The
// service does the local→UTC conversion internally, but every test that
// asserts late/early needs to author its check-in in the company TZ.
func hcmTime(t *testing.T, year int, month time.Month, day, hour, min int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	return time.Date(year, month, day, hour, min, 0, 0, loc)
}

// ---- T12: check-in happy path + late detection ----

func TestAttendance_CheckIn_OnTime(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "alice@example.com", "Alice")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	out, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	assert.False(t, out.IsLate, "8:30 < 9:00 threshold — should not be late")
	require.Len(t, out.Sessions, 1)
	assert.Nil(t, out.Sessions[0].CheckOut, "fresh check-in session must be open")
	assert.Equal(t, emp.ID, out.EmployeeID)
}

func TestAttendance_CheckIn_Late(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "bob@example.com", "Bob")

	ci := hcmTime(t, 2026, 5, 15, 9, 30) // 30 min after threshold
	out, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	assert.True(t, out.IsLate, "9:30 > 9:00 threshold — should be late")
}

// ---- T13: conflicts, multi-session, half-day ----

func TestAttendance_CheckIn_AlreadyOpen_Conflicts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "carol@example.com", "Carol")

	t1 := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &t1})
	require.NoError(t, err)

	t2 := hcmTime(t, 2026, 5, 15, 8, 35)
	_, err = svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &t2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already checked in")
}

func TestAttendance_CheckOut_WithoutCheckIn_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "dave@example.com", "Dave")

	_, err := svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No check-in")
}

func TestAttendance_CheckOut_FullDay_NotHalfDay(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "eve@example.com", "Eve")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := hcmTime(t, 2026, 5, 15, 17, 30) // 9h
	out, err := svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	assert.False(t, out.IsHalfDay)
	require.NotNil(t, out.HoursWorked)
	assert.InDelta(t, 9.0, *out.HoursWorked, 0.05)
}

func TestAttendance_CheckOut_ShortDay_FlagsHalfDay(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "frank@example.com", "Frank")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := hcmTime(t, 2026, 5, 15, 11, 0) // 2.5h < 4h threshold
	out, err := svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	assert.True(t, out.IsHalfDay)
}

func TestAttendance_CheckOut_DoubleClose_Conflicts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "grace@example.com", "Grace")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := hcmTime(t, 2026, 5, 15, 17, 0)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)

	// Second check-out → 409
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not currently checked in")
}

func TestAttendance_SecondSession_AfterCheckout(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "henry@example.com", "Henry")

	ci1 := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci1})
	require.NoError(t, err)
	co1 := hcmTime(t, 2026, 5, 15, 12, 0)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co1})
	require.NoError(t, err)

	ci2 := hcmTime(t, 2026, 5, 15, 13, 30)
	out, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci2})
	require.NoError(t, err)
	require.Len(t, out.Sessions, 2, "second check-in should append a session, not create a new row")
	assert.Nil(t, out.Sessions[1].CheckOut)
}

// is_late is computed once from the FIRST check-in (REVISION NOTES #5).
// A second on-time session must NOT clear an earlier late flag.
func TestAttendance_IsLate_NotRecomputed_OnSecondSession(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "isaac@example.com", "Isaac")

	// First check-in late (9:30).
	ci1 := hcmTime(t, 2026, 5, 15, 9, 30)
	out, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci1})
	require.NoError(t, err)
	require.True(t, out.IsLate)

	// Check out + check back in early in the afternoon.
	co := hcmTime(t, 2026, 5, 15, 12, 0)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	ci2 := hcmTime(t, 2026, 5, 15, 13, 0)
	out, err = svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci2})
	require.NoError(t, err)
	assert.True(t, out.IsLate, "is_late must remain true — never recomputed on subsequent sessions")
}

// ---- T14: list filters, ownership, admin CRUD ----

func TestAttendance_List_OwnerSeesOnlySelf(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	alice, aliceEmp := makeEmpUser(t, "alice2@example.com", "Alice")
	bob, _ := makeEmpUser(t, "bob2@example.com", "Bob")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, alice.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	_, err = svc.CheckIn(ctx, bob.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	// Non-admin: should see only own row.
	out, err := svc.List(ctx, alice.ID, false, dto.AttendanceListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, aliceEmp.ID, out.Items[0].EmployeeID)
}

func TestAttendance_List_ManagerSeesAll(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "manager@example.com", "Manager")
	alice, _ := makeEmpUser(t, "alice3@example.com", "Alice")
	bob, _ := makeEmpUser(t, "bob3@example.com", "Bob")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, alice.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	_, err = svc.CheckIn(ctx, bob.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	// asAdmin=true: should see both rows.
	out, err := svc.List(ctx, mgr.ID, true, dto.AttendanceListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(out.Items), 2)
}

func TestAttendance_List_DateRangeFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "mgr2@example.com", "Mgr")
	_, alice := makeEmpUser(t, "alice4@example.com", "Alice")

	// Two attendance rows via admin create.
	_, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	_, err = svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-20"})
	require.NoError(t, err)

	out, err := svc.List(ctx, mgr.ID, true, dto.AttendanceListQuery{
		Page: 1, PageSize: 50,
		StartDate: "2026-05-15", EndDate: "2026-05-31",
	})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "2026-05-20", out.Items[0].Date)
}

func TestAttendance_AdminCreate_DuplicateConflicts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	_, alice := makeEmpUser(t, "alice5@example.com", "Alice")

	_, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	_, err = svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAttendance_AdminCreate_UnknownEmployee_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, false)

	// Random employee ID that doesn't exist.
	out, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{
		EmployeeID: models.Attendance{}.ID, // zero UUID, definitely not in employees
		Date:       "2026-05-10",
	})
	require.Error(t, err)
	assert.Empty(t, out.ID)
}

func TestAttendance_AdminUpdate_ChangesFieldsAndAppendsSession(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	_, alice := makeEmpUser(t, "alice6@example.com", "Alice")

	created, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	assert.Empty(t, created.Sessions)

	newNotes := "manual late entry"
	trueP := true
	// No existing session → AdminUpdate should append a new one when CheckIn provided.
	ci := hcmTime(t, 2026, 5, 10, 8, 0)
	updated, err := svc.AdminUpdate(ctx, created.ID, dto.AttendanceAdminUpdateReq{
		Notes:   &newNotes,
		IsLate:  &trueP,
		CheckIn: &ci,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.Notes)
	assert.Equal(t, "manual late entry", *updated.Notes)
	assert.True(t, updated.IsLate)
	require.Len(t, updated.Sessions, 1)
}

func TestAttendance_AdminDelete_SoftDeletesAndHidesFromGet(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "mgr3@example.com", "Mgr")
	_, alice := makeEmpUser(t, "alice7@example.com", "Alice")

	created, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	require.NoError(t, svc.AdminDelete(ctx, created.ID))

	_, err = svc.Get(ctx, created.ID, mgr.ID, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAttendance_Get_OwnershipEnforced(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	_, alice := makeEmpUser(t, "alice8@example.com", "Alice")
	bob, _ := makeEmpUser(t, "bob4@example.com", "Bob")

	created, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)

	_, err = svc.Get(ctx, created.ID, bob.ID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "do not own")
}

// ---- Today + DepartmentID filter (regression coverage from end-to-end walk) ----

// TestAttendance_Today_AfterCheckOut exercises the Today endpoint after a
// full check-in/check-out cycle. Caught a Postgres "ambiguous column
// reference" 500 during the Phase-6 e2e walk: the joins in
// MonthlyCheckInCount / DatesWithCheckIn collided with the
// unqualified NotDeleted scope. Test fixes the regression.
func TestAttendance_Today_AfterCheckOut(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "today@example.com", "Today")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	co := hcmTime(t, 2026, 5, 15, 17, 30)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)

	// Today reads from a different perspective ("today in company TZ"),
	// so the in-month attendance row may not be visible — the call must
	// still succeed without a 500.
	out, err := svc.Today(ctx, u.ID)
	require.NoError(t, err)
	// Status is either checked_out (if the test day matches "today" in
	// the TZ) or not_checked_in (otherwise). Either way, monthly_count
	// and streak must surface as integers without erroring.
	assert.Contains(t, []string{"checked_out", "not_checked_in"}, out.Status)
	assert.GreaterOrEqual(t, out.MonthlyCount, 0)
	assert.GreaterOrEqual(t, out.Streak, 0)
}

// TestAttendance_List_DepartmentFilter regression-tests the optional
// employees-join branch in List() that prompted the same column-
// ambiguity 500 surfaced by the e2e walk.
func TestAttendance_List_DepartmentFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "mgr5@example.com", "Mgr")

	// One employee in a real department; one with no department.
	dept := &models.Department{Name: "Phase6Dept"}
	require.NoError(t, testDB.Create(dept).Error)
	u1 := makeUser(t, "withdept@example.com", "pw-Aa123456")
	e1 := &models.Employee{
		UserID: u1.ID, FirstName: "WithDept", LastName: "Test",
		ContractType: "official", ContractRenewal: 1, PaymentMethod: "bank_transfer",
		DepartmentID: &dept.ID,
	}
	require.NoError(t, testDB.Create(e1).Error)
	u2, _ := makeEmpUser(t, "nodept@example.com", "NoDept")

	_, err := svc.AdminCreate(ctx, dto.AttendanceAdminCreateReq{EmployeeID: e1.ID, Date: "2026-05-15"})
	require.NoError(t, err)
	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err = svc.CheckIn(ctx, u2.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	out, err := svc.List(ctx, mgr.ID, true, dto.AttendanceListQuery{
		Page: 1, PageSize: 50,
		DepartmentID: dept.ID.String(),
	})
	require.NoError(t, err)
	require.Len(t, out.Items, 1, "only the with-dept row should match")
	assert.Equal(t, e1.ID, out.Items[0].EmployeeID)
}

// ---- T15: matrix ----

func TestAttendance_Matrix_ManagerSeesAllEmployees(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "mgr4@example.com", "Mgr")
	makeEmpUser(t, "e1@example.com", "E1")
	makeEmpUser(t, "e2@example.com", "E2")

	out, err := svc.Matrix(ctx, mgr.ID, true, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, 2026, out.Year)
	assert.Equal(t, 5, out.Month)
	assert.Equal(t, 31, out.DaysInMonth)
	// mgr + e1 + e2 — all three rows surface for an admin caller.
	assert.GreaterOrEqual(t, len(out.Items), 3)
}

func TestAttendance_Matrix_EmployeeSeesOwnRow(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "soleemp@example.com", "Sole")
	makeEmpUser(t, "other@example.com", "Other") // exists but should be excluded

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, emp.ID, out.Items[0].EmployeeID)
}

func TestAttendance_Matrix_WeekendsMarked(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "weekendcheck@example.com", "Wend")

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	row := out.Items[0]
	// May 2026: 2 = Sat, 3 = Sun.
	assert.Equal(t, "weekend", row.Cells[2].Status)
	assert.Equal(t, "weekend", row.Cells[3].Status)
}
