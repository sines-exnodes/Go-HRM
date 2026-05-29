package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

type EmployeeHandler struct {
	svc *services.EmployeeService
}

func NewEmployeeHandler(svc *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// ---- Shared response helpers ----

func ok[T any](c *gin.Context, status int, data T, message string) {
	c.JSON(status, gin.H{"success": true, "message": message, "data": data})
}
func okEmpty(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": message})
}

// ---- Shared current-user helper ----

func currentUser(c *gin.Context) (*models.User, bool) {
	v, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		_ = c.Error(apperrors.ErrUnauthorized("unauthenticated"))
		return nil, false
	}
	u, isUser := v.(*models.User)
	if !isUser {
		_ = c.Error(apperrors.ErrUnauthorized("unauthenticated"))
		return nil, false
	}
	return u, true
}

func parseIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, apperrors.ErrBadRequest("invalid " + key)
	}
	return id, nil
}

// employeeFieldPerms derives the caller's salary/banking field-level perms
// (employees parity #6) from their JWT-loaded roles. The wildcard "*" grants
// all four. Does not error when the user is absent (returns the zero value);
// the route's RequirePerms already guards authentication.
func employeeFieldPerms(c *gin.Context) services.EmployeeFieldPerms {
	v, exists := c.Get(middleware.ContextKeyUser)
	if !exists {
		return services.EmployeeFieldPerms{}
	}
	u, ok := v.(*models.User)
	if !ok {
		return services.EmployeeFieldPerms{}
	}
	var p services.EmployeeFieldPerms
	for _, r := range u.Roles {
		for _, perm := range []string(r.Permissions) {
			switch perm {
			case string(permissions.PermAll):
				return services.AllEmployeeFieldPerms
			case string(permissions.PermUsersSalaryView):
				p.SalaryView = true
			case string(permissions.PermUsersSalaryManage):
				p.SalaryManage = true
			case string(permissions.PermUsersBankingView):
				p.BankingView = true
			case string(permissions.PermUsersBankingManage):
				p.BankingManage = true
			}
		}
	}
	return p
}

// ---- Avatar multipart helper ----

func readAvatar(c *gin.Context) ([]byte, string, string, error) {
	fh, err := c.FormFile("avatar")
	if err != nil {
		return nil, "", "", apperrors.ErrBadRequest("avatar file is required")
	}
	if fh.Size > 5*1024*1024 {
		return nil, "", "", apperrors.ErrBadRequest("Avatar must not exceed 5MB")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, "", "", apperrors.ErrBadRequest("cannot read avatar")
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, 5*1024*1024+1))
	if err != nil {
		return nil, "", "", apperrors.ErrBadRequest("cannot read avatar")
	}
	if len(content) > 5*1024*1024 {
		return nil, "", "", apperrors.ErrBadRequest("Avatar must not exceed 5MB")
	}
	ct := fh.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return nil, "", "", apperrors.ErrBadRequest("Avatar must be an image")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true}
	if !allowed[ext] {
		return nil, "", "", apperrors.ErrBadRequest("Only PNG, JPG, JPEG, and WEBP files are accepted")
	}
	return content, ct, ext, nil
}

// ---- Admin endpoints ----

// List godoc
// @Summary      List employees (admin)
// @Tags         employees
// @Security     BearerAuth
// @Produce      json
// @Param        page          query int    false "page (default 1)"
// @Param        page_size     query int    false "page size (default 20, max 100)"
// @Param        search        query string false "free text (full_name/phone/personal_email/user.email)"
// @Param        department_id query string false "department uuid"
// @Param        position_id   query string false "position uuid"
// @Param        manager_id    query string false "manager uuid"
// @Param        role_id       query string false "role uuid"
// @Param        is_active     query bool   false "filter by user.is_active"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees [get]
func (h *EmployeeHandler) List(c *gin.Context) {
	var q dto.EmployeeListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	perms := employeeFieldPerms(c)
	for i := range items {
		services.ApplyEmployeeFieldVisibility(&items[i], perms, false)
	}
	totalPages := 0
	if q.PageSize > 0 && total > 0 {
		totalPages = int((total + int64(q.PageSize) - 1) / int64(q.PageSize))
	}
	ok(c, http.StatusOK, gin.H{
		"items": items, "total": total,
		"page": q.Page, "page_size": q.PageSize, "total_pages": totalPages,
	}, "")
}

