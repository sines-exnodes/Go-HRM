package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

type DependentHandler struct {
	svc *services.DependentService
}

func NewDependentHandler(svc *services.DependentService) *DependentHandler {
	return &DependentHandler{svc: svc}
}

// hasManageAll inspects the user's loaded roles for PermDependentsManage or wildcard.
func hasManageAll(c *gin.Context) bool {
	u, okC := currentUser(c)
	if !okC {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermDependentsManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// List godoc
// @Summary      List dependents for an employee
// @Tags         dependents
// @Security     BearerAuth
// @Produce      json
// @Param        id         path string true "employee uuid"
// @Param        page       query int  false "page"
// @Param        page_size  query int  false "page size"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/dependents [get]
func (h *DependentHandler) List(c *gin.Context) {
	employeeID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	if err := h.svc.AuthorizeOwnerOrAdmin(c.Request.Context(), u.ID, employeeID, hasManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	var q dto.DependentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), employeeID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusOK, gin.H{"items": items, "total": total, "page": q.Page, "page_size": q.PageSize}, "")
}

// Create godoc
// @Summary      Create a dependent for an employee
// @Tags         dependents
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id         path string true "employee uuid"
// @Param        body body dto.DependentCreate true "dependent fields"
// @Success      201 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/dependents [post]
func (h *DependentHandler) Create(c *gin.Context) {
	employeeID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	if err := h.svc.AuthorizeOwnerOrAdmin(c.Request.Context(), u.ID, employeeID, hasManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	var in dto.DependentCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.Create(c.Request.Context(), employeeID, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusCreated, view, "Dependent created")
}

// Update godoc
// @Summary      Update a dependent
// @Tags         dependents
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id           path string true "employee uuid"
// @Param        dependentID  path string true "dependent uuid"
// @Param        body body dto.DependentUpdate true "fields"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/dependents/{dependentID} [patch]
func (h *DependentHandler) Update(c *gin.Context) {
	employeeID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	id, err := parseIDParam(c, "dependentID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	if err := h.svc.AuthorizeOwnerOrAdmin(c.Request.Context(), u.ID, employeeID, hasManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	// Sanity: the dependent must belong to the URL's employeeID.
	owner, err := h.svc.OwnerEmployeeIDForDependent(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if owner != employeeID {
		_ = c.Error(apperrors.ErrNotFound("Dependent"))
		return
	}
	var in dto.DependentUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusOK, view, "Dependent updated")
}

// Delete godoc
// @Summary      Delete a dependent
// @Tags         dependents
// @Security     BearerAuth
// @Produce      json
// @Param        id           path string true "employee uuid"
// @Param        dependentID  path string true "dependent uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/dependents/{dependentID} [delete]
func (h *DependentHandler) Delete(c *gin.Context) {
	employeeID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	id, err := parseIDParam(c, "dependentID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	if err := h.svc.AuthorizeOwnerOrAdmin(c.Request.Context(), u.ID, employeeID, hasManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	owner, err := h.svc.OwnerEmployeeIDForDependent(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if owner != employeeID {
		_ = c.Error(apperrors.ErrNotFound("Dependent"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Dependent deleted")
}
