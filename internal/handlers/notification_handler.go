package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// NotificationHandler owns the mobile in-app notification feed
// (/api/v1/mobile/notifications) plus /api/v1/notifications/test — the admin
// debug endpoint that pushes to the caller's own registered device tokens.
type NotificationHandler struct {
	svc      *services.PushNotificationService
	notifSvc *services.NotificationService
}

func NewNotificationHandler(
	svc *services.PushNotificationService,
	notifSvc *services.NotificationService,
) *NotificationHandler {
	return &NotificationHandler{svc: svc, notifSvc: notifSvc}
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

// List godoc
// @Summary      List the caller's in-app notifications
// @Description  Reverse-chronological feed of the authenticated employee's notifications. Always scoped to the caller — there is no way to read another employee's feed.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "page number (default 1)"
// @Param        page_size  query  int  false  "page size (default 50, max 100)"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.NotificationRead]]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Router       /mobile/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var q dto.NotificationListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.notifSvc.List(c.Request.Context(), u.ID, q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.NotificationRead]]{Success: true, Data: out})
}

// UnreadCount godoc
// @Summary      Count the caller's unread notifications
// @Description  Backs the dashboard header notification bell. Cheap to poll after marking a notification read.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[dto.NotificationUnreadCountRead]
// @Failure      401  {object}  dto.Response[any]
// @Router       /mobile/notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	out, aerr := h.notifSvc.UnreadCount(c.Request.Context(), u.ID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.NotificationUnreadCountRead]{Success: true, Data: out})
}

// MarkRead godoc
// @Summary      Mark a notification read
// @Description  Marks the notification read for the caller and returns it. Read is terminal — re-marking an already-read notification is a 200 no-op, not a conflict. A notification belonging to another employee returns 404, indistinguishable from a missing one.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "notification id"
// @Success      200  {object}  dto.Response[dto.NotificationRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /mobile/notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	// NOTE: the parseIDParam result is named `err`, not `aerr`, and this is
	// load-bearing. Reusing one variable across parseIDParam (returns `error`)
	// and a service call (returns *apperrors.AppError) boxes a typed-nil into
	// a non-nil interface and panics the error middleware — the exact bug
	// already found and fixed once in the holiday handler.
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, aerr := h.notifSvc.MarkRead(c.Request.Context(), id, u.ID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.NotificationRead]{Success: true, Data: out})
}
