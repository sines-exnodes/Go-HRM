package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
)

// HealthHandler is a stateless liveness probe.
type HealthHandler struct{}

// NewHealthHandler returns a ready-to-use HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthData is the payload of GET /health.
type HealthData struct {
	Status  string `json:"status"  example:"ok"`
	Service string `json:"service" example:"exnodes-hrm-api"`
}

// Health godoc
// @Summary      Liveness probe
// @Description  Returns {success: true, data: {status: "ok"}} when the server is up.
// @Tags         system
// @Produce      json
// @Success      200  {object}  dto.Response[handlers.HealthData]
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewResponse(HealthData{
		Status:  "ok",
		Service: "exnodes-hrm-api",
	}, ""))
}