// Create godoc
// @Summary      Create employee (admin) — creates user + employee in tx
// @Tags         employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.EmployeeCreate true "user creds + HR fields"
// @Success      201 {object} map[string]interface{}
// @Router       /api/v1/employees [post]
func (h *EmployeeHandler) Create(c *gin.Context) {
	var in dto.EmployeeCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	perms := employeeFieldPerms(c)
	// Salary/banking write-gating (employees parity #6) — 403 if the payload
	// sets fields the caller may not manage.
	if err := services.GuardSalaryWrite(in.BasicSalary != nil || in.InsuranceSalary != nil, perms); err != nil {
		_ = c.Error(err)
		return
	}
	if err := services.GuardBankingWrite(in.BankAccount != nil || in.BankName != nil || in.BankHolderName != nil || in.PaymentMethod != nil, perms); err != nil {
		_ = c.Error(err)
		return
	}
	view, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, perms, true) // write echo: unmasked
	ok(c, http.StatusCreated, view, "Employee created")
}

// Get godoc
// @Summary      Get employee by id (admin)
// @Tags         employees
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "employee uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id} [get]
func (h *EmployeeHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	view, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), false) // read: masked
	ok(c, http.StatusOK, view, "")
}

// Update godoc
// @Summary      Update employee (admin)
// @Tags         employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "employee uuid"
// @Param        body body dto.EmployeeUpdate true "fields"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id} [patch]
func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.EmployeeUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	perms := employeeFieldPerms(c)
	// Salary/banking write-gating (employees parity #6).
	if err := services.GuardSalaryWrite(in.BasicSalary != nil || in.InsuranceSalary != nil, perms); err != nil {
		_ = c.Error(err)
		return
	}
	if err := services.GuardBankingWrite(in.BankAccount != nil || in.BankName != nil || in.BankHolderName != nil || in.PaymentMethod != nil, perms); err != nil {
		_ = c.Error(err)
		return
	}
	view, err := h.svc.Update(c.Request.Context(), id, in, u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, perms, true) // write echo: unmasked
	ok(c, http.StatusOK, view, "Employee updated")
}

// Delete godoc
// @Summary      Soft-delete employee + deactivate user (admin)
// @Tags         employees
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "employee uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id} [delete]
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	u, okC := currentUser(c)
	if !okC {
		return
	}
	if err := h.svc.SoftDelete(c.Request.Context(), id, u.ID); err != nil {
		_ = c.Error(err)
		return
	}
	okEmpty(c, "Employee deleted")
}

// UpdateAvatarAdmin godoc
// @Summary      Update an employee's avatar (admin)
// @Tags         employees
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path     string true "employee uuid"
// @Param        avatar formData file   true "image (PNG/JPEG/WEBP, max 5MB)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/avatar [patch]
func (h *EmployeeHandler) UpdateAvatarAdmin(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	content, ct, ext, err := readAvatar(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	view, err := h.svc.UpdateAvatarAdmin(c.Request.Context(), id, content, ct, ext)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), true) // write echo
	ok(c, http.StatusOK, view, "Avatar updated")
}

// UpdateLeaveQuota godoc
// @Summary      Update employee leave quota (admin)
// @Tags         employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "employee uuid"
// @Param        body body dto.LeaveQuotaUpdateRequest true "quotas"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/{id}/leave-quota [patch]
func (h *EmployeeHandler) UpdateLeaveQuota(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var in dto.LeaveQuotaUpdateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.UpdateLeaveQuota(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), true) // write echo
	ok(c, http.StatusOK, view, "Leave quota updated")
}

// ---- Self-service endpoints ----

// GetMe godoc
// @Summary      Get current employee HR profile
// @Tags         employees
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/me [get]
func (h *EmployeeHandler) GetMe(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	view, err := h.svc.GetByUserID(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), false) // read: masked
	ok(c, http.StatusOK, view, "")
}

// UpdateMe godoc
// @Summary      Update current employee HR profile (restricted whitelist)
// @Tags         employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.EmployeeSelfUpdate true "allowed fields only"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/me [patch]
func (h *EmployeeHandler) UpdateMe(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var in dto.EmployeeSelfUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	view, err := h.svc.SelfUpdate(c.Request.Context(), u.ID, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), true) // own write echo
	ok(c, http.StatusOK, view, "Profile updated")
}

// UpdateMyAvatar godoc
// @Summary      Replace current employee avatar
// @Tags         employees
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar formData file true "Image (PNG/JPEG/WEBP, max 5MB)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/employees/me/avatar [patch]
func (h *EmployeeHandler) UpdateMyAvatar(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	content, ct, ext, err := readAvatar(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	view, err := h.svc.UpdateAvatarSelf(c.Request.Context(), u.ID, content, ct, ext)
	if err != nil {
		_ = c.Error(err)
		return
	}
	services.ApplyEmployeeFieldVisibility(view, employeeFieldPerms(c), true) // own write echo
	ok(c, http.StatusOK, view, "Avatar updated")
}
