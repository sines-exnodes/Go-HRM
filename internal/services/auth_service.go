package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// AuthConfig configures token TTLs and secret.
type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
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
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrUnauthorized("Invalid email or password")
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, apperr.ErrUnauthorized("Your account has been deactivated. Contact your administrator.")
	}
	if user.PasswordHash == "" {
		return nil, apperr.ErrUnauthorized("Please set your password using the invite link sent to your email.")
	}
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, apperr.ErrUnauthorized("Invalid email or password")
	}

	perms, err := s.resolvePermsFromUser(user.Roles)
	if err != nil {
		return nil, err
	}
	if !perms[permissions.PermAll] && !perms[permissions.PermAuthLogin] {
		return nil, apperr.ErrForbidden("You do not have permission to access this system.")
	}

	tokens, err := s.issueTokenPair(user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Tokens: *tokens, User: user}, nil
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

	tokens, err := s.issueTokenPair(user.ID)
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

func (s *AuthService) issueTokenPair(userID uuid.UUID) (*TokenPair, error) {
	access, err := utils.SignToken(userID.String(), utils.TokenTypeAccess, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := utils.SignToken(userID.String(), utils.TokenTypeRefresh, s.cfg.JWTSecret, s.cfg.RefreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}, nil
}
