package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// AttendanceHandler owns the /api/v1/attendance HTTP surface (10 endpoints
// + matrix). Two-layer access control: the route-level RequirePerms gate
// runs upstream; the handler precomputes asAdmin from the JWT-preloaded
// user.Roles and passes it down so the service can scope reads to own
// employee_id when the caller lacks PermAttendanceManage.
type AttendanceHandler struct {
	svc *services.AttendanceService
}

func NewAttendanceHandler(svc *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

// hasAttendanceManageAll walks the JWT-preloaded user.Roles for the
// wildcard "*" permission or PermAttendanceManage. Same shape as
// hasLeaveManageAll in leave_handler.go.
func hasAttendanceManageAll(c *gin.Context) bool {
	u, okC := currentUser(c)
	if !okC {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermAttendanceManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// CheckIn godoc
// @Summary      Record a check-in for the authenticated user
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AttendanceCheckInReq  true  "check-in payload"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/attendance/check-in [post]
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var req dto.AttendanceCheckInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.CheckIn(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Checked in", Data: out})
}

// CheckOut godoc
// @Summary      Record a check-out for the authenticated user
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AttendanceCheckOutReq  false  "check-out payload"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/attendance/check-out [post]
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var req dto.AttendanceCheckOutReq
	// Empty body is fine — defaults to "now".
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.CheckOut(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Checked out", Data: out})
}

// Today godoc
// @Summary      Get today's attendance status for the authenticated user
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/today [get]
func (h *AttendanceHandler) Today(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	out, err := h.svc.Today(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.TodayStatusRead]{Success: true, Data: out})
}

// Me godoc
// @Summary      List my own attendance rows
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int     false  "page"
// @Param        page_size  query  int     false  "page size"
// @Param        start_date query  string  false  "YYYY-MM-DD"
// @Param        end_date   query  string  false  "YYYY-MM-DD"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/me [get]
func (h *AttendanceHandler) Me(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.AttendanceListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	// /me is always scoped to self regardless of perms.
	out, err := h.svc.List(c.Request.Context(), u.ID, false, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AttendanceRead]]{Success: true, Data: out})
}

// List godoc
// @Summary      Flat list of attendance rows (Go convenience; not the BA matrix)
// @Description  Managers (with attendance:manage_data) see all rows; non-managers see only their own.
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        page          query  int     false  "page"
// @Param        page_size     query  int     false  "page size"
// @Param        employee_id   query  string  false  "filter by employee"
// @Param        department_id query  string  false  "filter by department"
// @Param        start_date    query  string  false  "YYYY-MM-DD"
// @Param        end_date      query  string  false  "YYYY-MM-DD"
// @Param        status        query  string  false  "on_time|late"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/records [get]
func (h *AttendanceHandler) List(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.AttendanceListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.List(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AttendanceRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary      Get an attendance row by ID (owner or admin)
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "attendance uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/{id} [get]
func (h *AttendanceHandler) Get(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id, u.ID, hasAttendanceManageAll(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Data: out})
}

// AdminCreate godoc
// @Summary      Admin manual create of an attendance row
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.AttendanceAdminCreateReq  true  "payload"
// @Success      201  {object}  map[string]interface{}
// @Router       /api/v1/attendance [post]
func (h *AttendanceHandler) AdminCreate(c *gin.Context) {
	var req dto.AttendanceAdminCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.AdminCreate(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.AttendanceRead]{Success: true, Message: "Created", Data: out})
}

// AdminUpdate godoc
// @Summary      Admin update of an attendance row
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                          true  "attendance uuid"
// @Param        body  body  dto.AttendanceAdminUpdateReq    true  "payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/{id} [patch]
func (h *AttendanceHandler) AdminUpdate(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.AttendanceAdminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Updated", Data: out})
}

// AdminDelete godoc
// @Summary      Admin soft-delete of an attendance row
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "attendance uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/{id} [delete]
func (h *AttendanceHandler) AdminDelete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Deleted"})
}

// Export godoc
// @Summary      Export the monthly attendance matrix to Excel (all visible employees)
// @Tags         attendance
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        month  query  int  false  "1-12"
// @Param        year   query  int  false  "YYYY"
// @Success      200  {file}  binary
// @Router       /api/v1/attendance/export [get]
func (h *AttendanceHandler) Export(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ExportMatrix(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}
	writeXlsx(c, data, "attendance")
}

// ExportEmployee godoc
// @Summary      Export a single employee's monthly attendance to Excel
// @Tags         attendance
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        employee_id  path   string  true  "employee uuid"
// @Param        month        query  int     false  "1-12"
// @Param        year         query  int     false  "YYYY"
// @Success      200  {file}  binary
// @Router       /api/v1/attendance/export/{employee_id} [get]
func (h *AttendanceHandler) ExportEmployee(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	empID, err := parseIDParam(c, "employee_id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ExportMatrix(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q, &empID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	writeXlsx(c, data, "attendance_"+empID.String())
}

// writeXlsx streams an xlsx byte slice as a download.
func writeXlsx(c *gin.Context, data []byte, basename string) {
	c.Header("Content-Disposition", `attachment; filename="`+basename+`.xlsx"`)
	c.Data(http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data)
}

// Matrix godoc
// @Summary      Monthly attendance matrix (managers: all employees; others: own row)
// @Description  Managers see all employees; non-managers see only their own row.
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        month         query  int     false  "1-12"
// @Param        year          query  int     false  "YYYY"
// @Param        page          query  int     false  "page"
// @Param        page_size     query  int     false  "page size"
// @Param        search        query  string  false  "name filter (managers only)"
// @Param        department_id query  string  false  "department UUID (managers only)"
// @Param        status        query  string  false  "CSV: on_time,late,absent,weekend,no_data"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance [get]
func (h *AttendanceHandler) Matrix(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Matrix(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceMatrixRead]{Success: true, Data: out})
}
