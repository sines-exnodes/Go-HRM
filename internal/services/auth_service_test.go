package services_test

import (
	"context"
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
		JWTSecret:  jwtSecret,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	})
}

func TestAuthService_Login_Success(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
	u := makeUser(t, "alice@test.com", "Secret123!", role)
	makeEmployee(t, u, "Alice Tester")

	svc := newAuthSvc()
	result, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
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
	_, err := svc.Login(context.Background(), "alice@test.com", "wrong")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Login_UnknownEmail(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newAuthSvc()
	_, err := svc.Login(context.Background(), "ghost@test.com", "anything")
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
	_, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
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
	_, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
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
	result, err := svc.Login(context.Background(), "boss@test.com", "Secret123!")
	if err != nil {
		t.Fatalf("login err: %v", err)
	}
	if result.Tokens.AccessToken == "" {
		t.Fatal("expected access token")
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
