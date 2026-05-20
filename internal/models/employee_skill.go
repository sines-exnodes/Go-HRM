package models

import "github.com/google/uuid"

// EmployeeSkill is the join row between an employee and a skill in the
// catalog. The Python source modelled this as User.skill_ids (UUID array);
// the Go schema split (users ⟂ employees ⟂ dependents) puts the assignment
// on the HR profile.
//
// FK semantics (see migration 000006_create_skills):
//   - employee_id REFERENCES employees(id) ON DELETE CASCADE — deleting an
//     employee removes their skill links.
//   - skill_id    REFERENCES skills(id)    (no cascade) — skill deletion is
//     blocked at the service layer with HTTP 409 when any live link exists
//     (mirrors the Phase 3 department/position delete guard).
//
// Pair uniqueness among live rows is enforced by the partial unique index
// uq_employee_skills_pair_live(employee_id, skill_id) WHERE is_deleted=false.
type EmployeeSkill struct {
	BaseModel
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	SkillID    uuid.UUID `gorm:"type:uuid;not null;index" json:"skill_id"`

	// Relations — preloaded on demand, omitted from JSON when nil.
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID" json:"employee,omitempty"`
	Skill    *Skill    `gorm:"foreignKey:SkillID;references:ID"    json:"skill,omitempty"`
}

func (EmployeeSkill) TableName() string { return "employee_skills" }
