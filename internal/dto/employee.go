package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---- Shared refs ----

type RefRead struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type RoleRead struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
}

// ---- Employee Read ----

type EmployeeRead struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"user_id"`
	Email                    string     `json:"email"`
	FullName                 string     `json:"full_name"`
	Phone                    *string    `json:"phone,omitempty"`
	PersonalEmail            *string    `json:"personal_email,omitempty"`
	Gender                   *string    `json:"gender,omitempty"`
	DOB                      *time.Time `json:"dob,omitempty"`
	Nationality              *string    `json:"nationality,omitempty"`
	IDNumber                 *string    `json:"id_number,omitempty"`
	IDIssueDate              *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage             *string    `json:"id_front_image,omitempty"`
	IDBackImage              *string    `json:"id_back_image,omitempty"`
	PermanentAddress         *string    `json:"permanent_address,omitempty"`
	CurrentAddress           *string    `json:"current_address,omitempty"`
	Education                *string    `json:"education,omitempty"`
	MaritalStatus            *string    `json:"marital_status,omitempty"`
	EmergencyContactName     *string    `json:"emergency_contact_name,omitempty"`
	EmergencyContactRelation *string    `json:"emergency_contact_relation,omitempty"`
	EmergencyContactPhone    *string    `json:"emergency_contact_phone,omitempty"`
	AvatarURL                *string    `json:"avatar_url,omitempty"`

	// Work
	Department *RefRead   `json:"department,omitempty"`
	Position   *RefRead   `json:"position,omitempty"`
	Manager    *RefRead   `json:"manager,omitempty"`
	JoinDate   *time.Time `json:"join_date,omitempty"`

	// Contract / salary / bank (admin view only — service-level helper may strip for non-admin)
	ContractType     *string    `json:"contract_type,omitempty"`
	ContractSignDate *time.Time `json:"contract_sign_date,omitempty"`
	ContractEndDate  *time.Time `json:"contract_end_date,omitempty"`
	ContractRenewal  *bool      `json:"contract_renewal,omitempty"`
	BasicSalary      *float64   `json:"basic_salary,omitempty"`
	InsuranceSalary  *float64   `json:"insurance_salary,omitempty"`
	BankAccount      *string    `json:"bank_account,omitempty"`
	BankName         *string    `json:"bank_name,omitempty"`
	BankHolderName   *string    `json:"bank_holder_name,omitempty"`
	PaymentMethod    *string    `json:"payment_method,omitempty"`

	// Auth side
	IsActive  bool       `json:"is_active"`
	Roles     []RoleRead `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Optional embed for nested dependent reads on /employees/me
	Dependents []DependentRead `json:"dependents,omitempty"`
}

// NOTE: EmployeeSummary is already declared in internal/dto/auth.go (created in
// Phase 2 Tasks 1-7) and is reused here by UserMeRead. It is intentionally NOT
// redeclared in this file to avoid a duplicate-type conflict; the auth.go shape
// (ID / FullName / AvatarURL / DepartmentID / PositionID / ManagerID) satisfies
// both the auth login flow and GET /users/me.

// ---- Employee Create (admin) — accepts both user creds and HR fields ----

type EmployeeCreate struct {
	// User credentials (created in tx)
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	IsActive *bool  `json:"is_active,omitempty"`

	// HR personal info
	FullName                 string     `json:"full_name" binding:"required,min=1,max=200"`
	Phone                    *string    `json:"phone,omitempty"`
	PersonalEmail            *string    `json:"personal_email,omitempty" binding:"omitempty,email"`
	Gender                   *string    `json:"gender,omitempty"          binding:"omitempty,oneof=male female other"`
	DOB                      *time.Time `json:"dob,omitempty"`
	Nationality              *string    `json:"nationality,omitempty"`
	IDNumber                 *string    `json:"id_number,omitempty"`
	IDIssueDate              *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage             *string    `json:"id_front_image,omitempty"`
	IDBackImage              *string    `json:"id_back_image,omitempty"`
	PermanentAddress         *string    `json:"permanent_address,omitempty"`
	CurrentAddress           *string    `json:"current_address,omitempty"`
	Education                *string    `json:"education,omitempty"`
	MaritalStatus            *string    `json:"marital_status,omitempty"  binding:"omitempty,oneof=single married divorced widowed"`
	EmergencyContactName     *string    `json:"emergency_contact_name,omitempty"`
	EmergencyContactRelation *string    `json:"emergency_contact_relation,omitempty"`
	EmergencyContactPhone    *string    `json:"emergency_contact_phone,omitempty"`

	// Work
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	JoinDate     *time.Time `json:"join_date,omitempty"`

	// Contract / salary / bank
	ContractType     *string    `json:"contract_type,omitempty"`
	ContractSignDate *time.Time `json:"contract_sign_date,omitempty"`
	ContractEndDate  *time.Time `json:"contract_end_date,omitempty"`
	ContractRenewal  *bool      `json:"contract_renewal,omitempty"`
	BasicSalary      *float64   `json:"basic_salary,omitempty"      binding:"omitempty,gte=0"`
	InsuranceSalary  *float64   `json:"insurance_salary,omitempty"  binding:"omitempty,gte=0"`
	BankAccount      *string    `json:"bank_account,omitempty"`
	BankName         *string    `json:"bank_name,omitempty"`
	BankHolderName   *string    `json:"bank_holder_name,omitempty"`
	PaymentMethod    *string    `json:"payment_method,omitempty"`

	// Roles assigned at creation
	RoleIDs []uuid.UUID `json:"role_ids,omitempty"`
}

