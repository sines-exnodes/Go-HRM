package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

type DepartmentHandler struct {
	svc *services.DepartmentService
}

func NewDepartmentHandler(svc *services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

// List godoc
// @Summary      List departments
// @Description  Paginated list with optional name search and parent filter ("root" returns top-level only).
// @Tags         departments
// @Security     BearerAuth
// @Produce      json
// @Param        page       query    int     false  "Page number"  default(1)
// @Param        page_size  query    int     false  "Page size"    default(10)
// @Param        search     query    string  false  "Substring match on name (ILIKE)"
// @Param        parent_id  query    string  false  "Filter by parent UUID, or \"root\" for top-level"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/departments [get]
func (h *DepartmentHandler) List(c *gin.Context) {
	var q dto.DepartmentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.DepartmentRead]]{Success: true, Data: data})
}

// Create godoc
// @Summary      Create department
// @Tags         departments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.DepartmentCreate  true  "Department payload"
// @Success      201   {object}  map[string]interface{}
// @Router       /api/v1/departments [post]
func (h *DepartmentHandler) Create(c *gin.Context) {
	var in dto.DepartmentCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.DepartmentRead]{
		Success: true,
		Message: "Department created",
		Data:    out,
	})
}

// Get godoc
// @Summary      Get department by ID
// @Tags         departments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Department UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/departments/{id} [get]
func (h *DepartmentHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.Response[*dto.DepartmentRead]{Success: true, Data: out})
}

// Update godoc
// @Summary      Update department
// @Tags         departments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string                true  "Department UUID"
// @Param        body  body      dto.DepartmentUpdate  true  "Fields to update (PATCH semantics)"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/departments/{id} [patch]
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	var in dto.DepartmentUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.DepartmentRead]{
		Success: true,
		Message: "Department updated",
		Data:    out,
	})
}

// Delete godoc
// @Summary      Delete department
// @Description  Soft-deletes a department. Rejected with 409 if it has child departments, active positions, or assigned employees.
// @Tags         departments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Department UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Department deleted"})
}
