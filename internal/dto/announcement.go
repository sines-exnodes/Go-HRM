package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/models"
)

// ---- Write inputs ----

// AnnouncementCreate is the body for POST /announcements. When status is
// omitted the row is created as a draft (unless send_now=true — see below).
// label_ids must reference live rows in the labels catalog. department_ids
// is meaningful only when target_audience == "department". recipient_ids
// is meaningful only when target_audience == "custom" (per-employee target).
//
// send_now is the Python-parity shortcut: when true and status is not
// explicitly set, the row is created already in the "published" state and
// the SSE event is broadcast immediately. When status IS explicitly set,
// send_now is ignored (explicit status wins).
type AnnouncementCreate struct {
	Title          string                             `json:"title"           binding:"required,min=1"`
	Description    string                             `json:"description"     binding:"required,min=1"`
	Summary        *string                            `json:"summary,omitempty"`
	Status         *models.AnnouncementStatus         `json:"status,omitempty"          binding:"omitempty,oneof=draft scheduled published archived"`
	ScheduledAt    *time.Time                         `json:"scheduled_at,omitempty"`
	TargetAudience *models.AnnouncementTargetAudience `json:"target_audience,omitempty" binding:"omitempty,oneof=all department custom"`
	Pinned         *bool                              `json:"pinned,omitempty"`
	CoverImageURL  *string                            `json:"cover_image_url,omitempty"`
	LabelIDs       []uuid.UUID                        `json:"label_ids,omitempty"`
	DepartmentIDs  []uuid.UUID                        `json:"department_ids,omitempty"`
	RecipientIDs   []uuid.UUID                        `json:"recipient_ids,omitempty"`
	SendNow        bool                               `json:"send_now,omitempty"`
}

// AnnouncementUpdate is the PATCH body. Pointer types preserve "not
// provided" semantics; nil slice means "leave label/department/recipient
// links unchanged"; empty slice (length 0) means "clear all links".
type AnnouncementUpdate struct {
	Title          *string                            `json:"title,omitempty"           binding:"omitempty,min=1"`
	Description    *string                            `json:"description,omitempty"     binding:"omitempty,min=1"`
	Summary        *string                            `json:"summary,omitempty"`
	Status         *models.AnnouncementStatus         `json:"status,omitempty"          binding:"omitempty,oneof=draft scheduled published archived"`
	ScheduledAt    *time.Time                         `json:"scheduled_at,omitempty"`
	TargetAudience *models.AnnouncementTargetAudience `json:"target_audience,omitempty" binding:"omitempty,oneof=all department custom"`
	Pinned         *bool                              `json:"pinned,omitempty"`
	CoverImageURL  *string                            `json:"cover_image_url,omitempty"`
	LabelIDs       *[]uuid.UUID                       `json:"label_ids,omitempty"`
	DepartmentIDs  *[]uuid.UUID                       `json:"department_ids,omitempty"`
	RecipientIDs   *[]uuid.UUID                       `json:"recipient_ids,omitempty"`
}

// ---- Read outputs ----

// AnnouncementAuthorBrief is the embedded {id, full_name, avatar_url}
// projection. Author maps to employees(id) per the schema split.
type AnnouncementAuthorBrief struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// AnnouncementLabelBrief is the minimal {id, name} projection. Labels in
// this codebase have no color field (see internal/models/label.go).
type AnnouncementLabelBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AnnouncementDepartmentBrief is the minimal {id, name} projection.
type AnnouncementDepartmentBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AnnouncementRecipientBrief is the per-employee projection for the
// target_audience='custom' join. Mirrors AnnouncementAuthorBrief.
type AnnouncementRecipientBrief struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// AnnouncementAttachmentRead is the per-attachment projection.
type AnnouncementAttachmentRead struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

