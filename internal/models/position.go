package models

// Position is a job role in the global catalog. Position names are unique
// across the whole table (case-insensitive) among non-deleted rows. The
// link from people to positions lives on employees.position_id; positions
// themselves are no longer associated with a department (see migration
// 000014).
type Position struct {
	BaseModel
	Name        string `gorm:"type:text;not null"            json:"name"`
	Description string `gorm:"type:text;not null;default:''" json:"description"`
}

func (Position) TableName() string { return "positions" }
