package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// AuthConfig configures token TTLs, secret, and brute-force-protection
// parameters for the login flow.
type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	// RememberMeRefreshTTL is the refresh-token lifetime when the caller
	// sets remember_me=true on /auth/login. Zero falls back to RefreshTTL.
	RememberMeRefreshTTL time.Duration

	// MaxFailedAttempts is the number of consecutive bad passwords that
	// trigger a temporary account lockout. Zero disables the feature.
	MaxFailedAttempts int

	// LockoutDuration is how long the account stays locked once the
	// threshold is hit. Ignored when MaxFailedAttempts == 0.
	LockoutDuration time.Duration
}

// TokenPair is the access+refresh result of Login/Refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string // always "Bearer"
}

// LoginResult bundles the token pair with the authenticated user (with
// roles and employee profile preloaded) so handlers can render the full
// login response shape without a second DB round-trip.
type LoginResult struct {
	Tokens TokenPair
	User   *models.User // includes Roles and Employee
}

// AuthService handles login, refresh, and permission resolution.
type AuthService struct {
	users repositories.UserRepository
	roles repositories.RoleRepository
	cfg   AuthConfig
}

// NewAuthService constructs an AuthService.
func NewAuthService(users repositories.UserRepository, roles repositories.RoleRepository, cfg AuthConfig) *AuthService {
	return &AuthService{users: users, roles: roles, cfg: cfg}
}

// Login authenticates an email/password pair and returns a token pair plus
// the authenticated user with Roles and Employee preloaded.
//
// Flow (mirrors the Python repo for parity):
//  1. Look up user by email.
//  2. Enforce account lockout if locked_until is in the future.
//  3. Reject if no password has been set (invite flow).
//  4. Verify password — on failure increment the counter and lock after
//     the configured threshold.
//  5. Reject if the account is deactivated (checked AFTER the password
//     so a wrong-password attempt cannot be told apart from an attempt
//     on a disabled account).
//  6. Require auth:login permission (or wildcard).
//  7. Reset counter + locked_until and issue a token pair. When
//     rememberMe=true the refresh token uses RememberMeRefreshTTL.
func (s *AuthService) Login(ctx context.Context, email, password string, rememberMe bool) (*LoginResult, error) {
	user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrUnauthorized("Invalid email or password")
		}
		return nil, err
	}

	now := time.Now().UTC()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		remaining := user.LockedUntil.Sub(now)
		// Round UP so the user is never told "0 minutes" — matches Python.
		mins := int(remaining/time.Minute) + 1
		if mins < 1 {
			mins = 1
		}
		return nil, apperr.ErrUnauthorized(fmt.Sprintf("Account temporarily locked. Try again in %d minutes.", mins))
	}

	if user.PasswordHash == "" {
		return nil, apperr.ErrUnauthorized("Please set your password using the invite link sent to your email.")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, s.recordFailedLogin(ctx, user)
	}

	if !user.IsActive {
		return nil, apperr.ErrUnauthorized("Your account has been deactivated. Contact your administrator.")
	}

	perms, err := s.resolvePermsFromUser(user.Roles)
	if err != nil {
		return nil, err
	}
	if !perms[permissions.PermAll] && !perms[permissions.PermAuthLogin] {
		return nil, apperr.ErrForbidden("You do not have permission to access this system.")
	}

	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		if err := s.users.SetLoginAttempts(ctx, user.ID, 0, nil); err != nil {
			return nil, err
		}
	}

	tokens, err := s.issueTokenPair(user.ID, rememberMe)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Tokens: *tokens, User: user}, nil
}

