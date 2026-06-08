package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// LeaveHandler owns the /api/v1/leave-requests HTTP surface (10 endpoints
// + history/dashboard self-service). Permission gating happens upstream
// via middleware.RequirePerms; this handler precomputes a "manage all"
// bool from the JWT-preloaded user.Roles and passes it down so the
// service can enforce ownership without an extra dep on AuthService
// (mirrors DependentHandler.hasManageAll).
type LeaveHandler struct {
	svc *services.LeaveService
}

func NewLeaveHandler(svc *services.LeaveService) *LeaveHandler {
	return &LeaveHandler{svc: svc}
}

// maxLeaveAttachmentBytes mirrors leave_service.leaveAttachmentMaxBytes
// (5 MB). Duplicated so the handler can reject oversized uploads before
// the service is called.
const maxLeaveAttachmentBytes = 5 * 1024 * 1024

// hasLeaveManageAll walks the JWT-preloaded user.Roles for the wildcard
// "*" permission or PermLeaveManage. Same shape as hasManageAll in
// dependent_handler.go.
func hasLeaveManageAll(c *gin.Context) bool {
	u, okC := currentUser(c)
	if !okC {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermLeaveManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// resolveApproveScope scans JWT-loaded roles for an approve permission.
// Priority: PermAll / PermLeaveApproveAll / legacy PermLeaveApprove → ApproveScopeAll;
//
//	PermLeaveApproveTeam → ApproveScopeTeam (keep scanning for stronger).
func resolveApproveScope(c *gin.Context) (services.ApproveScope, bool) {
	u, okC := currentUser(c)
	if !okC {
		return 0, false
	}
	var found services.ApproveScope
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			switch permissions.Permission(p) {
			case permissions.PermAll, permissions.PermLeaveApproveAll, permissions.PermLeaveApprove:
				return services.ApproveScopeAll, true
			case permissions.PermLeaveApproveTeam:
				found = services.ApproveScopeTeam
			}
		}
	}
	if found != 0 {
		return found, true
	}
	return 0, false
}

// readLeaveAttachment extracts the optional `attachment` multipart file.
// Returns (nil, nil) when the form field is absent — supported for both
// Create and Update. Mirrors readSkillIcon's shape.
func readLeaveAttachment(c *gin.Context) (*services.AttachmentUpload, error) {
	fh, err := c.FormFile("attachment")
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil
		}
		// Malformed/multipart-less bodies — treat as "no attachment" so a
		// JSON-only update body still works.
		return nil, nil
	}
	if fh == nil {
		return nil, nil
	}
	if fh.Size > maxLeaveAttachmentBytes {
		return nil, apperrors.ErrBadRequest("Attachment must not exceed 5MB")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, apperrors.ErrBadRequest("cannot read attachment")
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, maxLeaveAttachmentBytes+1))
	if err != nil {
		return nil, apperrors.ErrBadRequest("cannot read attachment")
	}
	if len(content) > maxLeaveAttachmentBytes {
		return nil, apperrors.ErrBadRequest("Attachment must not exceed 5MB")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	return &services.AttachmentUpload{
		Content:     content,
		ContentType: fh.Header.Get("Content-Type"),
		Ext:         ext,
	}, nil
}

// bindLeaveJSON reads a JSON payload either from the raw request body
// (Content-Type: application/json) or from a multipart `data` form field
// (Content-Type: multipart/form-data; matches the Python contract for
// attach-able requests). Caller passes a pointer to the target struct.
func bindLeaveJSON(c *gin.Context, dst any) error {
	// Multipart path — pull from the `data` form field.
	if ct := c.GetHeader("Content-Type"); strings.HasPrefix(strings.ToLower(ct), "multipart/") {
		raw := c.PostForm("data")
		if strings.TrimSpace(raw) == "" {
			return apperrors.ErrBadRequest("missing `data` form field")
		}
		if err := json.Unmarshal([]byte(raw), dst); err != nil {
			return apperrors.ErrBadRequest("invalid JSON in `data` field: " + err.Error())
		}
		return nil
	}
	// JSON path.
	if err := c.ShouldBindJSON(dst); err != nil {
		return apperrors.ErrBadRequest(err.Error())
	}
	return nil
}

// List godoc
// @Summary      List leave requests (admin / manager)
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        page          query int      false "page (default 1)"
// @Param        page_size     query int      false "page size (default 10, max 100)"
// @Param        search        query string   false "free-text search on employee name/email"
// @Param        status        query []string false "filter by status (repeat param: ?status=pending&status=approved)"  collectionFormat(multi)
// @Param        department_id query string   false "department uuid"
// @Param        position_id   query string   false "position uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests [get]
func (h *LeaveHandler) List(c *gin.Context) {
	var q dto.LeaveListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]{Success: true, Data: data})
}

// GetBalance godoc
// @Summary      Get leave balance for an employee
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id path  string true  "employee uuid"
// @Param        year        query int    false "calendar year (defaults to current UTC year)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/balance/{employee_id} [get]
func (h *LeaveHandler) GetBalance(c *gin.Context) {
	employeeID, err := parseIDParam(c, "employee_id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	year := 0
	if y := strings.TrimSpace(c.Query("year")); y != "" {
		parsed, perr := strconv.Atoi(y)
		if perr != nil || parsed < 1970 || parsed > 9999 {
			_ = c.Error(apperrors.ErrBadRequest("invalid year"))
			return
		}
		year = parsed
	}
	out, err := h.svc.GetBalance(c.Request.Context(), employeeID, year)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveBalanceSummary]{Success: true, Data: out})
}

