package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// ---------------------------------------------------------------------------
// Employees Python-shape parity (PR A — audit decisions #4/#5/#7/#8/#12/#13).
// Each test encodes WHY the behavior matters, not just what it does.
// ---------------------------------------------------------------------------

// #4 — emergency contacts are a list, persisted on create and surfaced on read.
func TestEmployeeParity_Create_WithEmergencyContacts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "ec@example.com", Password: "Pass12345", FullName: "EC Owner",
		EmergencyContacts: []dto.EmergencyContactInput{
			{FullName: "Mom", Relationship: "parent", PhoneNumber: "0900000001"},
			{FullName: "Dad", Relationship: "parent", PhoneNumber: "0900000002"},
		},
	})
	require.NoError(t, err)
	require.Len(t, view.EmergencyContacts, 2, "both contacts must round-trip on the read shape")
	names := []string{view.EmergencyContacts[0].FullName, view.EmergencyContacts[1].FullName}
	assert.Contains(t, names, "Mom")
	assert.Contains(t, names, "Dad")

	var n int64
	require.NoError(t, testDB.Raw(
		"SELECT count(*) FROM employee_emergency_contacts WHERE employee_id = ? AND is_deleted = false",
		view.ID).Scan(&n).Error)
	assert.EqualValues(t, 2, n, "two live contact rows must exist")
}

// #4 — replace-set semantics: non-empty replaces, nil leaves unchanged, [] clears.
func TestEmployeeParity_Update_ReplaceEmergencyContacts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "ec2@example.com", Password: "Pass12345", FullName: "EC2",
		EmergencyContacts: []dto.EmergencyContactInput{{FullName: "Initial", Relationship: "spouse"}},
	})
	require.NoError(t, err)
	require.Len(t, view.EmergencyContacts, 1)

	// non-empty -> replace
	two := []dto.EmergencyContactInput{{FullName: "A"}, {FullName: "B"}}
	out, err := svc.Update(ctx, view.ID, dto.EmployeeUpdate{EmergencyContacts: &two}, uuid.New())
	require.NoError(t, err)
	require.Len(t, out.EmergencyContacts, 2, "non-empty list replaces the set")

	// nil -> leave unchanged
	out, err = svc.Update(ctx, view.ID, dto.EmployeeUpdate{}, uuid.New())
	require.NoError(t, err)
	require.Len(t, out.EmergencyContacts, 2, "absent field must not touch the existing set")

	// [] -> clear
	empty := []dto.EmergencyContactInput{}
	out, err = svc.Update(ctx, view.ID, dto.EmployeeUpdate{EmergencyContacts: &empty}, uuid.New())
	require.NoError(t, err)
	require.Len(t, out.EmergencyContacts, 0, "empty list clears the set")
}

// #5 — leave quota is hydrated into the read shape (was previously write-only).
func TestEmployeeParity_Read_HydratesLeaveQuota(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "q@example.com", Password: "Pass12345", FullName: "Quota",
	})
	require.NoError(t, err)

	// No quota row yet -> Python-shape defaults (12 / 6).
	got, err := svc.Get(ctx, view.ID)
	require.NoError(t, err)
	assert.InDelta(t, 12.0, got.AnnualLeaveQuota, 0.01, "defaults to 12 when no quota row exists")
	assert.InDelta(t, 6.0, got.SickLeaveQuota, 0.01, "defaults to 6 when no quota row exists")

	// After an explicit quota update the read reflects it.
	_, err = svc.UpdateLeaveQuota(ctx, view.ID, dto.LeaveQuotaUpdateRequest{AnnualLeaveQuota: 20, SickLeaveQuota: 9})
	require.NoError(t, err)
	got, err = svc.Get(ctx, view.ID)
	require.NoError(t, err)
	assert.InDelta(t, 20.0, got.AnnualLeaveQuota, 0.01)
	assert.InDelta(t, 9.0, got.SickLeaveQuota, 0.01)
}

// #8 — skills are embedded on the detail read, hydrated from employee_skills.
func TestEmployeeParity_Read_EmbedsSkills(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "sk@example.com", Password: "Pass12345", FullName: "Skilled",
	})
	require.NoError(t, err)

	sid := uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO skills (id, name) VALUES (?, ?)", sid, "Golang").Error)
	require.NoError(t, repositories.NewEmployeeSkillRepository(testDB).
		ReplaceForEmployee(ctx, view.ID, []uuid.UUID{sid}))

	got, err := svc.Get(ctx, view.ID)
	require.NoError(t, err)
	require.Len(t, got.Skills, 1, "assigned skill must appear on the read shape")
	assert.Equal(t, "Golang", got.Skills[0].Name)
	assert.Equal(t, sid, got.Skills[0].ID)
}

