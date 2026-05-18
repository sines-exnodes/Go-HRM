package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

type PositionHandler struct {
	svc *services.PositionService
}

func NewPositionHandler(svc *services.PositionService) *PositionHandler {
	return &PositionHandler{svc: svc}
}

// List godoc
// @Summary      List positions
// @Tags         positions
// @Security     BearerAuth
// @Produce      json
// @Param        page           query    int     false  "Page"        default(1)
// @Param        page_size      query    int     false  "Page size"   default(10)
// @Param        search         query    string  false  "Substring match on name"
// @Param        department_id  query    string  false  "Filter by department UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/positions [get]
func (h *PositionHandler) List(c *gin.Context) {
	var q dto.PositionListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.PositionRead]]{Success: true, Data: data})
}

// Create godoc
// @Summary      Create position
// @Tags         positions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.PositionCreate  true  "Position payload"
// @Success      201   {object}  map[string]interface{}
// @Router       /api/v1/positions [post]
func (h *PositionHandler) Create(c *gin.Context) {
	var in dto.PositionCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.PositionRead]{
		Success: true,
		Message: "Position created",
		Data:    out,
	})
}

// Get godoc
// @Summary      Get position by ID
// @Tags         positions
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Position UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/positions/{id} [get]
func (h *PositionHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.Response[*dto.PositionRead]{Success: true, Data: out})
}

// Update godoc
// @Summary      Update position
// @Tags         positions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string              true  "Position UUID"
// @Param        body  body      dto.PositionUpdate  true  "Fields to update"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/positions/{id} [patch]
func (h *PositionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	var in dto.PositionUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PositionRead]{
		Success: true,
		Message: "Position updated",
		Data:    out,
	})
}

// Delete godoc
// @Summary      Delete position
// @Description  Soft-deletes a position. Rejected with 409 if any employee is still assigned.
// @Tags         positions
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Position UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/positions/{id} [delete]
func (h *PositionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Position deleted"})
}
