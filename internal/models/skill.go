package models

// Skill is a named, optionally-iconed capability in the catalog. Assigned
// to employees through the employee_skills join (see EmployeeSkill).
// Catalog rows are unique by case-insensitive name among live (NotDeleted)
// rows; uniqueness is enforced by the Postgres partial unique index
// uq_skills_name_lower_live and additionally pre-checked at the service
// layer to return ErrConflict with a friendly message.
type Skill struct {
	BaseModel
	Name        string  `gorm:"type:text;not null"            json:"name"`
	Description string  `gorm:"type:text;not null;default:''" json:"description"`
	IconURL     *string `gorm:"type:text"                     json:"icon_url,omitempty"`
}

func (Skill) TableName() string { return "skills" }