// recordFailedLogin increments the user's failed-attempt counter, stamps
// locked_until once the threshold is hit, and returns the user-facing
// 401 error. Persistence errors are surfaced (the caller treats them as
// transient and does NOT leak them as auth failures).
func (s *AuthService) recordFailedLogin(ctx context.Context, user *models.User) error {
	// Lockout disabled — preserve the original generic 401.
	if s.cfg.MaxFailedAttempts <= 0 {
		return apperr.ErrUnauthorized("Invalid email or password")
	}

	attempts := user.FailedLoginAttempts + 1
	if attempts >= s.cfg.MaxFailedAttempts {
		lockUntil := time.Now().UTC().Add(s.cfg.LockoutDuration)
		// Reset the counter to zero on lock so the lockout window resets
		// from scratch the next time the account unlocks — matches Python.
		if err := s.users.SetLoginAttempts(ctx, user.ID, 0, &lockUntil); err != nil {
			return err
		}
		mins := int(s.cfg.LockoutDuration / time.Minute)
		if mins < 1 {
			mins = 1
		}
		return apperr.ErrUnauthorized(fmt.Sprintf("Account temporarily locked. Try again in %d minutes.", mins))
	}

	if err := s.users.SetLoginAttempts(ctx, user.ID, attempts, nil); err != nil {
		return err
	}
	return apperr.ErrUnauthorized("Invalid email or password")
}

// Refresh exchanges a refresh token for a new token pair (and returns the
// refreshed User with Roles + Employee preloaded for the response shape).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	claims, err := utils.VerifyToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, apperr.ErrUnauthorized("Invalid or expired refresh token")
	}
	if claims.Type != utils.TokenTypeRefresh {
		return nil, apperr.ErrBadRequest("Invalid token type")
	}
	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperr.ErrUnauthorized("Invalid token subject")
	}
	user, err := s.users.FindByIDWithRolesAndEmployee(ctx, uid)
	if err != nil {
		return nil, apperr.ErrUnauthorized("User not found or inactive")
	}
	if !user.IsActive {
		return nil, apperr.ErrUnauthorized("User not found or inactive")
	}

	// Session invalidation: reject refresh tokens issued before a
	// credential change, mirroring the access-token check in
	// middleware.JWT so a stolen pre-change refresh token cannot mint
	// fresh access tokens.
	if claims.IssuedAt != nil {
		iat := claims.IssuedAt.Time
		if tokenInvalidatedBy(user.EmailChangedAt, iat) {
			return nil, apperr.ErrUnauthorized("Session expired due to email change — please log in again")
		}
		if tokenInvalidatedBy(user.PasswordResetAt, iat) {
			return nil, apperr.ErrUnauthorized("Session expired due to password reset — please log in again")
		}
	}

	tokens, err := s.issueTokenPair(user.ID, false)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Tokens: *tokens, User: user}, nil
}

// Logout is stateless in Phase 1 — the client discards its tokens. Future
// phases may add a server-side blacklist. The method exists so the handler
// has a stable seam.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	_ = ctx
	_ = userID
	return nil
}

// ResolveUserPermissions returns the union of permissions across the user's
// roles. Keys are the permission strings; presence == granted.
func (s *AuthService) ResolveUserPermissions(ctx context.Context, userID uuid.UUID) (map[permissions.Permission]bool, error) {
	user, err := s.users.FindByIDWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.resolvePermsFromUser(user.Roles)
}

func (s *AuthService) resolvePermsFromUser(roles []models.Role) (map[permissions.Permission]bool, error) {
	out := make(map[permissions.Permission]bool, 16)
	for _, r := range roles {
		for _, p := range r.Permissions {
			out[permissions.Permission(p)] = true
		}
	}
	return out, nil
}

// tokenInvalidatedBy reports whether a token issued at iat predates the
// credential-change timestamp ts (nil ts means no change recorded). Same
// semantics as middleware.invalidatedBy.
func tokenInvalidatedBy(ts *time.Time, iat time.Time) bool {
	return ts != nil && iat.Before(ts.UTC())
}

func (s *AuthService) issueTokenPair(userID uuid.UUID, rememberMe bool) (*TokenPair, error) {
	access, err := utils.SignToken(userID.String(), utils.TokenTypeAccess, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		return nil, err
	}
	refreshTTL := s.cfg.RefreshTTL
	if rememberMe && s.cfg.RememberMeRefreshTTL > 0 {
		refreshTTL = s.cfg.RememberMeRefreshTTL
	}
	refresh, err := utils.SignToken(userID.String(), utils.TokenTypeRefresh, s.cfg.JWTSecret, refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}, nil
}
