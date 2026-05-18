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
	auth *services.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// toUserSummary projects an auth-loaded User (with Roles + Employee preloaded)
// into the auth response shape.
func toUserSummary(u *models.User) dto.UserSummary {
	var emp *dto.EmployeeSummary
	if u.Employee != nil {
		emp = &dto.EmployeeSummary{
			ID:           u.Employee.ID,
			FullName:     u.Employee.FullName,
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
// @Description  Exchanges an email + password for a token pair. Required permission: auth:login.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.Response[dto.LoginResponse]
// @Failure      400   {object}  dto.Response[any]
// @Failure      401   {object}  dto.Response[any]
// @Failure      403   {object}  dto.Response[any]
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
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
