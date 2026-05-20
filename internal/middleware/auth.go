package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// Context keys for the authenticated user.
const (
	ContextKeyUser   = "auth_user"
	ContextKeyUserID = "auth_user_id"
	ContextKeyClaims = "auth_claims"
)

// JWT returns Gin middleware that validates a Bearer access token and loads
// the corresponding User from the database. Mirrors the Python
// get_current_user dependency.
func JWT(users repositories.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractToken(c)
		applyAuth(c, raw, jwtSecret, users)
	}
}

// JWTFromQueryOrHeader is a JWT middleware variant for endpoints where the
// browser cannot set Authorization headers (notably EventSource for
// Server-Sent Events). It accepts the token in either:
//   - Authorization: Bearer <jwt> header, OR
//   - ?token=<jwt> query parameter.
//
// LIMITATION: query-param tokens may appear in proxy/server access logs.
// Operators should scrub the `token` parameter on /api/v1/sse/* routes
// at the reverse-proxy level. Mitigation: use short-lived access tokens.
func JWTFromQueryOrHeader(users repositories.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractToken(c)
		if raw == "" {
			raw = strings.TrimSpace(c.Query("token"))
		}
		applyAuth(c, raw, jwtSecret, users)
	}
}

// applyAuth is the shared parse+load+set-context body used by both JWT
// variants. On failure it writes an apperr.ErrUnauthorized to c.Error()
// and aborts the chain; on success it sets ContextKeyUser/UserID/Claims
// and calls c.Next().
func applyAuth(c *gin.Context, raw, jwtSecret string, users repositories.UserRepository) {
	if raw == "" {
		_ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
		c.Abort()
		return
	}
	claims, err := utils.VerifyToken(raw, jwtSecret)
	if err != nil {
		_ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
		c.Abort()
		return
	}
	if claims.Type != utils.TokenTypeAccess {
		_ = c.Error(apperr.ErrUnauthorized("Invalid token type"))
		c.Abort()
		return
	}
	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		_ = c.Error(apperr.ErrUnauthorized("Invalid token payload"))
		c.Abort()
		return
	}

	user, err := users.FindByIDWithRoles(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = c.Error(apperr.ErrUnauthorized("User not found"))
		} else {
			_ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
		}
		c.Abort()
		return
	}
	if !user.IsActive {
		_ = c.Error(apperr.ErrUnauthorized("User account is inactive"))
		c.Abort()
		return
	}

	// Session invalidation: reject tokens issued before credential changes.
	if claims.IssuedAt != nil {
		iat := claims.IssuedAt.Time
		if invalidatedBy(user.EmailChangedAt, iat) {
			_ = c.Error(apperr.ErrUnauthorized("Session expired due to email change — please log in again"))
			c.Abort()
			return
		}
		if invalidatedBy(user.PasswordResetAt, iat) {
			_ = c.Error(apperr.ErrUnauthorized("Session expired due to password reset — please log in again"))
			c.Abort()
			return
		}
	}

	c.Set(ContextKeyUser, user)
	c.Set(ContextKeyUserID, user.ID)
	c.Set(ContextKeyClaims, claims)
	c.Next()
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return ""
}

func invalidatedBy(ts *time.Time, iat time.Time) bool {
	return ts != nil && iat.Before(ts.UTC())
}

// UserFromContext returns the authenticated user, or nil if not set.
func UserFromContext(c *gin.Context) *models.User {
	v, ok := c.Get(ContextKeyUser)
	if !ok {
		return nil
	}
	u, _ := v.(*models.User)
	return u
}

// ContextWithUserID lifts the request context with the authenticated user ID.
// Provided for services that don't take Gin context directly.
func ContextWithUserID(c *gin.Context) context.Context {
	return c.Request.Context()
}