// #8 — experience_year + cv_url round-trip through create and read.
func TestEmployeeParity_Read_ExposesExperienceAndCV(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	yrs := 2018 // experience_year is a career-start YEAR (parity round 2), not a count
	cv := "https://files.example.com/cv.pdf"
	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "xp@example.com", Password: "Pass12345", FullName: "Experienced",
		ExperienceYear: &yrs, CVURL: &cv,
	})
	require.NoError(t, err)
	require.NotNil(t, view.ExperienceYear)
	assert.Equal(t, 2018, *view.ExperienceYear)
	require.NotNil(t, view.CVURL)
	assert.Equal(t, cv, *view.CVURL)
}

// #7 — self-service may now edit identity fields (full_name/gender/dob), but
// the salary whitelist must still hold (the load-bearing security invariant).
func TestEmployeeParity_SelfUpdate_AllowsIdentityFields(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	salary := 1000.0
	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "id@example.com", Password: "Pass12345", FullName: "Before Name",
		BasicSalary: &salary,
	})
	require.NoError(t, err)

	name := "After Name"
	gender := "female"
	dob := time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)
	out, err := svc.SelfUpdate(ctx, view.UserID, dto.EmployeeSelfUpdate{
		FullName: &name, Gender: &gender, DOB: &dob,
	})
	require.NoError(t, err)
	assert.Equal(t, "After Name", out.FullName, "self may rename per decision #7")
	require.NotNil(t, out.Gender)
	assert.Equal(t, "female", *out.Gender)
	require.NotNil(t, out.DOB)

	// Salary must NOT have changed — the self whitelist is the security boundary.
	var dbSalary float64
	require.NoError(t, testDB.Raw("SELECT basic_salary FROM employees WHERE id = ?", view.ID).Scan(&dbSalary).Error)
	assert.InDelta(t, 1000.0, dbSalary, 0.01, "self-update must never touch salary")
}

// #7 — self may manage their own emergency-contact list.
func TestEmployeeParity_SelfUpdate_ManagesOwnEmergencyContacts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "ecself@example.com", Password: "Pass12345", FullName: "EC Self",
	})
	require.NoError(t, err)

	cs := []dto.EmergencyContactInput{{FullName: "Spouse", Relationship: "spouse", PhoneNumber: "0911"}}
	out, err := svc.SelfUpdate(ctx, view.UserID, dto.EmployeeSelfUpdate{EmergencyContacts: &cs})
	require.NoError(t, err)
	require.Len(t, out.EmergencyContacts, 1)
	assert.Equal(t, "Spouse", out.EmergencyContacts[0].FullName)
}

// #12 — you cannot soft-delete your own employee record.
func TestEmployeeParity_SoftDelete_RejectsSelf(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "selfdel@example.com", Password: "Pass12345", FullName: "Self Del",
	})
	require.NoError(t, err)

	err = svc.SoftDelete(ctx, view.ID, view.UserID) // caller IS the target
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	_, err = svc.Get(ctx, view.ID)
	require.NoError(t, err, "employee must still be alive after a rejected self-delete")
}

// #12 — deleting a manager clears manager_id on their direct reports so no live
// employee points at a soft-deleted manager (avoids a dangling reference).
func TestEmployeeParity_SoftDelete_ClearsSubordinateManager(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	mgr, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "boss@example.com", Password: "Pass12345", FullName: "Boss",
	})
	require.NoError(t, err)
	sub, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "report@example.com", Password: "Pass12345", FullName: "Report",
		ManagerID: &mgr.ID,
	})
	require.NoError(t, err)

	require.NoError(t, svc.SoftDelete(ctx, mgr.ID, uuid.New())) // non-self caller

	var nullCount int64
	require.NoError(t, testDB.Raw(
		"SELECT count(*) FROM employees WHERE id = ? AND manager_id IS NULL",
		sub.ID).Scan(&nullCount).Error)
	assert.EqualValues(t, 1, nullCount, "subordinate's manager_id must be cleared when the manager is deleted")
}

