package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---- Skill ----

// SkillCreate is the parsed input for POST /api/v1/skills. The handler
// fills these fields from multipart form values (`name`, `description`)
// and passes the optional icon bytes separately. Validation (regex,
// length, trimming) happens in SkillService — DTO-side binding tags
// only guard against an obviously empty `name` field.
type SkillCreate struct {
	Name        string `form:"name"        json:"name"                  binding:"required"`
	Description string `form:"description" json:"description,omitempty"`
}

// SkillUpdate is the parsed input for PATCH /api/v1/skills/:id. All
// fields are optional — pointer types let the service distinguish "not
// provided" (leave alone) from "set to empty string" (description only).
type SkillUpdate struct {
	Name        *string `form:"name"        json:"name,omitempty"`
	Description *string `form:"description" json:"description,omitempty"`
}

// SkillRead is the wire shape returned by every skill endpoint.
type SkillRead struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     *string   `json:"icon_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SkillListQuery binds the querystring for GET /api/v1/skills. Sort is
// fixed to name ASC (matches the Python source); only page/page_size/search
// are accepted.
type SkillListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string `form:"search"`
}

// SkillDeleteBlocked is the conflict-response payload returned when a
// skill cannot be deleted because employees are still assigned to it.
// EmployeeCount mirrors the Python source's structured conflict body so
// the FE can render a friendly "still assigned to N employees" message.
type SkillDeleteBlocked struct {
	SkillID       uuid.UUID `json:"skill_id"`
	SkillName     string    `json:"skill_name"`
	EmployeeCount int64     `json:"employee_count"`
}

// ---- Employee ↔ Skill assignment ----

// EmployeeSkillsReplace is the body for PUT /api/v1/employees/:id/skills.
// PUT-replace semantics matches the Python source's User.skill_ids array
// (the whole set is replaced atomically). An empty list clears all
// assignments. Duplicate IDs in the request are silently de-duplicated.
type EmployeeSkillsReplace struct {
	SkillIDs []uuid.UUID `json:"skill_ids" binding:"required"`
}
