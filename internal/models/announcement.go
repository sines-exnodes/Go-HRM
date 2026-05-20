package models

import (
	"time"

	"github.com/google/uuid"
)

// AnnouncementStatus is the lifecycle status of an announcement.
// Mirrors Python's StrEnum domain.
type AnnouncementStatus string

const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusScheduled AnnouncementStatus = "scheduled"
	AnnouncementStatusPublished AnnouncementStatus = "published"
	AnnouncementStatusArchived  AnnouncementStatus = "archived"
)

// AnnouncementTargetAudience controls how recipients are resolved.
// Phase 7 supports 'all' (everyone) and 'department' (target_departments
// join). 'custom' (per-user targeting) is deferred until BA confirms — no
// backing table exists yet (REVISION NOTES item #4).
type AnnouncementTargetAudience string

const (
	AnnouncementAudienceAll        AnnouncementTargetAudience = "all"
	AnnouncementAudienceDepartment AnnouncementTargetAudience = "department"
)

// Announcement is the publishable entity. author_id references the HR
// profile (employees(id)) per the Go schema split — mirrors Phase 5
// leave_requests.created_by. The seeded super admin already has an
// employee row so this never blocks admin authoring.
type Announcement struct {
	BaseModel
	Title          string                     `gorm:"type:text;not null"                        json:"title"`
	Body           string                     `gorm:"type:text;not null"                        json:"body"`
	Summary        *string                    `gorm:"type:text"                                  json:"summary,omitempty"`
	AuthorID       uuid.UUID                  `gorm:"type:uuid;not null;index"                   json:"author_id"`
	Status         AnnouncementStatus         `gorm:"type:text;not null;default:'draft';index"   json:"status"`
	ScheduledAt    *time.Time                 `                                                  json:"scheduled_at,omitempty"`
	PublishedAt    *time.Time                 `                                                  json:"published_at,omitempty"`
	TargetAudience AnnouncementTargetAudience `gorm:"type:text;not null;default:'all'"           json:"target_audience"`
	Pinned         bool                       `gorm:"not null;default:false"                     json:"pinned"`
	CoverImageURL  *string                    `gorm:"type:text"                                  json:"cover_image_url,omitempty"`

	// Relations — preloaded on demand. NOTE: AnnouncementLabels is the
	// explicit join model with audit cols (REVISION NOTES #10), NOT a
	// gorm:"many2many" tag. Labels is a derived projection populated by
	// the repo's preload for FE convenience.
	Author             *Employee                      `gorm:"foreignKey:AuthorID;references:ID"          json:"author,omitempty"`
	AnnouncementLabels []AnnouncementLabel            `gorm:"foreignKey:AnnouncementID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	TargetDepartments  []AnnouncementTargetDepartment `gorm:"foreignKey:AnnouncementID;references:ID;constraint:OnDelete:CASCADE" json:"target_departments,omitempty"`
	Attachments        []AnnouncementAttachment       `gorm:"foreignKey:AnnouncementID;references:ID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`
}

// TableName pins to "announcements" — GORM would pluralize to
// "announcements" anyway but the explicit table name protects against
// future model renames.
func (Announcement) TableName() string { return "announcements" }

// AnnouncementLabel is the explicit join row between an announcement and
// a label in the catalog (Phase 4). Mirrors EmployeeSkill — required
// because the join table carries the four audit columns. Using the
// gorm:"many2many" tag would skip them.
type AnnouncementLabel struct {
	AnnouncementID uuid.UUID  `gorm:"type:uuid;primaryKey;not null"             json:"announcement_id"`
	LabelID        uuid.UUID  `gorm:"type:uuid;primaryKey;not null;index"       json:"label_id"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"                    json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"                    json:"updated_at"`
	IsDeleted      bool       `gorm:"not null;default:false"                    json:"-"`
	DeletedAt      *time.Time `                                                 json:"-"`

	// Relations.
	Label *Label `gorm:"foreignKey:LabelID;references:ID" json:"label,omitempty"`
}

func (AnnouncementLabel) TableName() string { return "announcement_labels" }

// AnnouncementTargetDepartment is the join row for target_audience='department'.
type AnnouncementTargetDepartment struct {
	AnnouncementID uuid.UUID  `gorm:"type:uuid;primaryKey;not null"          json:"announcement_id"`
	DepartmentID   uuid.UUID  `gorm:"type:uuid;primaryKey;not null;index"    json:"department_id"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"                  json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"                  json:"updated_at"`
	IsDeleted      bool       `gorm:"not null;default:false"                  json:"-"`
	DeletedAt      *time.Time `                                               json:"-"`

	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
}

func (AnnouncementTargetDepartment) TableName() string { return "announcement_target_departments" }

// AnnouncementAttachment is one uploaded file linked to an announcement.
// Multiple per announcement supported (HR commonly attaches a PDF + a
// banner image together).
type AnnouncementAttachment struct {
	BaseModel
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index"  json:"announcement_id"`
	URL            string    `gorm:"type:text;not null"        json:"url"`
	Filename       string    `gorm:"type:text;not null"        json:"filename"`
	ContentType    string    `gorm:"type:text;not null"        json:"content_type"`
	SizeBytes      int64     `gorm:"not null;default:0"        json:"size_bytes"`
}

func (AnnouncementAttachment) TableName() string { return "announcement_attachments" }

// AnnouncementView is the per-user read marker (auth-level). user_id
// targets users(id) (NOT employees(id)) — the marker is keyed on the
// logged-in session, matching the Python source's read tracker.
type AnnouncementView struct {
	AnnouncementID uuid.UUID  `gorm:"type:uuid;primaryKey;not null"          json:"announcement_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey;not null;index"    json:"user_id"`
	ViewedAt       time.Time  `gorm:"not null;default:now()"                  json:"viewed_at"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"                  json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"                  json:"updated_at"`
	IsDeleted      bool       `gorm:"not null;default:false"                  json:"-"`
	DeletedAt      *time.Time `                                               json:"-"`
}

func (AnnouncementView) TableName() string { return "announcement_views" }
