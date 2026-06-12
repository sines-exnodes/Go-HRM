package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// UserContractHandler owns /api/v1/users/:id/contracts.
type UserContractHandler struct {
	svc *services.UserContractService
}

func NewUserContractHandler(svc *services.UserContractService) *UserContractHandler {
	return &UserContractHandler{svc: svc}
}

// hasContractsManage walks the JWT-preloaded roles for PermUsersContractsManage or wildcard.
func hasContractsManage(c *gin.Context) bool {
	u, ok := currentUser(c)
	if !ok {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermUsersContractsManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// List godoc
// @Summary      List contracts for a user
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path   string  true   "user uuid"
// @Param        page        query  int     false  "page (default 1)"
// @Param        page_size   query  int     false  "page size (default 10, max 50)"
// @Param        signed_from query  string  false  "signed date from (RFC3339)"
// @Param        signed_to   query  string  false  "signed date to (RFC3339)"
// @Param        expiry_from query  string  false  "expiry date from (RFC3339)"
// @Param        expiry_to   query  string  false  "expiry date to (RFC3339)"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.UserContractRead]]
// @Router       /users/{id}/contracts [get]
func (h *UserContractHandler) List(c *gin.Context) {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var q dto.UserContractListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.List(c.Request.Context(), userID, q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.UserContractRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary      Get a single contract
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path  string  true  "user uuid"
// @Param        contractID  path  string  true  "contract uuid"
// @Success      200  {object}  dto.Response[dto.UserContractRead]
// @Failure      404  {object}  dto.Response[any]
// @Router       /users/{id}/contracts/{contractID} [get]
func (h *UserContractHandler) Get(c *gin.Context) {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, aerr := h.svc.Get(c.Request.Context(), userID, contractID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create a contract for a user
// @Tags         contracts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true  "user uuid"
// @Param        body  body  dto.UserContractCreate  true  "create payload"
// @Success      201  {object}  dto.Response[dto.UserContractRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Router       /users/{id}/contracts [post]
func (h *UserContractHandler) Create(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.UserContractCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Create(c.Request.Context(), userID, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.UserContractRead]{Success: true, Message: "Contract has been created", Data: out})
}

// Update godoc
// @Summary      Partial-patch a contract
// @Tags         contracts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id          path  string                  true  "user uuid"
// @Param        contractID  path  string                  true  "contract uuid"
// @Param        body        body  dto.UserContractUpdate  true  "patch payload"
// @Success      200  {object}  dto.Response[dto.UserContractRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /users/{id}/contracts/{contractID} [patch]
func (h *UserContractHandler) Update(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.UserContractUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Update(c.Request.Context(), userID, contractID, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractRead]{Success: true, Message: "Contract has been updated", Data: out})
}

// Delete godoc
// @Summary      Soft-delete a contract
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path  string  true  "user uuid"
// @Param        contractID  path  string  true  "contract uuid"
// @Success      200  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /users/{id}/contracts/{contractID} [delete]
func (h *UserContractHandler) Delete(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if aerr := h.svc.Delete(c.Request.Context(), userID, contractID); aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Contract has been deleted"})
}

// UploadAttachment godoc
// @Summary      Upload or replace a contract attachment
// @Tags         contracts
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path      string  true  "user uuid"
// @Param        contractID  path      string  true  "contract uuid"
// @Param        file        formData  file    true  "attachment (PDF/PNG/JPG/DOCX, max 5MB)"
// @Success      200  {object}  dto.Response[dto.UserContractAttachmentResponse]
// @Failure      400  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Router       /users/{id}/contracts/{contractID}/attachment [post]
func (h *UserContractHandler) UploadAttachment(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	fileHeader, ferr := c.FormFile("file")
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("file is required"))
		return
	}
	f, ferr := fileHeader.Open()
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("cannot open uploaded file"))
		return
	}
	defer f.Close()
	content, ferr := io.ReadAll(f)
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("cannot read uploaded file"))
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	out, aerr := h.svc.UploadAttachment(c.Request.Context(), userID, contractID, content, ext)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractAttachmentResponse]{Success: true, Data: out})
}
