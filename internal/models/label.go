package models

// Label is an announcement label — a minimal name-only entity used to tag
// announcements (introduced in Phase 7). Per the Python source, the only
// supported operations are list and get-or-create by case-insensitive
// name (no update, no delete). Uniqueness is enforced both by the
// partial unique index uq_labels_name_lower_live and by a service-layer
// pre-check.
type Label struct {
	BaseModel
	Name string `gorm:"type:text;not null" json:"name"`
}

func (Label) TableName() string { return "labels" }
