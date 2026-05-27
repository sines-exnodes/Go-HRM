package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/exnodes/hrm-api/pkg/utils"
)

const (
	jwtSecret  = "phase1-test-secret"
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

func newAuthSvc() *services.AuthService {
	return services.NewAuthService(testUserRepo, testRoleRepo, services.AuthConfig{
		JWTSecret:            jwtSecret,
		AccessTTL:            accessTTL,
		RefreshTTL:           refreshTTL,
		RememberMeRefreshTTL: 30 * 24 * time.Hour,
		MaxFailedAttempts:    5,
		LockoutDuration:      15 * time.Minute,
	})
}

func TestAuthService_Login_Success(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)
	makeEmployee(t, u, "Alice Tester")

	svc := newAuthSvc()
	result, err := svc.Login(context.Background(), "alice@test.com", "Secret123!", false)
	if err != nil {
		t.Fatalf("login err: %v", err)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
	if result.User == nil || result.User.Employee == nil {
		t.Fatal("expected user with preloaded employee")
	}
	if result.User.Employee.FullName != "Alice Tester" {
		t.Errorf("employee.full_name: got %q", result.User.Employee.FullName)
	}
	claims, err := utils.VerifyToken(result.Tokens.AccessToken, jwtSecret)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Type != utils.TokenTypeAccess {
		t.Errorf("type: got %s", claims.Type)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	_, err := svc.Login(context.Background(), "alice@test.com", "wrong", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Login_UnknownEmail(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newAuthSvc()
	_, err := svc.Login(context.Background(), "ghost@test.com", "anything", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Login_InactiveAccount(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)
	u.IsActive = false
	if err := testUserRepo.Update(context.Background(), u); err != nil {
		t.Fatalf("update: %v", err)
	}

	svc := newAuthSvc()
	_, err := svc.Login(context.Background(), "alice@test.com", "Secret123!", false)
	if err == nil {
		t.Fatal("expected error for inactive account")
	}
}

func TestAuthService_Login_MissingAuthLoginPermission(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	// Role with no permissions at all
	role := makeRole(t, "NoLogin", []permissions.Permission{}, false)
	makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	_, err := svc.Login(context.Background(), "alice@test.com", "Secret123!", false)
	if err == nil {
		t.Fatal("expected error when user lacks auth:login")
	}
}

func TestAuthService_Login_WildcardBypasses(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Super Admin", []permissions.Permission{permissions.PermAll}, true)
	u := makeUser(t, "boss@test.com", "Secret123!", role)
	makeEmployee(t, u, "The Boss")

	svc := newAuthSvc()
	result, err := svc.Login(context.Background(), "boss@test.com", "Secret123!", false)
	if err != nil {
		t.Fatalf("login err: %v", err)
	}
	if result.Tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

// TestAuthService_Login_LocksAfterMaxFailedAttempts encodes the brute-force
// protection rule: the configured number of consecutive bad passwords must
// trigger a temporary lockout. Without this, an attacker can credential-stuff
// indefinitely — the exact regression the Python repo's lockout flow guards
// against, ported here for parity.
func TestAuthService_Login_LocksAfterMaxFailedAttempts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	ctx := context.Background()
	// Attempts 1..4 should be plain "invalid email or password"; the 5th
	// must announce the lockout. The lockout test config uses MaxFailedAttempts=5.
	for i := 1; i <= 4; i++ {
		if _, err := svc.Login(ctx, "alice@test.com", "wrong", false); err == nil {
			t.Fatalf("attempt %d: expected error", i)
		}
	}
	_, err := svc.Login(ctx, "alice@test.com", "wrong", false)
	if err == nil || !strings.Contains(err.Error(), "Account temporarily locked") {
		t.Fatalf("5th attempt: expected lockout error, got %v", err)
	}

	// And the correct password must now be refused while the lockout is in
	// effect — otherwise the lockout is cosmetic.
	if _, err := svc.Login(ctx, "alice@test.com", "Secret123!", false); err == nil ||
		!strings.Contains(err.Error(), "Account temporarily locked") {
		t.Fatalf("correct password during lockout: expected lockout error, got %v", err)
	}
}

// TestAuthService_Login_SuccessResetsFailedAttempts proves a partial-fail-
// then-succeed flow does not leave the counter armed. Without the reset, a
// user who mistypes once is one mistake away from lockout forever.
func TestAuthService_Login_SuccessResetsFailedAttempts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, _ = svc.Login(ctx, "alice@test.com", "wrong", false)
	}
	if _, err := svc.Login(ctx, "alice@test.com", "Secret123!", false); err != nil {
		t.Fatalf("good password after partial fails: %v", err)
	}
	got, err := testUserRepo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts: want 0 after success, got %d", got.FailedLoginAttempts)
	}
	if got.LockedUntil != nil {
		t.Errorf("LockedUntil: want nil after success, got %v", got.LockedUntil)
	}
}

// TestAuthService_Login_RememberMeIssuesLongerRefresh proves remember_me
// actually changes the refresh-token TTL. The test config sets the
// remember-me TTL to 30 days vs the 7-day base, so the issued refresh
// must expire materially later than a non-remember-me one.
func TestAuthService_Login_RememberMeIssuesLongerRefresh(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	ctx := context.Background()
	normal, err := svc.Login(ctx, "alice@test.com", "Secret123!", false)
	if err != nil {
		t.Fatalf("normal login: %v", err)
	}
	remembered, err := svc.Login(ctx, "alice@test.com", "Secret123!", true)
	if err != nil {
		t.Fatalf("remember_me login: %v", err)
	}

	normalClaims, err := utils.VerifyToken(normal.Tokens.RefreshToken, jwtSecret)
	if err != nil {
		t.Fatalf("verify normal refresh: %v", err)
	}
	rememberedClaims, err := utils.VerifyToken(remembered.Tokens.RefreshToken, jwtSecret)
	if err != nil {
		t.Fatalf("verify remembered refresh: %v", err)
	}
	if normalClaims.ExpiresAt == nil || rememberedClaims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt on both refresh tokens")
	}
	// Allow 2-second slack for clock drift between the two Sign calls; the
	// real gap is days, so anything tighter than that is a regression.
	gap := rememberedClaims.ExpiresAt.Time.Sub(normalClaims.ExpiresAt.Time)
	if gap < 24*time.Hour {
		t.Errorf("remember_me refresh expiry should be much later than normal; gap=%v", gap)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	refresh, err := utils.SignToken(u.ID.String(), utils.TokenTypeRefresh, jwtSecret, refreshTTL)
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}
	result, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("refresh err: %v", err)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("expected token pair")
	}
}

