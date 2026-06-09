package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// AnnouncementHandler owns /api/v1/announcements (web) + /api/v1/mobile/
// announcements (mobile). Two-layer access control: route-level
// RequirePerms upstream; hasAnnounceManageAll walks the JWT-preloaded
// user.Roles for the asAdmin bool — same shape as
// hasLeaveManageAll/hasAttendanceManageAll.
type AnnouncementHandler struct {
	svc *services.AnnouncementService
}

func NewAnnouncementHandler(svc *services.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc}
}

// hasAnnounceManageAll walks the JWT-preloaded roles for "*" or
// PermAnnounceManage.
func hasAnnounceManageAll(c *gin.Context) bool {
	u, ok := currentUser(c)
	if !ok {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermAnnounceManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// List godoc
// @Summary      List announcements (web)
// @Description  Non-admins see only rows that satisfy the visibility predicate (published + audience match OR author). Admins see all.
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        page          query  int     false  "page"
// @Param        page_size     query  int     false  "page size"
// @Param        search        query  string  false  "title/description search"
// @Param        status        query  string  false  "draft|scheduled|published|archived"
// @Param        label_id      query  string  false  "label UUID"
// @Param        pinned        query  bool    false  "filter by pinned"
// @Param        scope         query  string  false  "all|mine|targeted-at-me"
// @Param        department_id query  string  false  "department UUID (admin)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/announcements [get]
func (h *AnnouncementHandler) List(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var q dto.AnnouncementListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.List(c.Request.Context(), u.ID, hasAnnounceManageAll(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AnnouncementRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary      Get an announcement by ID
// @Description  Returns 403 when the caller cannot see this row per the visibility predicate.
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "announcement uuid"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/announcements/{id} [get]
func (h *AnnouncementHandler) Get(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id, u.ID, hasAnnounceManageAll(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AnnouncementRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create an announcement
// @Description  Saved as draft by default. Pass `status: "published"` in the
// @Description  body to publish immediately (broadcasts via SSE). The Python-
// @Description  parity shortcut `send_now: true` also works — when set and
// @Description  `status` is not explicitly provided, the row is created
// @Description  already published. Explicit `status` always wins.
// @Description
// @Description  `target_audience` accepts `all`, `department`, or `custom`.
// @Description  `department` requires at least one `department_ids` entry;
// @Description  `custom` requires at least one `recipient_ids` entry
// @Description  (employee_ids).
// @Tags         announcements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.AnnouncementCreate  true  "create payload"
// @Success      201   {object}  map[string]interface{}
// @Router       /api/v1/announcements [post]
func (h *AnnouncementHandler) Create(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.AnnouncementCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), u.ID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	msg := "Announcement saved as draft"
	if out != nil && out.PublishedAt != nil {
		msg = "Announcement published"
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.AnnouncementRead]{Success: true, Message: msg, Data: out})
}

// Update godoc
// @Summary      Update an announcement (owner or admin)
// @Tags         announcements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true  "announcement uuid"
// @Param        body  body  dto.AnnouncementUpdate  true  "patch payload"
// @Success      200   {object}  map[string]interface{}
// @Failure      409   {object}  map[string]interface{}  "announcement is published or archived"
// @Router       /api/v1/announcements/{id} [patch]
func (h *AnnouncementHandler) Update(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.AnnouncementUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, u.ID, hasAnnounceManageAll(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AnnouncementRead]{Success: true, Message: "Updated", Data: out})
}

// Delete godoc
// @Summary      Soft-delete an announcement (owner or admin)
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "announcement uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/announcements/{id} [delete]
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, u.ID, hasAnnounceManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Deleted"})
}

// Publish godoc
// @Summary      Publish an announcement
// @Description  Stamps published_at + broadcasts via SSE. No-op if already published.
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "announcement uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/announcements/{id}/publish [post]
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Publish(c.Request.Context(), id, u.ID, hasAnnounceManageAll(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AnnouncementRead]{Success: true, Message: "Published", Data: out})
}

// MarkViewed godoc
// @Summary      Mark an announcement as viewed (idempotent)
// @Description  Records a per-user read marker. Second call is a no-op (preserves the first view time).
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "announcement uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/announcements/{id}/view [post]
func (h *AnnouncementHandler) MarkViewed(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.MarkViewed(c.Request.Context(), id, u.ID, hasAnnounceManageAll(c)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Marked as viewed"})
}

// MobileBrief godoc
// @Summary      Top-5 announcements (mobile home widget)
// @Description  Returns the latest 5 published announcements visible to the
// @Description  current user, as `MobileAnnouncementBrief` items (description
// @Description  omitted; fetch full detail via MobileGet). Unpaginated.
// @Description  Mirrors Python's `GET /mobile/announcements/` contract.
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/mobile/announcements [get]
func (h *AnnouncementHandler) MobileBrief(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	items, err := h.svc.MobileBrief(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.MobileAnnouncementBrief]{Success: true, Data: items})
}

// MobileList godoc
// @Summary      List announcements (mobile, paginated)
// @Description  Always visibility-filtered to published + audience match.
// @Description  Description field omitted from each item; fetch detail via MobileGet.
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "page"
// @Param        page_size  query  int  false  "page size"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/mobile/announcements/list [get]
func (h *AnnouncementHandler) MobileList(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var q dto.MobileAnnouncementListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.MobileList(c.Request.Context(), u.ID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.MobileAnnouncementBrief]]{Success: true, Data: out})
}

// MobileGet godoc
// @Summary      Get an announcement (mobile)
// @Tags         announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "announcement uuid"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/mobile/announcements/{id} [get]
func (h *AnnouncementHandler) MobileGet(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.MobileGet(c.Request.Context(), id, u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.AnnouncementRead]{Success: true, Data: out})
}
