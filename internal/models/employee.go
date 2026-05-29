package models

import (
	"time"

	"github.com/google/uuid"
)

// Employee maps to the employees table. 1-1 with User via UserID.
type Employee struct {
	BaseModel
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

	// Personal info
	FullName         string     `gorm:"type:text;not null" json:"full_name"`
	Phone            *string    `gorm:"type:text" json:"phone,omitempty"`
	PersonalEmail    *string    `gorm:"type:citext" json:"personal_email,omitempty"`
	Gender           *string    `gorm:"type:text" json:"gender,omitempty"`
	PermanentAddress *string    `gorm:"type:text" json:"permanent_address,omitempty"`
	CurrentAddress   *string    `gorm:"type:text" json:"current_address,omitempty"`
	DOB              *time.Time `gorm:"type:date" json:"dob,omitempty"`
	Nationality      *string    `gorm:"type:text" json:"nationality,omitempty"`
	IDNumber         *string    `gorm:"type:text" json:"id_number,omitempty"`
	IDIssueDate      *time.Time `gorm:"type:date" json:"id_issue_date,omitempty"`
	IDFrontImage     *string    `gorm:"type:text" json:"id_front_image,omitempty"`
	IDBackImage      *string    `gorm:"type:text" json:"id_back_image,omitempty"`
	AvatarURL        *string    `gorm:"type:text" json:"avatar_url,omitempty"`
	Education        *string    `gorm:"type:text" json:"education,omitempty"`
	MaritalStatus    *string    `gorm:"type:text" json:"marital_status,omitempty"`
	ExperienceYear   *int       `gorm:"type:int" json:"experience_year,omitempty"`
	CVURL            *string    `gorm:"type:text" json:"cv_url,omitempty"`

	// Emergency contacts are a 1-N list (migration 000017); the former single
	// emergency_contact_{name,relation,phone} columns were dropped. See the
	// EmergencyContacts relation below.

	// Work info — department_id / position_id have NO FK constraint until
	// Phase 3 introduces departments/positions tables.
	DepartmentID     *uuid.UUID `gorm:"type:uuid" json:"department_id,omitempty"`
	PositionID       *uuid.UUID `gorm:"type:uuid" json:"position_id,omitempty"`
	ManagerID        *uuid.UUID `gorm:"type:uuid" json:"manager_id,omitempty"`
	JoinDate         *time.Time `gorm:"type:date" json:"join_date,omitempty"`
	ContractType     string     `gorm:"type:text;not null;default:'official'" json:"contract_type"`
	ContractSignDate *time.Time `gorm:"type:date" json:"contract_sign_date,omitempty"`
	ContractEndDate  *time.Time `gorm:"type:date" json:"contract_end_date,omitempty"`
	ContractRenewal  int        `gorm:"not null;default:1" json:"contract_renewal"`

	// Salary & insurance
	BasicSalary     float64 `gorm:"type:numeric(18,2);not null;default:0" json:"basic_salary"`
	InsuranceSalary float64 `gorm:"type:numeric(18,2);not null;default:0" json:"insurance_salary"`

	// Banking
	BankAccount    *string `gorm:"type:text" json:"bank_account,omitempty"`
	BankName       *string `gorm:"type:text" json:"bank_name,omitempty"`
	BankHolderName *string `gorm:"type:text" json:"bank_holder_name,omitempty"`
	PaymentMethod  string  `gorm:"type:text;not null;default:'bank_transfer'" json:"payment_method"`

	// Relations
	User              *User                      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Manager           *Employee                  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Department        *Department                `gorm:"foreignKey:DepartmentID;references:ID" json:"-"`
	Position          *Position                  `gorm:"foreignKey:PositionID;references:ID" json:"-"`
	Subordinates      []Employee                 `gorm:"foreignKey:ManagerID" json:"subordinates,omitempty"`
	Dependents        []Dependent                `gorm:"foreignKey:EmployeeID" json:"dependents,omitempty"`
	EmergencyContacts []EmployeeEmergencyContact `gorm:"foreignKey:EmployeeID" json:"emergency_contacts,omitempty"`
	EmployeeSkills    []EmployeeSkill            `gorm:"foreignKey:EmployeeID" json:"employee_skills,omitempty"`
	LeaveQuota        *EmployeeLeaveQuota        `gorm:"foreignKey:EmployeeID" json:"leave_quota,omitempty"`
}

func (Employee) TableName() string { return "employees" }
