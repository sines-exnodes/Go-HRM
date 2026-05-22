package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Helpers ----

// inviteTestConfig builds a Config with SMTP disabled — the test path
// exercises the "graceful degradation" branch where EmailService
// returns ErrEmailDisabled and last_email_error is populated.
func inviteTestConfig() *config.Config {
	return &config.Config{
		AppName:                 "Test HRM",
		FrontendURL:             "http://localhost:3000",
		InviteTokenExpireHours:  72,
		SMTPHost:                "", // disabled
		SMTPFromEmail:           "no-reply@example.com",
		SMTPFromName:            "Test HRM",
		FirebaseCredentialsPath: "",
	}
}

func newInviteSvc(t *testing.T) *services.InviteService {
	t.Helper()
	cfg := inviteTestConfig()
	emailSvc, err := services.NewEmailService(cfg)
	require.NoError(t, err)
	// The InviteService needs the full repo + service stack; mirror
	// production wiring with the test DB.
	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	settingsRepo := repositories.NewNotificationSettingsRepository(testDB)
	empRepo := repositories.NewEmployeeRepository(testDB)
	depRepo := repositories.NewDependentRepository(testDB)
	quotaRepo := repositories.NewLeaveQuotaRepository(testDB)
	empSvc := services.NewEmployeeService(testDB, empRepo, depRepo, testUserRepo, testRoleRepo, quotaRepo, nil)
	_ = services.NewUserService(testUserRepo, empRepo, tokenRepo, settingsRepo, empSvc)
	return services.NewInviteService(
		cfg,
		repositories.NewInviteRepository(testDB),
		empRepo,
		testUserRepo,
		testRoleRepo,
		empSvc,
		emailSvc,
		testDB,
	)
}

// ---- Create ----

func TestInvite_Create_Success_RecordsEmailErrorWhenSMTPDisabled(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin-inv@example.com", "Admin")

	out, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "newbie@example.com", out.Email)
	assert.Equal(t, "pending", out.Status)
	require.NotNil(t, out.LastEmailError, "SMTP disabled → last_email_error must be populated")
	assert.Contains(t, *out.LastEmailError, "SMTP not configured")
}

func TestInvite_Create_ExistingUserEmail_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	// Existing user with target email.
	_, _ = makeEmpUser(t, "taken@example.com", "Taken")

	_, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "taken@example.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInvite_Create_DuplicatePending_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	_, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending invite")
}

// ---- Resend ----

func TestInvite_Resend_StampsAndKeepsToken(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	created, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	// Fetch raw row to confirm token doesn't rotate on resend.
	originalToken := readInviteToken(t, created.ID)

	_, err = svc.Resend(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, originalToken, readInviteToken(t, created.ID), "Resend must NOT rotate the token")
}

func TestInvite_Resend_AlreadyAccepted_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	created, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	// Manually mark accepted.
	require.NoError(t, testDB.Exec(`UPDATE invites SET accepted_at = NOW() WHERE id = ?`, created.ID).Error)

	_, err = svc.Resend(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already been accepted")
}

// ---- Revoke ----

func TestInvite_Revoke_SoftDeletes(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	created, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)

	require.NoError(t, svc.Revoke(ctx, created.ID))

	// Subsequent Get must 404 (NotDeleted scope hides it).
	_, err = svc.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---- Accept ----

func TestInvite_Accept_Success_CreatesUserAndStampsRow(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")
	employeeRole := makeRole(t, "Employee", nil, true)

	in := dto.InviteCreate{
		Email:   "newbie@example.com",
		RoleIDs: []uuid.UUID{employeeRole.ID},
	}
	created, err := svc.Create(ctx, admin.ID, in)
	require.NoError(t, err)
	token := readInviteToken(t, created.ID)

	out, err := svc.Accept(ctx, dto.InviteAccept{
		Token:    token,
		Password: "Strong!12345",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "newbie@example.com", out.Email)
	assert.NotEqual(t, uuid.Nil, out.UserID)

	// invite row stamped.
	var inv models.Invite
	require.NoError(t, testDB.First(&inv, "id = ?", created.ID).Error)
	require.NotNil(t, inv.AcceptedAt)
	require.NotNil(t, inv.AcceptedUserID)
	assert.Equal(t, out.UserID, *inv.AcceptedUserID)

	// New user can be looked up.
	u, err := testUserRepo.FindByEmail(ctx, "newbie@example.com")
	require.NoError(t, err)
	require.NotNil(t, u)
}

func TestInvite_Accept_UnknownToken_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	_, err := svc.Accept(ctx, dto.InviteAccept{Token: "definitely-not-real", Password: "Strong!12345"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInvite_Accept_ReusedToken_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	created, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	token := readInviteToken(t, created.ID)

	_, err = svc.Accept(ctx, dto.InviteAccept{Token: token, Password: "Strong!12345"})
	require.NoError(t, err)

	// Second accept with same token → 409.
	_, err = svc.Accept(ctx, dto.InviteAccept{Token: token, Password: "Strong!12345"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already been used")
}

func TestInvite_Accept_RevokedToken_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	created, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "newbie@example.com"})
	require.NoError(t, err)
	token := readInviteToken(t, created.ID)
	require.NoError(t, svc.Revoke(ctx, created.ID))

	_, err = svc.Accept(ctx, dto.InviteAccept{Token: token, Password: "Strong!12345"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---- List ----

func TestInvite_List_StatusFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newInviteSvc(t)
	admin, _ := makeEmpUser(t, "admin@example.com", "Admin")

	_, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "a@example.com"})
	require.NoError(t, err)
	created2, err := svc.Create(ctx, admin.ID, dto.InviteCreate{Email: "b@example.com"})
	require.NoError(t, err)
	// Mark b accepted.
	require.NoError(t, testDB.Exec(`UPDATE invites SET accepted_at = NOW() WHERE id = ?`, created2.ID).Error)

	pending, err := svc.List(ctx, dto.InviteListQuery{Page: 1, PageSize: 50, Status: "pending"})
	require.NoError(t, err)
	require.Len(t, pending.Items, 1)
	assert.Equal(t, "a@example.com", pending.Items[0].Email)

	accepted, err := svc.List(ctx, dto.InviteListQuery{Page: 1, PageSize: 50, Status: "accepted"})
	require.NoError(t, err)
	require.Len(t, accepted.Items, 1)
	assert.Equal(t, "b@example.com", accepted.Items[0].Email)
}

// ---- Helpers ----

func readInviteToken(t *testing.T, id uuid.UUID) string {
	t.Helper()
	var inv models.Invite
	require.NoError(t, testDB.First(&inv, "id = ?", id).Error)
	return inv.Token
}
