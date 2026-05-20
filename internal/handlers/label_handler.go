package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

type LabelHandler struct {
	svc *services.LabelService
}

func NewLabelHandler(svc *services.LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

// List godoc
// @Summary      List announcement labels
// @Description  Returns every live label sorted by name ASC (no pagination).
// @Tags         labels
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/announcement-labels [get]
func (h *LabelHandler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.LabelRead]{Success: true, Data: out})
}

// GetOrCreate godoc
// @Summary      Get-or-create an announcement label by name
// @Description  Idempotent. Case-insensitive lookup: returns the existing
// @Description  label with HTTP 200 if one matches, otherwise inserts and
// @Description  returns the new row with HTTP 201.
// @Tags         labels
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LabelCreate  true  "Label name"
// @Success      201   {object}  map[string]interface{}
// @Router       /api/v1/announcement-labels [post]
func (h *LabelHandler) GetOrCreate(c *gin.Context) {
	var in dto.LabelCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	res, err := h.svc.GetOrCreate(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	status := http.StatusOK
	message := "Label already exists"
	if res.Created {
		status = http.StatusCreated
		message = "Label created"
	}
	c.JSON(status, dto.Response[*dto.LabelRead]{
		Success: true,
		Message: message,
		Data:    &res.Label,
	})
}
