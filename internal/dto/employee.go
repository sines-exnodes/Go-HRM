package dto

import (
	"time"

	"github.com/google/uuid"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// ---- Shared refs ----

type RefRead struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// RoleRef is the brief role embed used inside EmployeeRead and UserMeRead/
// UserAdminRead. For the full role API shape see dto.RoleRead in role.go.
type RoleRef struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
}

// ---- Emergency contacts (1-N list — employees parity #4) ----

// EmergencyContactRead is one emergency contact in a read response.
type EmergencyContactRead struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Relationship string    `json:"relationship"`
	PhoneNumber  string    `json:"phone_number"`
}

// EmergencyContactInput is one emergency contact in a create/update payload.
// relationship/phone are optional (mirrors Python's defaulted sub-document);
// full_name is required so we never persist a nameless contact.
type EmergencyContactInput struct {
	FullName     string `json:"full_name"    binding:"required,min=1,max=200"`
	Relationship string `json:"relationship" binding:"omitempty,max=100"`
	PhoneNumber  string `json:"phone_number" binding:"omitempty,max=50"`
}

// ---- Line manager (employees parity #10) ----

// ManagerBrief is the resolved line-manager embedded on an employee read:
// name + email + contact + position/department names + activity flag (mirrors
// Python's LineManagerRead). Position/Department/Phone/AvatarURL are nil when unset.
type ManagerBrief struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Phone      *string   `json:"phone,omitempty"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	Position   *string   `json:"position,omitempty"`
	Department *string   `json:"department,omitempty"`
	IsActive   bool      `json:"is_active"`
}

// ManagerCandidateRead is a row in the line-manager picker.
type ManagerCandidateRead struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Position   *string   `json:"position,omitempty"`
	Department *string   `json:"department,omitempty"`
	IsActive   bool      `json:"is_active"`
}

// DirectReportRead is a row in the "direct reports" card (includes inactive).
type DirectReportRead struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	Position   *string   `json:"position,omitempty"`
	Department *string   `json:"department,omitempty"`
	IsActive   bool      `json:"is_active"`
}

// ManagerCandidateQuery binds the querystring for the picker endpoint.
// ForEmployeeID is a string (parsed in the handler) — gin cannot bind a
// uuid.UUID directly from a query param.
type ManagerCandidateQuery struct {
	ForEmployeeID string `form:"for_employee_id"`
	Search        string `form:"search"`
	Limit         int    `form:"limit,default=50" binding:"gte=1,lte=200"`
}

// ---- Employee Read ----

type EmployeeRead struct {
	ID                      uuid.UUID  `json:"id"`
	UserID                  uuid.UUID  `json:"user_id"`
	Email                   string     `json:"email"`
	FirstName               string     `json:"first_name"`
	LastName                string     `json:"last_name"`
	Phone                   *string    `json:"phone,omitempty"`
	PersonalEmail           *string    `json:"personal_email,omitempty"`
	Gender                  *string    `json:"gender,omitempty"`
	DOB                     *time.Time `json:"dob,omitempty"`
	Nationality             *string    `json:"nationality,omitempty"`
	IDNumber                *string    `json:"id_number,omitempty"`
	IDIssueDate             *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage            *string    `json:"id_front_image,omitempty"`
	IDBackImage             *string    `json:"id_back_image,omitempty"`
	PermanentAddress        *string    `json:"permanent_address,omitempty"`
	CurrentAddress          *string    `json:"current_address,omitempty"`
	Education               *string    `json:"education,omitempty"`
	MaritalStatus           *string    `json:"marital_status,omitempty"`
	ExperienceYear          *int       `json:"experience_year,omitempty"`
	CVURL                   *string    `json:"cv_url,omitempty"`
	AvatarURL               *string    `json:"avatar_url,omitempty"`
	SocialInsuranceNumber   *string    `json:"social_insurance_number,omitempty"`
	TaxIdentificationNumber *string    `json:"tax_identification_number,omitempty"`

	// Emergency contacts — 1-N list (employees parity #4). Always present
	// (empty array when none) so the FE can render the section unconditionally.
	EmergencyContacts []EmergencyContactRead `json:"emergency_contacts"`

	// Work
	Department *RefRead      `json:"department,omitempty"`
	Position   *RefRead      `json:"position,omitempty"`
	Manager    *ManagerBrief `json:"manager,omitempty"` // rich brief (employees parity #10)
	Skills     []RefRead     `json:"skills"`
	JoinDate   *time.Time    `json:"join_date,omitempty"`

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

	// Leave quota (employees parity #5) — hydrated from employee_leave_quotas;
	// falls back to the seeded defaults (12 / 6) when no row exists, so the
	// FE always sees concrete numbers like the Python shape did.
	AnnualLeaveQuota float64 `json:"annual_leave_quota"`
	SickLeaveQuota   float64 `json:"sick_leave_quota"`

	// Auth side
	IsActive  bool      `json:"is_active"`
	Roles     []RoleRef `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Optional embed for nested dependent reads on /employees/me
	Dependents []DependentRead `json:"dependents,omitempty"`
}

