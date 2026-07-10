package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/services"
)

type DashboardHandler struct {
	svc *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Get godoc
// @Summary      Get the common dashboard
// @Description  Returns fixed-order widgets visible to the authenticated user based on source module permissions.
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[dto.DashboardRead]
// @Router       /api/v1/dashboard [get]
func (h *DashboardHandler) Get(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	out, err := h.svc.Get(c.Request.Context(), u)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.DashboardRead]{Success: true, Data: out})
}
