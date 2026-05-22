package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// OrganizationSettingsHandler owns /api/v1/organization-settings — two
// sub-resources: /attendance (admin-only read + write) and
// /company-profile (open read, admin-only write).
type OrganizationSettingsHandler struct {
	svc *services.OrganizationSettingsService
}

func NewOrganizationSettingsHandler(svc *services.OrganizationSettingsService) *OrganizationSettingsHandler {
	return &OrganizationSettingsHandler{svc: svc}
}

// GetAttendance godoc
// @Summary      Get attendance thresholds
// @Description  Returns the four attendance threshold fields (late/checkout, hour/minute).
// @Tags         organization-settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/organization-settings/attendance [get]
func (h *OrganizationSettingsHandler) GetAttendance(c *gin.Context) {
	out, err := h.svc.GetAttendance(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AttendanceSettingsRead]{Success: true, Data: out})
}

// UpdateAttendance godoc
// @Summary      Update attendance thresholds (partial)
// @Tags         organization-settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AttendanceSettingsUpdate  true  "patch payload"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/organization-settings/attendance [patch]
func (h *OrganizationSettingsHandler) UpdateAttendance(c *gin.Context) {
	var req dto.AttendanceSettingsUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.UpdateAttendance(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AttendanceSettingsRead]{Success: true, Message: "Attendance settings updated", Data: out})
}

// GetCompanyProfile godoc
// @Summary      Get company profile (address + lat/lng)
// @Description  Open to any authenticated user — the FE renders the map preview on shared screens.
// @Tags         organization-settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/organization-settings/company-profile [get]
func (h *OrganizationSettingsHandler) GetCompanyProfile(c *gin.Context) {
	out, err := h.svc.GetCompanyProfile(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.CompanyProfileRead]{Success: true, Data: out})
}

// UpdateCompanyProfile godoc
// @Summary      Update company profile (admin only)
// @Description  Stamps company_address_updated_at + updated_by whenever any of the three address fields are supplied.
// @Tags         organization-settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CompanyProfileUpdate  true  "patch payload"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/organization-settings/company-profile [patch]
func (h *OrganizationSettingsHandler) UpdateCompanyProfile(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.CompanyProfileUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.UpdateCompanyProfile(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.CompanyProfileRead]{Success: true, Message: "Company profile updated", Data: out})
}
