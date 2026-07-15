package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// fakeUploader is an in-memory Uploader: no network/S3 calls. It records the
// last uploaded URL and every Delete target so the avatar test can assert
// without touching real object storage.
type fakeUploader struct {
	uploadedURL string
	subdir      string
	deleted     []string
}

func (f *fakeUploader) Upload(_ context.Context, subdir, ext string, _ []byte, _ string) (string, error) {
	f.subdir = subdir
	f.uploadedURL = "https://fake-bucket.s3.ap-southeast-1.amazonaws.com/" + subdir + "/" + uuid.NewString() + ext
	return f.uploadedURL, nil
}

func (f *fakeUploader) Delete(_ context.Context, publicURL string) error {
	f.deleted = append(f.deleted, publicURL)
	return nil
}

func (f *fakeUploader) PublicURL(key string) string {
	return "https://fake-bucket.s3.ap-southeast-1.amazonaws.com/" + key
}

// empSvcDeps bundles the secondary returns from newEmpSvc. It embeds
// *fakeUploader so existing callers that do `svc, up := newEmpSvc(...)` and
// then access `up.uploadedURL` / `up.deleted` continue to compile unchanged.
type empSvcDeps struct {
	*fakeUploader
	callerUserID uuid.UUID
}

func newEmpSvc(db *gorm.DB) (*services.EmployeeService, empSvcDeps) {
	up := &fakeUploader{}
	empRepo := repositories.NewEmployeeRepository(db)
	skillSvc := services.NewSkillService(
		repositories.NewSkillRepository(db),
		repositories.NewEmployeeSkillRepository(db),
		empRepo,
		up,
	)
	svc := services.NewEmployeeService(
		db,
		empRepo,
		repositories.NewDependentRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewRoleRepository(db),
		repositories.NewLeaveQuotaRepository(db),
		up,
		skillSvc,
	)
	return svc, empSvcDeps{fakeUploader: up, callerUserID: uuid.New()}
}

func TestEmployeeService_CreateAndGet(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email:     "alice@example.com",
		Password:  "StrongPass123",
		FirstName: "Alice",
		LastName:  "Smith",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", view.Email)
	assert.Equal(t, "Alice", view.FirstName)
	assert.Equal(t, "Smith", view.LastName)
	assert.True(t, view.IsActive)

	// The create must persist both a user row and an employee row in one tx.
	var userCount, empCount int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users WHERE id = ?", view.UserID).Scan(&userCount).Error)
	require.NoError(t, testDB.Raw("SELECT count(*) FROM employees WHERE id = ?", view.ID).Scan(&empCount).Error)
	assert.Equal(t, int64(1), userCount)
	assert.Equal(t, int64(1), empCount)

	got, err := svc.Get(ctx, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.FirstName)
	assert.Equal(t, "Smith", got.LastName)
}

func TestEmployeeService_Create_DuplicateEmail_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "dup@example.com", Password: "Pass12345", FirstName: "A", LastName: "Test"})
	require.NoError(t, err)

	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "dup@example.com", Password: "Pass12345", FirstName: "B", LastName: "Test"})
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeConflict, ae.Code)

	// No orphan: still exactly one user with that email, exactly one employee.
	var userCount, empCount int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users WHERE email = ?", "dup@example.com").Scan(&userCount).Error)
	require.NoError(t, testDB.Raw("SELECT count(*) FROM employees").Scan(&empCount).Error)
	assert.Equal(t, int64(1), userCount)
	assert.Equal(t, int64(1), empCount)
}

// TestEmployeeService_Create_RollbackOnFailure proves the user+employee write is
// atomic: a duplicate email (caught at the unique constraint inside the tx)
// must NOT leave an orphan user behind.
func TestEmployeeService_Create_RollbackOnFailure(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	// Seed a user row directly (no employee) sharing the target email so that
	// the inner tx INSERT into users hits the unique constraint and rolls back.
	u := makeUser(t, "race@example.com", "Pass12345")
	require.NotEqual(t, uuid.Nil, u.ID)

	before := int64(0)
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users").Scan(&before).Error)

	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "race@example.com", Password: "Pass12345", FirstName: "Racer", LastName: "Test"})
	require.Error(t, err)

	after := int64(0)
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users").Scan(&after).Error)
	assert.Equal(t, before, after, "failed Create must not create an orphan user")

	var empCount int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM employees").Scan(&empCount).Error)
	assert.Equal(t, int64(0), empCount, "failed Create must not create an orphan employee")
}

