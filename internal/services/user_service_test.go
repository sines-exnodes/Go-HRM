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
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newUserSvc(db *gorm.DB, empSvc *services.EmployeeService) *services.UserService {
	return services.NewUserService(
		repositories.NewUserRepository(db),
		repositories.NewEmployeeRepository(db),
		repositories.NewDeviceTokenRepository(db),
		repositories.NewNotificationSettingsRepository(db),
		empSvc,
	)
}

func TestUserService_ChangePassword(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "pw@example.com", Password: "OldPass1234", FullName: "PW User",
	})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(testDB)
	u, err := userRepo.FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)

	// Wrong current password rejected.
	err = userSvc.ChangePassword(ctx, u, "Wrong", "NewPass1234")
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Correct current password succeeds.
	require.NoError(t, userSvc.ChangePassword(ctx, u, "OldPass1234", "NewPass1234"))

	// New password is what's persisted: old one no longer authenticates.
	fresh, err := userRepo.FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)
	err = userSvc.ChangePassword(ctx, fresh, "OldPass1234", "Whatever1234")
	require.Error(t, err, "old password should no longer work")
	require.NoError(t, userSvc.ChangePassword(ctx, fresh, "NewPass1234", "Whatever1234"))
}

func TestUserService_ChangeEmail(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "old@example.com", Password: "Pass12345", FullName: "Email User",
	})
	require.NoError(t, err)
	// A second user occupying a target email for the conflict case.
	_, err = empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "taken@example.com", Password: "Pass12345", FullName: "Taken",
	})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(testDB)
	u, err := userRepo.FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)

	// Wrong password rejected.
	_, err = userSvc.ChangeEmail(ctx, u, "new@example.com", "wrong")
	require.Error(t, err)

	// Conflict with an existing email.
	_, err = userSvc.ChangeEmail(ctx, u, "taken@example.com", "Pass12345")
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeConflict, ae.Code)

	// Happy path.
	out, err := userSvc.ChangeEmail(ctx, u, "new@example.com", "Pass12345")
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", out.Email)

	var dbEmail string
	require.NoError(t, testDB.Raw("SELECT email FROM users WHERE id = ?", view.UserID).Scan(&dbEmail).Error)
	assert.Equal(t, "new@example.com", dbEmail)
}

func TestUserService_AssignRoles(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	role := makeRole(t, "manager", []permissions.Permission{permissions.PermEmployeesRead}, false)
	target, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "target@example.com", Password: "Pass12345", FullName: "Target",
	})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(testDB)
	admin := makeUser(t, "admin2@example.com", "Pass12345")

	// Admin cannot change own role.
	err = userSvc.AssignRoles(ctx, admin.ID, []uuid.UUID{role.ID}, admin)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Assign to a different user OK.
	require.NoError(t, userSvc.AssignRoles(ctx, target.UserID, []uuid.UUID{role.ID}, admin))
	u, err := userRepo.FindByIDWithRoles(ctx, target.UserID)
	require.NoError(t, err)
	require.Len(t, u.Roles, 1)
	assert.Equal(t, "manager", u.Roles[0].Name)
}

func TestUserService_DeviceTokens(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "dt@example.com", Password: "Pass12345", FullName: "DT User",
	})
	require.NoError(t, err)

	require.NoError(t, userSvc.RegisterDeviceToken(ctx, view.UserID, dto.FcmTokenRequest{
		DeviceID: "dev-1", Token: "tok-abc", Platform: "android",
	}))

	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	tokens, err := tokenRepo.ListByUser(ctx, view.UserID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "tok-abc", tokens[0].Token)

	// Re-register same device with a new token replaces the old one.
	require.NoError(t, userSvc.RegisterDeviceToken(ctx, view.UserID, dto.FcmTokenRequest{
		DeviceID: "dev-1", Token: "tok-xyz", Platform: "android",
	}))
	tokens, err = tokenRepo.ListByUser(ctx, view.UserID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "tok-xyz", tokens[0].Token)

	require.NoError(t, userSvc.RemoveDeviceToken(ctx, view.UserID, "tok-xyz"))
	tokens, err = tokenRepo.ListByUser(ctx, view.UserID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestUserService_GetMe_EmbedsEmployeeSummary(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "me@example.com", Password: "Pass12345", FullName: "Me User",
	})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(testDB)
	u, err := userRepo.FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)

	out, err := userSvc.GetMe(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, out.Employee)
	assert.Equal(t, "Me User", out.Employee.FullName)
	assert.Equal(t, "me@example.com", out.Email)
	assert.True(t, out.NotificationsEnabled, "defaults to enabled when no settings row exists")
}

func TestUserService_NotificationSettings_Toggle(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	view, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "ns@example.com", Password: "Pass12345", FullName: "NS",
	})
	require.NoError(t, err)

	require.NoError(t, userSvc.UpdateNotificationSettings(ctx, view.UserID, false))

	userRepo := repositories.NewUserRepository(testDB)
	u, err := userRepo.FindByIDWithRoles(ctx, view.UserID)
	require.NoError(t, err)
	out, err := userSvc.GetMe(ctx, u)
	require.NoError(t, err)
	assert.False(t, out.NotificationsEnabled)

	// Toggle back on.
	require.NoError(t, userSvc.UpdateNotificationSettings(ctx, view.UserID, true))
	out, err = userSvc.GetMe(ctx, u)
	require.NoError(t, err)
	assert.True(t, out.NotificationsEnabled)
}

func TestUserService_ListAndGet(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()

	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	a, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "list-a@example.com", Password: "Pass12345", FullName: "List A", RoleIDs: []uuid.UUID{role.ID},
	})
	require.NoError(t, err)
	_, err = empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "list-b@example.com", Password: "Pass12345", FullName: "List B",
	})
	require.NoError(t, err)

	// List returns both users with pagination metadata.
	page, err := userSvc.List(ctx, dto.UserListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 2, page.Total)
	assert.Len(t, page.Items, 2)

	// Email search narrows the result.
	filtered, err := userSvc.List(ctx, dto.UserListQuery{Page: 1, PageSize: 10, Search: "list-a"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, filtered.Total)
	require.Len(t, filtered.Items, 1)
	assert.Equal(t, "list-a@example.com", filtered.Items[0].Email)

	// Get returns the user with roles + embedded employee summary.
	got, err := userSvc.Get(ctx, a.UserID)
	require.NoError(t, err)
	assert.Equal(t, "list-a@example.com", got.Email)
	require.NotNil(t, got.Employee)
	assert.Equal(t, "List A", got.Employee.FullName)
	require.Len(t, got.Roles, 1)
	assert.Equal(t, "Employee", got.Roles[0].Name)

	// Unknown ID -> not found.
	_, err = userSvc.Get(ctx, uuid.New())
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.HTTP)
}
