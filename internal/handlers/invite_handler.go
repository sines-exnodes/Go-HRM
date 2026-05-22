package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// InviteHandler owns /api/v1/invites. The 5 admin endpoints are gated
// by RequirePerms(PermInviteManage); POST /accept is public (the token
// is the credential).
type InviteHandler struct {
	svc *services.InviteService
}

func NewInviteHandler(svc *services.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

// List godoc
// @Summary      List invites
// @Tags         invites
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int     false  "page"
// @Param        page_size  query  int     false  "page size"
// @Param        email      query  string  false  "filter by email substring"
// @Param        status     query  string  false  "pending|accepted|expired"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/invites [get]
func (h *InviteHandler) List(c *gin.Context) {
	var q dto.InviteListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.InviteRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary      Get an invite by ID
// @Tags         invites
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "invite uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/invites/{id} [get]
func (h *InviteHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.InviteRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Issue a new invite
// @Description  Generates a token, attempts to send the invite email, and returns the invite. SMTP misconfiguration is non-fatal — check last_email_error on the response to see if the email was actually sent.
// @Tags         invites
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.InviteCreate  true  "invite payload"
// @Success      201   {object}  map[string]interface{}
// @Failure      409   {object}  map[string]interface{}
// @Router       /api/v1/invites [post]
func (h *InviteHandler) Create(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.InviteCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.InviteRead]{Success: true, Message: "Invite created", Data: out})
}

// Resend godoc
// @Summary      Resend the invite email (same token)
// @Tags         invites
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "invite uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/invites/{id}/resend [post]
func (h *InviteHandler) Resend(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Resend(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.InviteRead]{Success: true, Message: "Invite resent", Data: out})
}

// Revoke godoc
// @Summary      Revoke an invite (soft delete)
// @Tags         invites
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "invite uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/invites/{id} [delete]
func (h *InviteHandler) Revoke(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Invite revoked"})
}

// Accept godoc
// @Summary      Accept an invite — creates the user (public)
// @Description  Token comes from the email link. Public endpoint — no JWT required (the token is the credential).
// @Tags         invites
// @Accept       json
// @Produce      json
// @Param        body  body  dto.InviteAccept  true  "accept payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Router       /api/v1/invites/accept [post]
func (h *InviteHandler) Accept(c *gin.Context) {
	var req dto.InviteAccept
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Accept(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.InviteAcceptResult]{Success: true, Data: out})
}
