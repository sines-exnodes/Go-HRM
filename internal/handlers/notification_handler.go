package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// NotificationHandler owns /api/v1/notifications/test — admin debug
// endpoint that pushes to the caller's own registered device tokens.
type NotificationHandler struct {
	svc *services.PushNotificationService
}

func NewNotificationHandler(svc *services.PushNotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// SendTest godoc
// @Summary      Send a test push notification to the caller's devices
// @Description  Admin debug endpoint. Looks up the caller's registered device tokens and dispatches the payload to each. When FCM is not configured, every token counts as 'skipped'.
// @Tags         notifications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.NotificationTestRequest  true  "push payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/notifications/test [post]
func (h *NotificationHandler) SendTest(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.NotificationTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.SendToUser(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.NotificationTestResult]{Success: true, Data: out})
}
