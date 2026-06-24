package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

type UserHandler struct {
	svc   *services.UserService
	reset *services.PasswordResetService
}

func NewUserHandler(svc *services.UserService, reset *services.PasswordResetService) *UserHandler {
	return &UserHandler{svc: svc, reset: reset}
}

// GetMe godoc
// @Summary      Get current user (auth profile + embedded employee summary)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	view, err := h.svc.GetMe(c.Request.Context(), u)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusOK, view, "")
}

// List godoc
// @Summary      List users (admin)
// @Description  Paginated list with optional email search and is_active filter. Each item embeds roles and the employee summary.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page       query    int     false  "Page number"  default(1)
// @Param        page_size  query    int     false  "Page size"    default(10)
// @Param        search     query    string  false  "Substring match on email (ILIKE)"
// @Param        is_active  query    bool    false  "Filter by active status"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q dto.UserListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.UserAdminRead]]{Success: true, Data: data})
}

// Get godoc
// @Summary      Get user by ID (admin)
// @Description  Returns the user with roles and embedded employee summary.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.Response[*dto.UserAdminRead]{Success: true, Data: out})
}

// ChangeMyPassword godoc
// @Summary      Change current user password
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePasswordRequest true "current and new password"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me/change-password [post]
func (h *UserHandler) ChangeMyPassword(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if in.CurrentPassword == "" {
		_ = c.Error(apperrors.ErrBadRequest("Current password is required"))
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), u, in.CurrentPassword, in.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Password changed successfully")
}

// ChangeMyEmail godoc
// @Summary      Change current user email
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangeEmailRequest true "new email + current password"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me/change-email [post]
func (h *UserHandler) ChangeMyEmail(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.ChangeEmailRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.ChangeEmail(c.Request.Context(), u, in.NewEmail, in.CurrentPassword)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusOK, view, "Email updated successfully")
}

// RegisterDeviceToken godoc
// @Summary      Register a device token for push notifications
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.FcmTokenRequest true "device id + token"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me/device-tokens [post]
func (h *UserHandler) RegisterDeviceToken(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.FcmTokenRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.RegisterDeviceToken(c.Request.Context(), u.ID, in); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Device token registered")
}

// RemoveDeviceToken godoc
// @Summary      Remove a device token
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        token path string true "FCM token"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me/device-tokens/{token} [delete]
func (h *UserHandler) RemoveDeviceToken(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	token := c.Param("token")
	if token == "" {
		_ = c.Error(apperrors.ErrBadRequest("token is required"))
		return
	}
	if err := h.svc.RemoveDeviceToken(c.Request.Context(), u.ID, token); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Device token removed")
}

// UpdateMyNotificationSettings godoc
// @Summary      Update push notification settings
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.NotificationSettingsRequest true "settings"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/me/notification-settings [patch]
func (h *UserHandler) UpdateMyNotificationSettings(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.NotificationSettingsRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.UpdateNotificationSettings(c.Request.Context(), u.ID, in.NotificationsEnabled); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Notification settings updated")
}

// ---- Admin user endpoints ----

// AdminChangePassword godoc
// @Summary      Admin resets a user's password
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "user uuid"
// @Param        body body dto.ChangePasswordRequest true "new password (current ignored)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/{id}/change-password [patch]
func (h *UserHandler) AdminChangePassword(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var in dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AdminChangePassword(c.Request.Context(), id, in.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Password changed")
}

// AdminPatch godoc
// @Summary      Admin toggle is_active on a user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "user uuid"
// @Param        body body dto.AdminUserPatch true "fields"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/{id} [patch]
func (h *UserHandler) AdminPatch(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	admin, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.AdminUserPatch
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AdminPatch(c.Request.Context(), id, in, admin); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "User updated")
}

// AdminChangeEmail godoc
// @Summary      Admin changes a user's email (employees parity #13)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "user uuid"
// @Param        body body dto.AdminChangeEmailRequest true "new email"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/{id}/change-email [post]
func (h *UserHandler) AdminChangeEmail(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var in dto.AdminChangeEmailRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.AdminChangeEmail(c.Request.Context(), id, in.NewEmail)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ok(c, http.StatusOK, view, "User email updated")
}

// AdminDelete godoc
// @Summary      Soft-delete a user account (admin reauth)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "user uuid"
// @Param        body body dto.DeleteUserRequest true "admin's current password"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) AdminDelete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	admin, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.DeleteUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), id, admin, in.CurrentPassword); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "User deleted")
}

// AssignRoles godoc
// @Summary      Assign roles to a user (admin)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "user uuid"
// @Param        body body dto.RoleAssignmentRequest true "role ids"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/users/{id}/roles [put]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	admin, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.RoleAssignmentRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AssignRoles(c.Request.Context(), id, in.RoleIDs, admin); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Roles assigned")
}

// AdminSendResetLink godoc
// @Summary      Admin — send password-reset link to a user
// @Description  Triggers a "reset your password" email for the given user. Same token flow as forgot-password.
// @Description  Requires users:change_password permission.
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      string  true  "User UUID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /api/v1/users/{id}/reset-password [post]
func (h *UserHandler) AdminSendResetLink(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	_ = h.reset.RequestReset(c.Request.Context(), user.Email)
	okEmpty(c, "Password reset link sent to "+user.Email)
}

// AdminSendInvite godoc
// @Summary      Admin — send (or resend) set-password invite to a user
// @Description  Sends a "set your password" invite email. Useful for resending the initial invite.
// @Description  Requires users:create permission.
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      string  true  "User UUID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /api/v1/users/{id}/send-invite [post]
func (h *UserHandler) AdminSendInvite(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	_ = h.reset.RequestReset(c.Request.Context(), user.Email)
	okEmpty(c, "Invite sent to "+user.Email)
}