// ---- Employee Update (admin — anything allowed) ----

type EmployeeUpdate struct {
	FullName                 *string    `json:"full_name,omitempty"`
	Phone                    *string    `json:"phone,omitempty"`
	PersonalEmail            *string    `json:"personal_email,omitempty" binding:"omitempty,email"`
	Gender                   *string    `json:"gender,omitempty"          binding:"omitempty,oneof=male female other"`
	DOB                      *time.Time `json:"dob,omitempty"`
	Nationality              *string    `json:"nationality,omitempty"`
	IDNumber                 *string    `json:"id_number,omitempty"`
	IDIssueDate              *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage             *string    `json:"id_front_image,omitempty"`
	IDBackImage              *string    `json:"id_back_image,omitempty"`
	PermanentAddress         *string    `json:"permanent_address,omitempty"`
	CurrentAddress           *string    `json:"current_address,omitempty"`
	Education                *string    `json:"education,omitempty"`
	MaritalStatus            *string    `json:"marital_status,omitempty"  binding:"omitempty,oneof=single married divorced widowed"`
	EmergencyContactName     *string    `json:"emergency_contact_name,omitempty"`
	EmergencyContactRelation *string    `json:"emergency_contact_relation,omitempty"`
	EmergencyContactPhone    *string    `json:"emergency_contact_phone,omitempty"`

	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	ClearDept    *bool      `json:"clear_department,omitempty"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	ClearPos     *bool      `json:"clear_position,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	ClearManager *bool      `json:"clear_manager,omitempty"`
	JoinDate     *time.Time `json:"join_date,omitempty"`

	ContractType     *string    `json:"contract_type,omitempty"`
	ContractSignDate *time.Time `json:"contract_sign_date,omitempty"`
	ContractEndDate  *time.Time `json:"contract_end_date,omitempty"`
	ContractRenewal  *bool      `json:"contract_renewal,omitempty"`
	BasicSalary      *float64   `json:"basic_salary,omitempty"      binding:"omitempty,gte=0"`
	InsuranceSalary  *float64   `json:"insurance_salary,omitempty"  binding:"omitempty,gte=0"`
	BankAccount      *string    `json:"bank_account,omitempty"`
	BankName         *string    `json:"bank_name,omitempty"`
	BankHolderName   *string    `json:"bank_holder_name,omitempty"`
	PaymentMethod    *string    `json:"payment_method,omitempty"`

	IsActive *bool `json:"is_active,omitempty"` // toggles user.is_active
}

// ---- Employee Self Update (RESTRICTED — server-side whitelist) ----
//
// IMPORTANT: This DTO MUST NOT contain any of: department_id, position_id, manager_id,
// basic_salary, insurance_salary, contract_*, role_ids, is_active.
// The service enforces the whitelist with a manual field-by-field copy.
type EmployeeSelfUpdate struct {
	Phone                    *string `json:"phone,omitempty"`
	PersonalEmail            *string `json:"personal_email,omitempty" binding:"omitempty,email"`
	PermanentAddress         *string `json:"permanent_address,omitempty"`
	CurrentAddress           *string `json:"current_address,omitempty"`
	MaritalStatus            *string `json:"marital_status,omitempty"  binding:"omitempty,oneof=single married divorced widowed"`
	EmergencyContactName     *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactRelation *string `json:"emergency_contact_relation,omitempty"`
	EmergencyContactPhone    *string `json:"emergency_contact_phone,omitempty"`
}

// ---- Leave quota update (admin) ----

type LeaveQuotaUpdateRequest struct {
	AnnualLeaveQuota float64 `json:"annual_leave_quota" binding:"gte=0,lte=365"`
	SickLeaveQuota   float64 `json:"sick_leave_quota"   binding:"gte=0,lte=365"`
}

// ---- List query ----

type EmployeeListQuery struct {
	Page         int        `form:"page,default=1"       binding:"gte=1"`
	PageSize     int        `form:"page_size,default=20" binding:"gte=1,lte=100"`
	Search       string     `form:"search"`
	DepartmentID *uuid.UUID `form:"department_id"`
	PositionID   *uuid.UUID `form:"position_id"`
	ManagerID    *uuid.UUID `form:"manager_id"`
	RoleID       *uuid.UUID `form:"role_id"`
	IsActive     *bool      `form:"is_active"`
}
