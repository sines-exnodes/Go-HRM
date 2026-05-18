package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/permissions"
)

// RoleHandler handles role-related endpoints. Phase 1 ships only the
// permission catalog endpoint; full role CRUD comes in Phase 2.
type RoleHandler struct{}

// NewRoleHandler constructs a RoleHandler.
func NewRoleHandler() *RoleHandler { return &RoleHandler{} }

// ListPermissions godoc
// @Summary      List grouped permissions (for the role-creation picker)
// @Description  Returns the structured permission catalog. Requires authentication only (no specific permission).
// @Tags         Roles
// @Produce      json
// @Success      200  {object}  dto.Response[[]permissions.PermissionGroup]
// @Failure      401  {object}  dto.Response[any]
// @Security     BearerAuth
// @Router       /api/v1/roles/permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response[[]permissions.PermissionGroup]{
		Success: true,
		Data:    permissions.PermissionGroups,
	})
}
