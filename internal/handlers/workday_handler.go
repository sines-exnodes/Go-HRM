package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// WorkdayHandler handles /api/v1/workdays endpoints.
type WorkdayHandler struct {
	svc *services.WorkdayService
}

// NewWorkdayHandler constructs a WorkdayHandler.
func NewWorkdayHandler(svc *services.WorkdayService) *WorkdayHandler {
	return &WorkdayHandler{svc: svc}
}

// GetYear godoc
// @Summary      Monthly workday summary for a year
// @Description  Returns the workday count for each month of the given year, computed live from the company holiday calendar. Workdays = Total Days − Weekends − Holidays. A holiday falling on a weekend still reduces Workdays.
// @Tags         workdays
// @Security     BearerAuth
// @Produce      json
// @Param        year  query  int  true  "calendar year (2000–2100)"
// @Success      200  {object}  dto.Response[dto.WorkdayYearRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Router       /workdays [get]
func (h *WorkdayHandler) GetYear(c *gin.Context) {
	var q dto.WorkdayQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.GetYear(c.Request.Context(), q.Year)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.WorkdayYearRead]{Success: true, Data: out})
}