// GetMyDashboard godoc
// @Summary      Get the current user's leave dashboard (balance + upcoming + history)
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/dashboard/me [get]
func (h *LeaveHandler) GetMyDashboard(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	out, err := h.svc.GetMyDashboard(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveDashboardRead]{Success: true, Data: out})
}

// ListMyHistory godoc
// @Summary      List the current user's leave history (paginated)
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        page       query int      false "page (default 1)"
// @Param        page_size  query int      false "page size (default 10, max 100)"
// @Param        status     query []string false "filter by status"  collectionFormat(multi)
// @Param        start_date query string   false "lower bound on from_date (YYYY-MM-DD)"
// @Param        end_date   query string   false "upper bound on to_date (YYYY-MM-DD)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/history/me [get]
func (h *LeaveHandler) ListMyHistory(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.LeaveHistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ListMyHistory(c.Request.Context(), u.ID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]{Success: true, Data: data})
}

// Create godoc
// @Summary      Create a leave request (optionally with an attachment)
// @Description  Accepts either application/json (body is the request DTO) or multipart/form-data with a `data` JSON field plus an optional `attachment` file.
// @Tags         leave-requests
// @Security     BearerAuth
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        data       formData string  false "JSON-encoded LeaveRequestCreate (multipart only)"
// @Param        attachment formData file    false "Optional attachment (PDF/PNG/JPG/GIF/WEBP, <=5MB)"
// @Success      201 {object} map[string]interface{}
// @Router       /api/v1/leave-requests [post]
func (h *LeaveHandler) Create(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.LeaveRequestCreate
	if err := bindLeaveJSON(c, &in); err != nil {
		_ = c.Error(err)
		return
	}
	attachment, err := readLeaveAttachment(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Create(c.Request.Context(), u.ID, hasLeaveManageAll(c), in, attachment)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.LeaveRequestWriteResult]{
		Success: true,
		Message: "Leave request created",
		Data:    out,
	})
}

// Get godoc
// @Summary      Get a leave request by ID (owner or admin)
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "leave request uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id} [get]
func (h *LeaveHandler) Get(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id, u.ID, hasLeaveManageAll(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
}

// Update godoc
// @Summary      Update a leave request (owner-of-pending or admin)
// @Description  Multipart or JSON. Pointer fields in the body distinguish "not provided" from explicit values. If currently approved, an admin patch reverts status to pending (Python contract).
// @Tags         leave-requests
// @Security     BearerAuth
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        id         path     string true  "leave request uuid"
// @Param        data       formData string false "JSON-encoded LeaveRequestUpdate (multipart only)"
// @Param        attachment formData file   false "Replacement attachment (PDF/PNG/JPG/GIF/WEBP, <=10MB)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id} [patch]
func (h *LeaveHandler) Update(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var in dto.LeaveRequestUpdate
	if err := bindLeaveJSON(c, &in); err != nil {
		_ = c.Error(err)
		return
	}
	attachment, err := readLeaveAttachment(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, u.ID, hasLeaveManageAll(c), in, attachment)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestWriteResult]{Success: true, Data: out})
}

// Approve godoc
// @Summary      Approve a pending leave request
// @Description  Requires approve_team (own subordinate chain only) or approve_all (any employee).
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "leave request UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/approve [post]
func (h *LeaveHandler) Approve(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	scope, ok := resolveApproveScope(c)
	if !ok {
		_ = c.Error(apperrors.ErrForbidden("Insufficient approve permission"))
		return
	}
	out, err := h.svc.Approve(c.Request.Context(), id, u.ID, scope)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
}

// Reject godoc
// @Summary      Reject a pending leave request
// @Description  Same permission semantics as Approve — requires approve_team or approve_all.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "leave request UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/reject [post]
func (h *LeaveHandler) Reject(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	scope, ok := resolveApproveScope(c)
	if !ok {
		_ = c.Error(apperrors.ErrForbidden("Insufficient approve permission"))
		return
	}
	out, err := h.svc.Reject(c.Request.Context(), id, u.ID, scope)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
}

// Cancel godoc
// @Summary      Cancel a leave request (owner or admin)
// @Description  Cancel transitions pending or approved -> cancelled. Owner can cancel their own; admin can cancel anyone's.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "leave request uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/cancel [post]
func (h *LeaveHandler) Cancel(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, wasApproved, err := h.svc.Cancel(c.Request.Context(), id, u.ID, hasLeaveManageAll(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	type cancelResult struct {
		*dto.LeaveRequestRead
		WasApproved bool `json:"was_approved"`
	}
	c.JSON(http.StatusOK, dto.Response[cancelResult]{
		Success: true,
		Data:    cancelResult{LeaveRequestRead: out, WasApproved: wasApproved},
	})
}

// Delete godoc
// @Summary      Soft-delete a leave request (owner-of-pending or admin)
// @Description  POST /leave-requests/{id}/delete (not DELETE) — matches the Python source. Admin may delete any status; non-admin owner only `pending`.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "leave request uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/delete [post]
func (h *LeaveHandler) Delete(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, u.ID, hasLeaveManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Leave request deleted"})
}

// Export godoc
// @Summary      Export leave requests as Excel
// @Description  Returns an .xlsx download matching the same filters as the list endpoint. Requires leave_requests:read.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        status        query  []string  false  "filter by status"
// @Param        department_id query  string    false  "filter by department UUID"
// @Param        position_id   query  string    false  "filter by position UUID"
// @Param        search        query  string    false  "search employee name"
// @Success      200  {string}  binary
// @Router       /api/v1/leave-requests/export [get]
func (h *LeaveHandler) Export(c *gin.Context) {
	var q dto.LeaveListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ExportLeave(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	writeLeaveXlsx(c, data, "leave-requests")
}

func writeLeaveXlsx(c *gin.Context, data []byte, basename string) {
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, basename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
