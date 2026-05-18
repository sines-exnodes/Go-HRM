package middleware

import (
	"github.com/gin-gonic/gin"

	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// RequirePerms returns Gin middleware that gates the route on the user's
// effective permission set (union of permissions across their roles).
//
// Semantics:
//   - JWT middleware must run first (it sets the user on the context).
//   - Wildcard "*" bypasses all checks.
//   - ALL listed permissions must be present in the user's set.
//   - On failure, returns ErrForbidden with details {required, missing}.
func RequirePerms(authSvc *services.AuthService, required ...permissions.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserFromContext(c)
		if user == nil {
			_ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
			c.Abort()
			return
		}

		perms, err := authSvc.ResolveUserPermissions(c.Request.Context(), user.ID)
		if err != nil {
			_ = c.Error(apperr.ErrUnauthorized("Failed to resolve permissions"))
			c.Abort()
			return
		}

		if perms[permissions.PermAll] {
			c.Next()
			return
		}

		missing := make([]string, 0)
		for _, p := range required {
			if !perms[p] {
				missing = append(missing, string(p))
			}
		}
		if len(missing) > 0 {
			reqList := make([]string, 0, len(required))
			for _, p := range required {
				reqList = append(reqList, string(p))
			}
			err := apperr.ErrForbidden("Insufficient permissions")
			err.Details = map[string]any{
				"required": reqList,
				"missing":  missing,
			}
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}