// TestEmployeeService_SelfUpdate_WhitelistEnforced is the critical assertion:
// salary/contract/department/position/manager columns MUST be untouched after
// a self-update; only the whitelisted fields change.
func TestEmployeeService_SelfUpdate_WhitelistEnforced(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	salary := 99999.0
	insurance := 88888.0
	contract := "official"
	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "self@example.com", Password: "Pass12345",
		FirstName:       "Self",
		LastName:        "Test",
		BasicSalary:     &salary,
		InsuranceSalary: &insurance,
		ContractType:    &contract,
	})
	require.NoError(t, err)

	// A real employee to satisfy the manager_id FK (employees.manager_id
	// REFERENCES employees(id)). Since Phase 3, employees.department_id and
	// employees.position_id also carry FK constraints, so we must reference
	// real departments/positions rows (synthetic UUIDs would violate the FK).
	boss, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "boss@example.com", Password: "Pass12345", FirstName: "The", LastName: "Boss",
	})
	require.NoError(t, err)

	// Create a real department + position to satisfy the employees FKs.
	dept := uuid.New()
	require.NoError(t, testDB.Exec(
		"INSERT INTO departments (id, name) VALUES (?, ?)",
		dept, "Self-Update Test Dept").Error)
	pos := uuid.New()
	require.NoError(t, testDB.Exec(
		"INSERT INTO positions (id, name) VALUES (?, ?)",
		pos, "Self-Update Test Position").Error)
	mgr := boss.ID
	require.NoError(t, testDB.Exec(
		"UPDATE employees SET department_id = ?, position_id = ?, manager_id = ? WHERE id = ?",
		dept, pos, mgr, view.ID).Error)

	phone := "0123456789"
	addr := "123 New Street"
	updated, err := svc.SelfUpdate(ctx, view.UserID, dto.EmployeeSelfUpdate{
		Phone:            &phone,
		PermanentAddress: &addr,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.Phone)
	assert.Equal(t, "0123456789", *updated.Phone)
	require.NotNil(t, updated.PermanentAddress)
	assert.Equal(t, "123 New Street", *updated.PermanentAddress)

	// Spot-check the raw row: restricted columns UNCHANGED.
	var row struct {
		BasicSalary     float64
		InsuranceSalary float64
		ContractType    string
		DepartmentID    *uuid.UUID
		PositionID      *uuid.UUID
		ManagerID       *uuid.UUID
	}
	require.NoError(t, testDB.Raw(
		"SELECT basic_salary, insurance_salary, contract_type, department_id, position_id, manager_id FROM employees WHERE id = ?",
		view.ID).Scan(&row).Error)
	assert.InDelta(t, 99999.0, row.BasicSalary, 0.01, "salary must not change on self-update")
	assert.InDelta(t, 88888.0, row.InsuranceSalary, 0.01, "insurance salary must not change")
	assert.Equal(t, "official", row.ContractType, "contract type must not change")
	require.NotNil(t, row.DepartmentID)
	assert.Equal(t, dept, *row.DepartmentID, "department must not change on self-update")
	require.NotNil(t, row.PositionID)
	assert.Equal(t, pos, *row.PositionID, "position must not change on self-update")
	require.NotNil(t, row.ManagerID)
	assert.Equal(t, mgr, *row.ManagerID, "manager must not change on self-update")
}

func TestEmployeeService_AdminUpdate_AllowsRestrictedFields(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "admin@example.com", Password: "Pass12345", FirstName: "Admin", LastName: "Test",
	})
	require.NoError(t, err)

	newSalary := 50000.0
	contract := "probation"
	out, err := svc.Update(ctx, view.ID, dto.EmployeeUpdate{
		BasicSalary:  &newSalary,
		ContractType: &contract,
	}, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, out.BasicSalary)
	assert.InDelta(t, 50000.0, *out.BasicSalary, 0.01)
	require.NotNil(t, out.ContractType)
	assert.Equal(t, "probation", *out.ContractType)
}