// NOTE: EmployeeSummary is already declared in internal/dto/auth.go (created in
// Phase 2 Tasks 1-7) and is reused here by UserMeRead. It is intentionally NOT
// redeclared in this file to avoid a duplicate-type conflict; the auth.go shape
// (ID / FirstName / LastName / AvatarURL / DepartmentID / PositionID / ManagerID) satisfies
// both the auth login flow and GET /users/me.

// ---- Employee Create (admin) — accepts both user creds and HR fields ----

type EmployeeCreate struct {
	// User credentials (created in tx).
	// Password is intentionally not accepted from the HTTP payload (json:"-") —
	// it must be set via the forgot-password / invite-accept flow after creation,
	// matching Python's passwordless admin-create behaviour.
	// InviteService.Accept populates this field in Go code before calling
	// empSvc.Create, so invite-based accounts still get a password on first use.
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"-"`
	IsActive *bool  `json:"is_active,omitempty"`

	// When true, a "set your password" email is sent to the new employee using
	// the same forgot-password token mechanism (PasswordResetService.RequestReset).
	SendInvite bool `json:"send_invite,omitempty"`

	// HR personal info
	FirstName               string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName                string     `json:"last_name"  binding:"required,min=1,max=100"`
	Phone                   *string    `json:"phone,omitempty"`
	PersonalEmail           *string    `json:"personal_email,omitempty" binding:"omitempty,optemail"`
	Gender                  *string    `json:"gender,omitempty"          binding:"omitempty,oneof=male female other"`
	DOB                     *time.Time `json:"dob,omitempty"`
	Nationality             *string    `json:"nationality,omitempty"`
	IDNumber                *string    `json:"id_number,omitempty"`
	IDIssueDate             *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage            *string    `json:"id_front_image,omitempty"`
	IDBackImage             *string    `json:"id_back_image,omitempty"`
	PermanentAddress        *string    `json:"permanent_address,omitempty"`
	CurrentAddress          *string    `json:"current_address,omitempty"`
	Education               *string    `json:"education,omitempty"        binding:"omitempty,oneof=high_school college bachelor master doctorate"`
	MaritalStatus           *string    `json:"marital_status,omitempty"  binding:"omitempty,oneof=single married other"`
	ExperienceYear          *int       `json:"experience_year,omitempty"` // career-start year; range validated in the service (validateExperienceYear)
	CVURL                   *string    `json:"cv_url,omitempty"`
	SocialInsuranceNumber   *string    `json:"social_insurance_number,omitempty"   binding:"omitempty,max=50"`
	TaxIdentificationNumber *string    `json:"tax_identification_number,omitempty" binding:"omitempty,max=50"`

	// Emergency contacts — full replacement list at create (employees parity #4).
	EmergencyContacts []EmergencyContactInput `json:"emergency_contacts,omitempty" binding:"omitempty,dive"`

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

	// Skills assigned at creation (inline — Python parity). Empty/absent = none.
	SkillIDs []uuid.UUID `json:"skill_ids,omitempty"`
}

// ---- Employee Update (admin — anything allowed) ----

type EmployeeUpdate struct {
	FirstName               *string    `json:"first_name,omitempty" binding:"omitempty,min=1,max=100"`
	LastName                *string    `json:"last_name,omitempty"  binding:"omitempty,min=1,max=100"`
	Phone                   *string    `json:"phone,omitempty"`
	PersonalEmail           *string    `json:"personal_email,omitempty" binding:"omitempty,optemail"`
	Gender                  *string    `json:"gender,omitempty"          binding:"omitempty,oneof=male female other"`
	DOB                     *time.Time `json:"dob,omitempty"`
	Nationality             *string    `json:"nationality,omitempty"`
	IDNumber                *string    `json:"id_number,omitempty"`
	IDIssueDate             *time.Time `json:"id_issue_date,omitempty"`
	IDFrontImage            *string    `json:"id_front_image,omitempty"`
	IDBackImage             *string    `json:"id_back_image,omitempty"`
	PermanentAddress        *string    `json:"permanent_address,omitempty"`
	CurrentAddress          *string    `json:"current_address,omitempty"`
	Education               *string    `json:"education,omitempty"        binding:"omitempty,oneof=high_school college bachelor master doctorate"`
	MaritalStatus           *string    `json:"marital_status,omitempty"  binding:"omitempty,oneof=single married other"`
	ExperienceYear          *int       `json:"experience_year,omitempty"` // career-start year; range validated in the service (validateExperienceYear)
	CVURL                   *string    `json:"cv_url,omitempty"`
	SocialInsuranceNumber   *string    `json:"social_insurance_number,omitempty"   binding:"omitempty,max=50"`
	TaxIdentificationNumber *string    `json:"tax_identification_number,omitempty" binding:"omitempty,max=50"`

	// Emergency contacts — pointer-to-slice partial-PATCH semantics:
	// nil/absent = leave unchanged, [] = clear all, non-empty = replace the set.
	EmergencyContacts *[]EmergencyContactInput `json:"emergency_contacts,omitempty" binding:"omitempty,dive"`

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

	// Skills replace-set: nil/absent = leave unchanged, [] = clear all,
	// non-empty = replace the whole set (inline — Python parity).
	SkillIDs *[]uuid.UUID `json:"skill_ids,omitempty"`
}

// ---- Employee Self Update (RESTRICTED — server-side whitelist) ----
//
// Per audit decision #7 the self-editable set now includes the personal
// identity fields full_name / gender / dob (matching Python's /users/me).
//
// IMPORTANT: This DTO MUST NOT contain any of: department_id, position_id, manager_id,
// basic_salary, insurance_salary, contract_*, role_ids, is_active.
// The service enforces the whitelist with a manual field-by-field copy, so
// even a future DTO widening cannot mutate those admin-only columns.
type EmployeeSelfUpdate struct {
	FirstName               *string    `json:"first_name,omitempty"     binding:"omitempty,min=1,max=100"`
	LastName                *string    `json:"last_name,omitempty"      binding:"omitempty,min=1,max=100"`
	Gender                  *string    `json:"gender,omitempty"         binding:"omitempty,oneof=male female other"`
	DOB                     *time.Time `json:"dob,omitempty"`
	Phone                   *string    `json:"phone,omitempty"`
	PersonalEmail           *string    `json:"personal_email,omitempty" binding:"omitempty,optemail"`
	PermanentAddress        *string    `json:"permanent_address,omitempty"`
	CurrentAddress          *string    `json:"current_address,omitempty"`
	MaritalStatus           *string    `json:"marital_status,omitempty" binding:"omitempty,oneof=single married other"`
	SocialInsuranceNumber   *string    `json:"social_insurance_number,omitempty"   binding:"omitempty,max=50"`
	TaxIdentificationNumber *string    `json:"tax_identification_number,omitempty" binding:"omitempty,max=50"`

	// Emergency contacts — self may manage their own list (pointer-to-slice:
	// nil = leave unchanged, [] = clear all, non-empty = replace the set).
	EmergencyContacts *[]EmergencyContactInput `json:"emergency_contacts,omitempty" binding:"omitempty,dive"`
}

// ---- Leave quota update (admin) ----

type LeaveQuotaUpdateRequest struct {
	AnnualLeaveQuota float64 `json:"annual_leave_quota" binding:"gte=0,lte=365"`
	SickLeaveQuota   float64 `json:"sick_leave_quota"   binding:"gte=0,lte=365"`
}

// ---- List query ----

type EmployeeListQuery struct {
	Page     int    `form:"page,default=1"       binding:"gte=1"`
	PageSize int    `form:"page_size,default=20" binding:"gte=1,lte=100"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`

	// Raw repeated query params (HTTP boundary). gin binds repeated keys
	// (?department_id=a&department_id=b) into these string slices; uuid.UUID
	// cannot be bound directly from a query param (that was the 400 bug).
	DepartmentIDsRaw []string `form:"department_id"`
	PositionIDsRaw   []string `form:"position_id"`
	ManagerIDsRaw    []string `form:"manager_id"`
	RoleIDsRaw       []string `form:"role_id"`

	// Parsed by the handler (ParseFilters); consumed by the repo. Not bound.
	DepartmentIDs []uuid.UUID `form:"-"`
	PositionIDs   []uuid.UUID `form:"-"`
	ManagerIDs    []uuid.UUID `form:"-"`
	RoleIDs       []uuid.UUID `form:"-"`
}

// ParseFilters converts the raw repeated-param strings into parsed UUID
// slices, returning a 400 AppError on the first invalid value. Empty/absent
// params yield empty slices (= no filter). Duplicate values are passed
// through unchanged; PostgreSQL treats IN (X, X) as IN (X) so no dedup is
// performed.
func (q *EmployeeListQuery) ParseFilters() error {
	parse := func(name string, raw []string, dst *[]uuid.UUID) error {
		for _, s := range raw {
			if s == "" {
				continue
			}
			id, err := uuid.Parse(s)
			if err != nil {
				return apperrors.ErrBadRequest("invalid " + name)
			}
			*dst = append(*dst, id)
		}
		return nil
	}
	if err := parse("department_id", q.DepartmentIDsRaw, &q.DepartmentIDs); err != nil {
		return err
	}
	if err := parse("position_id", q.PositionIDsRaw, &q.PositionIDs); err != nil {
		return err
	}
	if err := parse("manager_id", q.ManagerIDsRaw, &q.ManagerIDs); err != nil {
		return err
	}
	return parse("role_id", q.RoleIDsRaw, &q.RoleIDs)
}
