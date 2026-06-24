package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/services"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	auth  *services.AuthService
	reset *services.PasswordResetService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(auth *services.AuthService, reset *services.PasswordResetService) *AuthHandler {
	return &AuthHandler{auth: auth, reset: reset}
}

// toUserSummary projects an auth-loaded User (with Roles + Employee preloaded)
// into the auth response shape.
func toUserSummary(u *models.User) dto.UserSummary {
	var emp *dto.EmployeeSummary
	if u.Employee != nil {
		emp = &dto.EmployeeSummary{
			ID:           u.Employee.ID,
			FirstName:    u.Employee.FirstName,
			LastName:     u.Employee.LastName,
			AvatarURL:    u.Employee.AvatarURL,
			DepartmentID: u.Employee.DepartmentID,
			PositionID:   u.Employee.PositionID,
			ManagerID:    u.Employee.ManagerID,
		}
	}
	roles := make([]dto.RoleSummary, 0, len(u.Roles))
	for _, r := range u.Roles {
		perms := make([]string, 0, len(r.Permissions))
		for _, p := range r.Permissions {
			perms = append(perms, p)
		}
		roles = append(roles, dto.RoleSummary{
			ID:          r.ID,
			Name:        r.Name,
			IsSystem:    r.IsSystem,
			Permissions: perms,
		})
	}
	return dto.UserSummary{
		ID:       u.ID,
		Email:    u.Email,
		IsActive: u.IsActive,
		Employee: emp,
		Roles:    roles,
	}
}

// Login godoc
// @Summary      Authenticate and receive access + refresh tokens
// @Description  Exchanges an email + password for a token pair. Required permission: `auth:login`.
// @Description
// @Description  **Brute-force protection.** After `MAX_FAILED_LOGIN_ATTEMPTS` (default 5)
// @Description  consecutive bad passwords the account is locked for `ACCOUNT_LOCKOUT_MINUTES`
// @Description  (default 15). During the lockout window, every login attempt — including the
// @Description  correct one — returns 401 with `"Account temporarily locked. Try again in N minutes."`.
// @Description  A successful login resets the failed-attempt counter.
// @Description
// @Description  **`remember_me`.** When `true`, the refresh token is issued with the long-lived
// @Description  TTL (`REMEMBER_ME_REFRESH_TOKEN_EXPIRE_DAYS`, default 30 days) instead of the
// @Description  default refresh-token TTL. The access token TTL is unaffected.
// @Description
// @Description  **`is_active`.** A deactivated account is rejected with 401 *after* password
// @Description  verification so the response cannot be used to enumerate which accounts exist.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.Response[dto.LoginResponse]
// @Failure      400   {object}  dto.Response[any]  "Malformed body"
// @Failure      401   {object}  dto.Response[any]  "Invalid credentials, account deactivated, or temporarily locked"
// @Failure      403   {object}  dto.Response[any]  "Missing auth:login permission"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, req.RememberMe)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.LoginResponse]{
		Success: true,
		Message: "Login successful",
		Data: dto.LoginResponse{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    "Bearer",
			User:         toUserSummary(result.User),
		},
	})
}

// Refresh godoc
// @Summary      Exchange a refresh token for a new token pair
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  dto.Response[dto.LoginResponse]
// @Failure      400   {object}  dto.Response[any]
// @Failure      401   {object}  dto.Response[any]
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	result, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.LoginResponse]{
		Success: true,
		Message: "Token refreshed",
		Data: dto.LoginResponse{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    "Bearer",
			User:         toUserSummary(result.User),
		},
	})
}

// Logout godoc
// @Summary      Acknowledge logout
// @Description  Stateless logout — the client must discard its tokens.
// @Tags         Authentication
// @Produce      json
// @Success      200   {object}  dto.Response[dto.LogoutResponse]
// @Failure      401   {object}  dto.Response[any]
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	user := middleware.UserFromContext(c)
	if user == nil {
		_ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
		return
	}
	if err := h.auth.Logout(c.Request.Context(), user.ID); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.LogoutResponse]{
		Success: true,
		Message: "Logged out",
		Data:    dto.LogoutResponse{Message: "Logged out"},
	})
}

// ForgotPassword godoc
// @Summary      Request a password reset email
// @Description  Sends a password reset link to the given email address.
// @Description  Always returns 200 regardless of whether the email is registered
// @Description  (enumerate guard — never reveal which accounts exist).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ForgotPasswordRequest  true  "Email address"
// @Success      200   {object}  dto.Response[any]
// @Failure      400   {object}  dto.Response[any]  "Malformed body"
// @Router       /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	// RequestReset is always nil — enumerate guard.
	_ = h.reset.RequestReset(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, dto.Response[any]{
		Success: true,
		Message: "If that email address is registered, a reset link has been sent.",
	})
}

// VerifyResetToken godoc
// @Summary      Verify a password-reset token
// @Description  Checks whether the token is valid (exists, not used, not expired). Does not consume the token.
// @Description  Call this when the user lands on the reset-password page to show an error before they fill in the form.
// @Tags         Authentication
// @Produce      json
// @Param        token  query     string  true  "Reset token from the email link"
// @Success      200    {object}  dto.Response[any]
// @Failure      400    {object}  dto.Response[any]  "Invalid, expired, or already-used token"
// @Router       /api/v1/auth/verify-token [get]
func (h *AuthHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		_ = c.Error(apperr.ErrBadRequest("token query parameter is required"))
		return
	}
	user, err := h.reset.VerifyToken(c.Request.Context(), token)
	if err != nil {
		_ = c.Error(err)
		return
	}
	firstName := ""
	if user.Employee != nil {
		firstName = user.Employee.FirstName
	}
	c.JSON(http.StatusOK, dto.Response[dto.TokenVerifyResponse]{
		Success: true,
		Message: "Token is valid.",
		Data: dto.TokenVerifyResponse{
			Email:     user.Email,
			FirstName: firstName,
		},
	})
}

// ResetPassword godoc
// @Summary      Reset password using a token
// @Description  Validates the one-time token and sets a new password.
// @Description  Returns 400 if the token is missing, expired, or already used.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ResetPasswordRequest  true  "Reset token and new password"
// @Success      200   {object}  dto.Response[any]
// @Failure      400   {object}  dto.Response[any]  "Invalid/expired token or bad request"
// @Router       /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	if err := h.reset.ResetWithToken(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{
		Success: true,
		Message: "Password has been reset successfully.",
	})
}
