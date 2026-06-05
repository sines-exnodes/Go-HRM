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
	"github.com/exnodes/hrm-api/internal/services"
)

type SkillHandler struct {
	svc *services.SkillService
}

func NewSkillHandler(svc *services.SkillService) *SkillHandler {
	return &SkillHandler{svc: svc}
}

// maxSkillIconBytes mirrors skill_service.skillIconMaxBytes (2MB, per BA
// DR-008-003-02). Kept duplicated here so the handler can reject oversized
// uploads before the service is even called.
const maxSkillIconBytes = 2 * 1024 * 1024

// readSkillIcon extracts the optional `icon` multipart file from the
// request. Returns (nil, nil) when the form field is absent — that is
// the supported "no icon" case for both Create and Update.
func readSkillIcon(c *gin.Context) (*services.SkillIconUpload, error) {
	fh, err := c.FormFile("icon")
	if err != nil {
		// Distinguish "field not present" (OK) from a real upload error.
		if err == http.ErrMissingFile {
			return nil, nil
		}
		// Gin's multipart parser returns its own errors for malformed bodies;
		// treat them as "no icon" so a JSON-style update body still works.
		return nil, nil
	}
	if fh == nil {
		return nil, nil
	}
	if fh.Size > maxSkillIconBytes {
		return nil, apperrors.ErrBadRequest("Icon must not exceed 2MB")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, apperrors.ErrBadRequest("cannot read icon")
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, maxSkillIconBytes+1))
	if err != nil {
		return nil, apperrors.ErrBadRequest("cannot read icon")
	}
	if len(content) > maxSkillIconBytes {
		return nil, apperrors.ErrBadRequest("Icon must not exceed 2MB")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	return &services.SkillIconUpload{
		Content:     content,
		ContentType: fh.Header.Get("Content-Type"),
		Ext:         ext,
	}, nil
}

// List godoc
// @Summary      List skills
// @Description  Paginated, sorted by name ASC. Optional ILIKE search by name.
// @Tags         skills
// @Security     BearerAuth
// @Produce      json
// @Param        page       query    int     false  "Page number"  default(1)
// @Param        page_size  query    int     false  "Page size"    default(20)
// @Param        search     query    string  false  "Substring match on name (ILIKE)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/skills [get]
func (h *SkillHandler) List(c *gin.Context) {
	var q dto.SkillListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.SkillRead]]{Success: true, Data: data})
}

// Create godoc
// @Summary      Create skill (admin)
// @Tags         skills
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        name        formData  string  true   "Skill name (1..100, regex [a-zA-Z0-9 &.+#/()-])"
// @Param        description formData  string  false  "Skill description (<=500)"
// @Param        icon        formData  file    false  "Optional icon (image, <=2MB)"
// @Success      201  {object}  map[string]interface{}
// @Router       /api/v1/skills [post]
func (h *SkillHandler) Create(c *gin.Context) {
	var in dto.SkillCreate
	// ShouldBind handles either multipart or x-www-form-urlencoded.
	if err := c.ShouldBind(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	icon, err := readSkillIcon(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in, icon)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.SkillRead]{
		Success: true,
		Message: "Skill created",
		Data:    out,
	})
}

// Get godoc
// @Summary      Get skill by ID
// @Tags         skills
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Skill UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/skills/{id} [get]
func (h *SkillHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.Response[*dto.SkillRead]{Success: true, Data: out})
}

// Update godoc
// @Summary      Update skill (admin)
// @Description  Multipart PATCH. Omitted text fields are left alone; an
// @Description  uploaded `icon` replaces the previous icon and the old
// @Description  object is best-effort deleted from storage.
// @Tags         skills
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path      string  true   "Skill UUID"
// @Param        name        formData  string  false  "New name"
// @Param        description formData  string  false  "New description"
// @Param        icon        formData  file    false  "New icon (image, <=2MB)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/skills/{id} [patch]
func (h *SkillHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	// Manual binding: PATCH fields are optional pointers. ShouldBind would
	// happily accept an absent `name` form key but populate it as "" — which
	// the service would then reject as "name cannot be blank". We instead
	// inspect PostForm presence explicitly.
	var in dto.SkillUpdate
	if v, ok := c.GetPostForm("name"); ok {
		in.Name = &v
	}
	if v, ok := c.GetPostForm("description"); ok {
		in.Description = &v
	}
	icon, err := readSkillIcon(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in, icon)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.SkillRead]{
		Success: true,
		Message: "Skill updated",
		Data:    out,
	})
}

// Delete godoc
// @Summary      Delete skill (admin)
// @Description  Soft-deletes a skill. Returns 409 with employee_count
// @Description  in details if any live employee is still assigned.
// @Tags         skills
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Skill UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/skills/{id} [delete]
func (h *SkillHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Skill deleted"})
}

// ---- Employee ↔ Skill assignment (nested under /employees) ----

// ListForEmployee godoc
// @Summary      List skills assigned to an employee
// @Tags         skills
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Employee UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/employees/{id}/skills [get]
func (h *SkillHandler) ListForEmployee(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.ListForEmployee(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.SkillRead]{Success: true, Data: out})
}

// ReplaceForEmployee godoc
// @Summary      Replace an employee's skill set (PUT semantics)
// @Description  The whole skill_ids set is replaced atomically — passing
// @Description  an empty list unassigns everything.
// @Tags         skills
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string                       true  "Employee UUID"
// @Param        body  body      dto.EmployeeSkillsReplace    true  "Desired skill_ids"
// @Success      200   {object}  map[string]interface{}
// @Router       /api/v1/employees/{id}/skills [put]
func (h *SkillHandler) ReplaceForEmployee(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	var in dto.EmployeeSkillsReplace
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.ReplaceForEmployee(c.Request.Context(), id, in.SkillIDs)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.SkillRead]{
		Success: true,
		Message: "Skills updated",
		Data:    out,
	})
}