func TestEmployeeService_List_SearchPagination(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	for _, name := range []string{"Anne", "Brian", "Chloe", "Diana"} {
		_, err := svc.Create(ctx, dto.EmployeeCreate{
			Email: name + "@example.com", Password: "Pass12345", FirstName: name, LastName: "Test",
		})
		require.NoError(t, err)
	}
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 2, Search: "an"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total) // Anne, Brian, Diana — substring "an" in first_name
	assert.Len(t, items, 2)
}

func TestEmployeeService_List_FilterByActive(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	inactive := false
	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "on@example.com", Password: "Pass12345", FirstName: "On", LastName: "Test"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "off@example.com", Password: "Pass12345", FirstName: "Off", LastName: "Test", IsActive: &inactive})
	require.NoError(t, err)

	active := true
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 20, IsActive: &active})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "On", items[0].FirstName)
}

func TestEmployeeService_SoftDelete_CascadesUserDeactivate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "kill@example.com", Password: "Pass12345", FirstName: "Kill", LastName: "Me",
	})
	require.NoError(t, err)
	require.NoError(t, svc.SoftDelete(ctx, view.ID, uuid.New()))

	// Employee no longer fetchable (soft-deleted).
	_, err = svc.Get(ctx, view.ID)
	require.Error(t, err)

	// employees.is_deleted = true
	var isDeleted bool
	require.NoError(t, testDB.Raw("SELECT is_deleted FROM employees WHERE id = ?", view.ID).Scan(&isDeleted).Error)
	assert.True(t, isDeleted, "employee row must be soft-deleted")

	// users.is_active = false (cascade deactivation)
	var isActive bool
	require.NoError(t, testDB.Raw("SELECT is_active FROM users WHERE id = ?", view.UserID).Scan(&isActive).Error)
	assert.False(t, isActive, "linked user must be deactivated")
}

func TestEmployeeService_UpdateAvatar_ChecksImageType(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, up := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "a@example.com", Password: "Pass12345", FirstName: "Avatar", LastName: "Test",
	})
	require.NoError(t, err)

	// Non-image rejected before any upload happens.
	_, err = svc.UpdateAvatarAdmin(ctx, view.ID, []byte("not-an-image"), "application/pdf", ".pdf")
	require.Error(t, err)
	assert.Empty(t, up.uploadedURL, "no upload should occur for a rejected content type")

	pixel := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic
	out, err := svc.UpdateAvatarAdmin(ctx, view.ID, pixel, "image/png", ".png")
	require.NoError(t, err)
	require.NotNil(t, out.AvatarURL)
	assert.Equal(t, up.uploadedURL, *out.AvatarURL)

	// Persisted to the employee row.
	var avatar *string
	require.NoError(t, testDB.Raw("SELECT avatar_url FROM employees WHERE id = ?", view.ID).Scan(&avatar).Error)
	require.NotNil(t, avatar)
	assert.Equal(t, up.uploadedURL, *avatar)
	assert.Equal(t, "hrm-app/avatars", up.subdir)
}

func TestEmployeeService_UpdateAvatar_RejectsSpoofedContentType(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, up := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "spoof@example.com", Password: "Pass12345", FirstName: "Spoof", LastName: "Test",
	})
	require.NoError(t, err)

	// Bytes are plain text (sniffs as text/plain) but the client claims
	// image/jpeg — the authoritative byte sniff must reject this.
	evil := []byte("GIF-looking but actually not; <script>alert(1)</script>")
	_, err = svc.UpdateAvatarAdmin(ctx, view.ID, evil, "image/jpeg", ".jpg")
	require.Error(t, err)
	assert.Empty(t, up.uploadedURL, "spoofed content type must not reach storage")
}

func TestEmployeeService_UpdateAvatar_SelfReplacesAndDeletesOld(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, up := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "av2@example.com", Password: "Pass12345", FirstName: "Avatar", LastName: "Self",
	})
	require.NoError(t, err)

	pixel := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	first, err := svc.UpdateAvatarSelf(ctx, view.UserID, pixel, "image/png", ".png")
	require.NoError(t, err)
	require.NotNil(t, first.AvatarURL)
	oldURL := *first.AvatarURL

	second, err := svc.UpdateAvatarSelf(ctx, view.UserID, pixel, "image/png", ".png")
	require.NoError(t, err)
	require.NotNil(t, second.AvatarURL)
	assert.NotEqual(t, oldURL, *second.AvatarURL)
	assert.Contains(t, up.deleted, oldURL, "previous avatar should be deleted on replace")
}