func TestAuthService_Refresh_RejectsAccessToken(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	access, _ := utils.SignToken(u.ID.String(), utils.TokenTypeAccess, jwtSecret, accessTTL)
	if _, err := svc.Refresh(context.Background(), access); err == nil {
		t.Fatal("expected error: access token must not work as refresh")
	}
}

func TestAuthService_Refresh_RejectsTokenIssuedBeforePasswordReset(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)

	svc := newAuthSvc()
	refresh, err := utils.SignToken(u.ID.String(), utils.TokenTypeRefresh, jwtSecret, refreshTTL)
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}

	// Reset the password AFTER the refresh token was issued — this stamps
	// password_reset_at = NOW(), which is strictly after the token iat.
	time.Sleep(1100 * time.Millisecond)
	newHash, err := utils.HashPassword("NewSecret123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := testUserRepo.UpdatePassword(context.Background(), u.ID, newHash); err != nil {
		t.Fatalf("update password: %v", err)
	}

	if _, err := svc.Refresh(context.Background(), refresh); err == nil {
		t.Fatal("expected error: refresh token issued before password reset must be rejected")
	}
}

func TestAuthService_ResolveUserPermissions_UnionAcrossRoles(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	r1 := makeRole(t, "R1", []permissions.Permission{permissions.PermUsersRead}, false)
	r2 := makeRole(t, "R2", []permissions.Permission{permissions.PermRolesRead, permissions.PermUsersRead}, false)
	u := makeUser(t, "alice@test.com", "Secret123!", r1, r2)

	svc := newAuthSvc()
	perms, err := svc.ResolveUserPermissions(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !perms[permissions.PermUsersRead] || !perms[permissions.PermRolesRead] {
		t.Fatalf("union missing: %v", perms)
	}
	if perms[permissions.PermAll] {
		t.Fatal("wildcard should not appear unless granted")
	}
}

func TestAuthService_ResolveUserPermissions_Wildcard(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	r := makeRole(t, "Super Admin", []permissions.Permission{permissions.PermAll}, true)
	u := makeUser(t, "boss@test.com", "Secret123!", r)

	svc := newAuthSvc()
	perms, _ := svc.ResolveUserPermissions(context.Background(), u.ID)
	if !perms[permissions.PermAll] {
		t.Fatal("expected wildcard permission")
	}
}
