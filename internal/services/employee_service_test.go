package services_test

import (
	"context"
	"testing"

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
	deleted     []string
}

func (f *fakeUploader) Upload(ctx context.Context, subdir, ext string, content []byte, contentType string) (string, error) {
	f.uploadedURL = "https://fake.supabase.co/storage/v1/object/public/hrm-uploads/" + subdir + "/" + uuid.NewString() + ext
	return f.uploadedURL, nil
}

func (f *fakeUploader) Delete(ctx context.Context, publicURL string) error {
	f.deleted = append(f.deleted, publicURL)
	return nil
}

func (f *fakeUploader) PublicURL(key string) string {
	return "https://fake.supabase.co/storage/v1/object/public/hrm-uploads/" + key
}

func newEmpSvc(db *gorm.DB) (*services.EmployeeService, *fakeUploader) {
	up := &fakeUploader{}
	return services.NewEmployeeService(
		db,
		repositories.NewEmployeeRepository(db),
		repositories.NewDependentRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewRoleRepository(db),
		repositories.NewLeaveQuotaRepository(db),
		up,
	), up
}

func TestEmployeeService_CreateAndGet(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email:    "alice@example.com",
		Password: "StrongPass123",
		FullName: "Alice Smith",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", view.Email)
	assert.Equal(t, "Alice Smith", view.FullName)
	assert.True(t, view.IsActive)

	// The create must persist both a user row and an employee row in one tx.
	var userCount, empCount int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users WHERE id = ?", view.UserID).Scan(&userCount).Error)
	require.NoError(t, testDB.Raw("SELECT count(*) FROM employees WHERE id = ?", view.ID).Scan(&empCount).Error)
	assert.Equal(t, int64(1), userCount)
	assert.Equal(t, int64(1), empCount)

	got, err := svc.Get(ctx, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", got.FullName)
}

func TestEmployeeService_Create_DuplicateEmail_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "dup@example.com", Password: "Pass12345", FullName: "A"})
	require.NoError(t, err)

	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "dup@example.com", Password: "Pass12345", FullName: "B"})
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

	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "race@example.com", Password: "Pass12345", FullName: "Racer"})
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
		FullName:        "Self Test",
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
		Email: "boss@example.com", Password: "Pass12345", FullName: "The Boss",
	})
	require.NoError(t, err)

	// Create a real department + position to satisfy the employees FKs.
	dept := uuid.New()
	require.NoError(t, testDB.Exec(
		"INSERT INTO departments (id, name) VALUES (?, ?)",
		dept, "Self-Update Test Dept").Error)
	pos := uuid.New()
	require.NoError(t, testDB.Exec(
		"INSERT INTO positions (id, name, department_id) VALUES (?, ?, ?)",
		pos, "Self-Update Test Position", dept).Error)
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
		Email: "admin@example.com", Password: "Pass12345", FullName: "Admin Test",
	})
	require.NoError(t, err)

	newSalary := 50000.0
	contract := "probation"
	out, err := svc.Update(ctx, view.ID, dto.EmployeeUpdate{
		BasicSalary:  &newSalary,
		ContractType: &contract,
	})
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
			Email: name + "@example.com", Password: "Pass12345", FullName: name,
		})
		require.NoError(t, err)
	}
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 2, Search: "an"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total) // Anne, Brian, Diana — substring "an"
	assert.Len(t, items, 2)
}

func TestEmployeeService_List_FilterByActive(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	inactive := false
	_, err := svc.Create(ctx, dto.EmployeeCreate{Email: "on@example.com", Password: "Pass12345", FullName: "On"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "off@example.com", Password: "Pass12345", FullName: "Off", IsActive: &inactive})
	require.NoError(t, err)

	active := true
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 20, IsActive: &active})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "On", items[0].FullName)
}

func TestEmployeeService_SoftDelete_CascadesUserDeactivate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "kill@example.com", Password: "Pass12345", FullName: "Kill Me",
	})
	require.NoError(t, err)
	require.NoError(t, svc.SoftDelete(ctx, view.ID))

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
		Email: "a@example.com", Password: "Pass12345", FullName: "Avatar Test",
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
}

func TestEmployeeService_UpdateAvatar_SelfReplacesAndDeletesOld(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, up := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "av2@example.com", Password: "Pass12345", FullName: "Avatar Self",
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