func TestEmployeeService_List_SingleValueDepartmentFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	dID := uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dID, "Solo").Error)
	v, err := svc.Create(ctx, dto.EmployeeCreate{Email: "solo@x.com", Password: "Pass12345", FirstName: "Solo", LastName: "One"})
	require.NoError(t, err)
	require.NoError(t, testDB.Exec("UPDATE employees SET department_id = ? WHERE id = ?", dID, v.ID).Error)
	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "none@x.com", Password: "Pass12345", FirstName: "No", LastName: "Dept"})
	require.NoError(t, err)

	items, total, err := svc.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 20, DepartmentIDs: []uuid.UUID{dID}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "Solo", items[0].FirstName)
	assert.Equal(t, "One", items[0].LastName)
}

func TestEmployeeService_List_MultiSelectDepartmentFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	// Two real departments to satisfy the employees.department_id FK.
	dA, dB := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dA, "Alpha").Error)
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dB, "Beta").Error)

	mk := func(email, name string, dept *uuid.UUID) {
		v, err := svc.Create(ctx, dto.EmployeeCreate{Email: email, Password: "Pass12345", FirstName: name, LastName: "Test"})
		require.NoError(t, err)
		if dept != nil {
			require.NoError(t, testDB.Exec("UPDATE employees SET department_id = ? WHERE id = ?", *dept, v.ID).Error)
		}
	}
	mk("a@x.com", "Ann", &dA)
	mk("b@x.com", "Bob", &dB)
	mk("c@x.com", "Cara", nil) // unassigned department -> excluded by the filter

	// Filter by BOTH departments -> OR within the filter -> 2 rows.
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{
		Page: 1, PageSize: 20, DepartmentIDs: []uuid.UUID{dA, dB},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
}

func TestEmployeeService_Create_RejectsBadExperienceYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	future := time.Now().UTC().Year() + 1
	_, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "future@x.com", Password: "Pass12345", FirstName: "Future", LastName: "Year",
		ExperienceYear: &future,
	})
	require.Error(t, err, "a future experience_year must be rejected")

	old := 1800
	_, err = svc.Create(ctx, dto.EmployeeCreate{
		Email: "old@x.com", Password: "Pass12345", FirstName: "Too", LastName: "Old",
		ExperienceYear: &old,
	})
	require.Error(t, err, "experience_year <= 1900 must be rejected")

	good := 2018
	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "good@x.com", Password: "Pass12345", FirstName: "Good", LastName: "Year",
		ExperienceYear: &good,
	})
	require.NoError(t, err)
	require.NotNil(t, v.ExperienceYear)
	assert.Equal(t, 2018, *v.ExperienceYear)
}

func TestEmployeeService_Update_RejectsBadExperienceYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "upd-xp@x.com", Password: "Pass12345", FirstName: "Upd", LastName: "Year",
	})
	require.NoError(t, err)

	future := time.Now().UTC().Year() + 1
	_, err = svc.Update(ctx, v.ID, dto.EmployeeUpdate{ExperienceYear: &future}, uuid.New())
	require.Error(t, err, "Update must reject a future experience_year")

	good := 2019
	out, err := svc.Update(ctx, v.ID, dto.EmployeeUpdate{ExperienceYear: &good}, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, out.ExperienceYear)
	assert.Equal(t, 2019, *out.ExperienceYear)
}

func TestEmployeeService_Read_ResolvesDepartmentAndPosition(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	dID, pID := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dID, "Engineering").Error)
	require.NoError(t, testDB.Exec("INSERT INTO positions (id, name) VALUES (?, ?)", pID, "Senior Engineer").Error)

	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "wp@x.com", Password: "Pass12345", FirstName: "Work", LastName: "Profile",
		DepartmentID: &dID, PositionID: &pID,
	})
	require.NoError(t, err)

	got, err := svc.Get(ctx, v.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Department, "department ref must be resolved on read")
	assert.Equal(t, "Engineering", got.Department.Name)
	require.NotNil(t, got.Position, "position ref must be resolved on read")
	assert.Equal(t, "Senior Engineer", got.Position.Name)
}