// AnnouncementRead is the canonical wire shape returned by every web
// announcement endpoint. Labels/TargetDepartments/Attachments are
// inflated by the service when their FKs are live.
type AnnouncementRead struct {
	ID                uuid.UUID                         `json:"id"`
	Title             string                            `json:"title"`
	Description       string                            `json:"description"`
	Summary           *string                           `json:"summary,omitempty"`
	Status            models.AnnouncementStatus         `json:"status"`
	ScheduledAt       *time.Time                        `json:"scheduled_at,omitempty"`
	PublishedAt       *time.Time                        `json:"published_at,omitempty"`
	TargetAudience    models.AnnouncementTargetAudience `json:"target_audience"`
	Pinned            bool                              `json:"pinned"`
	CoverImageURL     *string                           `json:"cover_image_url,omitempty"`
	Author            *AnnouncementAuthorBrief          `json:"author,omitempty"`
	Labels            []AnnouncementLabelBrief          `json:"labels"`
	TargetDepartments []AnnouncementDepartmentBrief     `json:"target_departments"`
	TargetRecipients  []AnnouncementRecipientBrief      `json:"target_recipients"`
	Attachments       []AnnouncementAttachmentRead      `json:"attachments"`
	HasViewed         bool                              `json:"has_viewed"`
	CreatedAt         time.Time                         `json:"created_at"`
	UpdatedAt         time.Time                         `json:"updated_at"`
}

// ---- List queries ----

// AnnouncementListQuery binds the querystring for GET /announcements.
// scope: "all" (default — visibility-filtered), "mine" (rows authored
// by the current user, manager surface), "targeted-at-me" (rows that
// would surface to the user via target_audience matching, even drafts
// included if author = me).
type AnnouncementListQuery struct {
	Page         int    `form:"page,default=1"        binding:"min=1"`
	PageSize     int    `form:"page_size,default=20"  binding:"min=1,max=100"`
	Search       string `form:"search"`
	Status       string `form:"status"`
	LabelID      string `form:"label_id"`
	Pinned       *bool  `form:"pinned"`
	Scope        string `form:"scope"`
	DepartmentID string `form:"department_id"`
}

// ---- Mobile shapes ----

// MobileAnnouncementBrief is the home-screen widget projection: minimal
// fields needed to render the top-N cards. Excludes Body (rendered only
// on detail) to keep the payload small.
type MobileAnnouncementBrief struct {
	ID            uuid.UUID                 `json:"id"`
	Title         string                    `json:"title"`
	Summary       *string                   `json:"summary,omitempty"`
	CoverImageURL *string                   `json:"cover_image_url,omitempty"`
	Status        models.AnnouncementStatus `json:"status"`
	Pinned        bool                      `json:"pinned"`
	PublishedAt   *time.Time                `json:"published_at,omitempty"`
	Labels        []AnnouncementLabelBrief  `json:"labels"`
	HasViewed     bool                      `json:"has_viewed"`
}

// MobileAnnouncementListQuery binds the querystring for GET
// /mobile/announcements. No status filter — mobile clients only see
// published rows. No scope filter — always visibility-filtered.
type MobileAnnouncementListQuery struct {
	Page     int `form:"page,default=1"        binding:"min=1"`
	PageSize int `form:"page_size,default=20"  binding:"min=1,max=100"`
}

// ---- SSE event payload ----

// SSEAnnouncementPublishedEvent is the Data shape for the
// "announcement_published" event broadcast on /sse/announcements.
// Includes only the fields the FE needs to render a toast or refresh
// its list — Body is omitted to keep frames small.
//
// RecipientIDs carries employee_ids when target_audience='custom'; FE may
// short-circuit the refetch when the current user is not in the set.
// Server-side visibility is still enforced on the subsequent GET.
type SSEAnnouncementPublishedEvent struct {
	ID             uuid.UUID                         `json:"id"`
	Title          string                            `json:"title"`
	Summary        *string                           `json:"summary,omitempty"`
	TargetAudience models.AnnouncementTargetAudience `json:"target_audience"`
	DepartmentIDs  []uuid.UUID                       `json:"department_ids,omitempty"`
	RecipientIDs   []uuid.UUID                       `json:"recipient_ids,omitempty"`
	Pinned         bool                              `json:"pinned"`
	PublishedAt    time.Time                         `json:"published_at"`
}
