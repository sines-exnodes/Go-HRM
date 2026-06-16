package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// HolidayHandler handles /api/v1/holidays endpoints.
type HolidayHandler struct {
	svc *services.HolidayService
}

// NewHolidayHandler constructs a HolidayHandler.
func NewHolidayHandler(svc *services.HolidayService) *HolidayHandler {
	return &HolidayHandler{svc: svc}
}

// List godoc
// @Summary      List holidays for a year
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        year       query  int     true  "calendar year (e.g. 2025)"
// @Param        search     query  string  false "name search"
// @Param        page       query  int     false "page number (default 1)"
// @Param        page_size  query  int     false "page size (default 20, max 100)"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.HolidayRead]]
// @Failure      400  {object}  dto.Response[any]
// @Router       /holidays [get]
func (h *HolidayHandler) List(c *gin.Context) {
	var q dto.HolidayListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.List(c.Request.Context(), q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.HolidayRead]]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.HolidayCreate  true  "create payload"
// @Success      201  {object}  dto.Response[dto.HolidayRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /holidays [post]
func (h *HolidayHandler) Create(c *gin.Context) {
	var req dto.HolidayCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, _, aerr := h.svc.Create(c.Request.Context(), req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.HolidayRead]{
		Success: true,
		Message: "Holiday has been created",
		Data:    out,
	})
}

// Update godoc
// @Summary      Partial-update a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string             true  "holiday uuid"
// @Param        body  body  dto.HolidayUpdate  true  "patch payload"
// @Success      200  {object}  dto.Response[dto.HolidayRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /holidays/{id} [patch]
func (h *HolidayHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.HolidayUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, _, aerr := h.svc.Update(c.Request.Context(), id, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.HolidayRead]{
		Success: true,
		Message: "Holiday has been updated",
		Data:    out,
	})
}

// Delete godoc
// @Summary      Soft-delete a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "holiday uuid"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /holidays/{id} [delete]
func (h *HolidayHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	affected, aerr := h.svc.Delete(c.Request.Context(), id)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	msg := "Holiday has been deleted"
	if affected > 0 {
		msg = fmt.Sprintf("Holiday deleted. %d leave request(s) recalculated.", affected)
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: msg})
}

// GetYears godoc
// @Summary      List years that have holidays
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[[]int]
// @Router       /holidays/years [get]
func (h *HolidayHandler) GetYears(c *gin.Context) {
	years, aerr := h.svc.GetYears(c.Request.Context())
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]int]{Success: true, Data: years})
}

// ListTemplates godoc
// @Summary      List Vietnamese holiday presets for a year
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        year  query  int  true  "calendar year"
// @Success      200  {object}  dto.Response[[]dto.HolidayTemplateRead]
// @Router       /holidays/templates [get]
func (h *HolidayHandler) ListTemplates(c *gin.Context) {
	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		_ = c.Error(apperrors.ErrBadRequest("year query param must be a valid year (2000-2100)"))
		return
	}
	out, aerr := h.svc.ListTemplates(c.Request.Context(), year)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.HolidayTemplateRead]{Success: true, Data: out})
}

// Import godoc
// @Summary      Import selected holiday presets into a year
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.HolidayImportRequest  true  "import payload"
// @Success      200  {object}  dto.Response[dto.HolidayImportResult]
// @Failure      400  {object}  dto.Response[any]
// @Router       /holidays/import [post]
func (h *HolidayHandler) Import(c *gin.Context) {
	var req dto.HolidayImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Import(c.Request.Context(), req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	msg := fmt.Sprintf("%d holiday(s) imported for %d", out.Imported, req.Year)
	if out.Skipped > 0 {
		msg = fmt.Sprintf("%d holiday(s) imported for %d, %d skipped (already exist)", out.Imported, req.Year, out.Skipped)
	}
	c.JSON(http.StatusOK, dto.Response[*dto.HolidayImportResult]{Success: true, Message: msg, Data: out})
}
