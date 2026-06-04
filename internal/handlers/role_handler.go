package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// RoleHandler handles role-management endpoints.
type RoleHandler struct {
	svc *services.RoleService
}

// NewRoleHandler constructs a RoleHandler.
func NewRoleHandler(svc *services.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }

// ListPermissions godoc
// @Summary      List grouped permissions (for the role-creation picker)
// @Description  Returns the structured permission catalog. Requires roles:read.
// @Tags         Roles
// @Produce      json
// @Success      200  {object}  dto.Response[[]permissions.PermissionGroup]
// @Security     BearerAuth
// @Router       /api/v1/roles/permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response[[]permissions.PermissionGroup]{
		Success: true,
		Data:    permissions.PermissionGroups,
	})
}

// List godoc
// @Summary      List roles (paginated, name search)
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Param        search     query  string  false  "Substring match on name (ILIKE)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var q dto.RoleListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.RoleRead]]{Success: true, Data: data})
}

// Get godoc
// @Summary      Get role by ID
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Role UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.RoleRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RoleCreate  true  "Role payload"
// @Success      201  {object}  map[string]interface{}
// @Router       /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var in dto.RoleCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.RoleRead]{Success: true, Message: "Role created successfully", Data: out})
}

// Update godoc
// @Summary      Update role (PATCH semantics)
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string          true  "Role UUID"
// @Param        body  body  dto.RoleUpdate  true  "Fields to update"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [patch]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	var in dto.RoleUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.RoleRead]{Success: true, Message: "Role updated successfully", Data: out})
}

// Delete godoc
// @Summary      Delete a non-system role
// @Description  Soft-deletes the role (name becomes reusable). Rejected with 400
// @Description  for system roles and 409 if any user is still assigned.
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Role UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Role deleted"})
}