// #12 — you cannot deactivate your own account via the admin employee update.
func TestEmployeeParity_Update_RejectsSelfDeactivate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	view, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "selfdeact@example.com", Password: "Pass12345", FullName: "Self Deact",
	})
	require.NoError(t, err)

	no := false
	_, err = svc.Update(ctx, view.ID, dto.EmployeeUpdate{IsActive: &no}, view.UserID) // caller IS target
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	var active bool
	require.NoError(t, testDB.Raw("SELECT is_active FROM users WHERE id = ?", view.UserID).Scan(&active).Error)
	assert.True(t, active, "account must remain active after a rejected self-deactivate")
}

// #12 — the same guard on the auth-side admin user patch.
func TestUserParity_AdminPatch_RejectsSelfDeactivate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "adminself@example.com", Password: "Pass12345", FullName: "Admin Self",
	})
	require.NoError(t, err)
	admin, err := repositories.NewUserRepository(testDB).FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)

	no := false
	err = userSvc.AdminPatch(ctx, admin.ID, dto.AdminUserPatch{IsActive: &no}, admin)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Deactivating a different user is allowed.
	other, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "otheruser@example.com", Password: "Pass12345", FullName: "Other",
	})
	require.NoError(t, err)
	require.NoError(t, userSvc.AdminPatch(ctx, other.UserID, dto.AdminUserPatch{IsActive: &no}, admin))
}

// Inline skill_ids on create + update (Python parity).
func TestEmployeeParity_InlineSkillAssignment(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, deps := newEmpSvc(testDB)
	ctx := context.Background()

	// Two skills in the catalog.
	s1, s2 := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO skills (id, name, description) VALUES (?, ?, '')", s1, "Go").Error)
	require.NoError(t, testDB.Exec("INSERT INTO skills (id, name, description) VALUES (?, ?, '')", s2, "SQL").Error)

	// Create with skills inline.
	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "skilled@x.com", Password: "Pass12345", FullName: "Skilled One",
		SkillIDs: []uuid.UUID{s1, s2},
	})
	require.NoError(t, err)
	require.Len(t, v.Skills, 2, "both skills must be assigned on create")

	// Update -> replace down to one.
	one := []uuid.UUID{s1}
	v2, err := svc.Update(ctx, v.ID, dto.EmployeeUpdate{SkillIDs: &one}, deps.callerUserID)
	require.NoError(t, err)
	require.Len(t, v2.Skills, 1)
	assert.Equal(t, "Go", v2.Skills[0].Name)

	// Update with empty slice -> clear all.
	empty := []uuid.UUID{}
	v3, err := svc.Update(ctx, v.ID, dto.EmployeeUpdate{SkillIDs: &empty}, deps.callerUserID)
	require.NoError(t, err)
	require.Len(t, v3.Skills, 0)

	// Invalid skill on create -> 400, no user created (skills are pre-validated
	// before the user/employee tx, so a bad skill_id must not leave an orphan).
	_, err = svc.Create(ctx, dto.EmployeeCreate{
		Email: "bad@x.com", Password: "Pass12345", FullName: "Bad Skill",
		SkillIDs: []uuid.UUID{uuid.New()},
	})
	require.Error(t, err)
	var orphan int64
	require.NoError(t, testDB.Raw("SELECT count(*) FROM users WHERE email = ?", "bad@x.com").Scan(&orphan).Error)
	assert.Zero(t, orphan, "a bad skill_id must not create an orphan user")
}

// #13 — admin can change another user's email; the change stamps
// email_changed_at, and uniqueness / no-op rules are enforced.
func TestUserParity_AdminChangeEmail(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "old2@example.com", Password: "Pass12345", FullName: "Old Email",
	})
	require.NoError(t, err)
	_, err = empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "taken2@example.com", Password: "Pass12345", FullName: "Taken",
	})
	require.NoError(t, err)

	var ae *apperrors.AppError

	// Same email -> bad request (no-op rejected).
	_, err = userSvc.AdminChangeEmail(ctx, view.UserID, "old2@example.com")
	require.Error(t, err)
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Conflict with another user's email.
	_, err = userSvc.AdminChangeEmail(ctx, view.UserID, "taken2@example.com")
	require.Error(t, err)
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeConflict, ae.Code)

	// Happy path.
	out, err := userSvc.AdminChangeEmail(ctx, view.UserID, "fresh2@example.com")
	require.NoError(t, err)
	assert.Equal(t, "fresh2@example.com", out.Email)

	var stamped int64
	require.NoError(t, testDB.Raw(
		"SELECT count(*) FROM users WHERE id = ? AND email_changed_at IS NOT NULL",
		view.UserID).Scan(&stamped).Error)
	assert.EqualValues(t, 1, stamped, "admin email change must stamp email_changed_at to invalidate sessions")
}
