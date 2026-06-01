package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// defaultEmployeeRoleName is the system role assigned to a newly created
// employee when the admin does not supply explicit role_ids. It must match
// the role name created by the seed service.
const defaultEmployeeRoleName = "Employee"

// managerTreeLockKey is a fixed key for the Postgres transaction-level advisory
// lock that serializes line-manager reparenting (employees parity #10). All
// reparent operations contend on this single key so the in-tx cycle re-check
// always sees other reparents' committed state.
const managerTreeLockKey int64 = 0x6D67727472 // "mgrtr"

// boolToRenewal converts the DTO *bool contract-renewal flag to the model's
// integer column (1 = renew, 0 = do not renew).
func boolToRenewal(b *bool) int {
	if b != nil && !*b {
		return 0
	}
	return 1
}

// validateExperienceYear enforces the BA contract (DR-001-005-02/03/04):
// experience_year is a career-start (4-digit) year, must be > 1900 and not in
// the future. nil = not provided = valid.
func validateExperienceYear(y *int) error {
	if y == nil {
		return nil
	}
	cur := time.Now().UTC().Year()
	if *y <= 1900 || *y > cur {
		return apperrors.ErrBadRequest(fmt.Sprintf("experience_year must be a year between 1901 and %d", cur))
	}
	return nil
}

// toEmergencyModels converts the DTO emergency-contact inputs to model rows
// (trimmed), for the ReplaceEmergencyContacts repo call (employees parity #4).
func toEmergencyModels(in []dto.EmergencyContactInput) []models.EmployeeEmergencyContact {
	out := make([]models.EmployeeEmergencyContact, 0, len(in))
	for _, c := range in {
		out = append(out, models.EmployeeEmergencyContact{
			FullName:     strings.TrimSpace(c.FullName),
			Relationship: strings.TrimSpace(c.Relationship),
			PhoneNumber:  strings.TrimSpace(c.PhoneNumber),
		})
	}
	return out
}

// positionName / departmentName resolve the preloaded Position/Department
// names off an employee (nil when unset or not preloaded). Used for the rich
// manager brief + line-manager picker/direct-report rows (employees parity #10).
func positionName(e *models.Employee) *string {
	if e != nil && e.Position != nil && e.Position.Name != "" {
		n := e.Position.Name
		return &n
	}
	return nil
}

func departmentName(e *models.Employee) *string {
	if e != nil && e.Department != nil && e.Department.Name != "" {
		n := e.Department.Name
		return &n
	}
	return nil
}

func departmentRef(e *models.Employee) *dto.RefRead {
	if e != nil && e.Department != nil && e.Department.Name != "" {
		return &dto.RefRead{ID: e.Department.ID, Name: e.Department.Name}
	}
	return nil
}

func positionRef(e *models.Employee) *dto.RefRead {
	if e != nil && e.Position != nil && e.Position.Name != "" {
		return &dto.RefRead{ID: e.Position.ID, Name: e.Position.Name}
	}
	return nil
}

// EmployeeFieldPerms captures the caller's field-level salary/banking
// permissions (employees parity #6). The handler builds it from the caller's
// roles; the wildcard "*" grants all four. *_view gates whether the section
// is returned on reads; *_manage gates whether it may be set on write.
type EmployeeFieldPerms struct {
	SalaryView    bool
	SalaryManage  bool
	BankingView   bool
	BankingManage bool
}

// AllEmployeeFieldPerms grants every field-level perm — used for internal
// callers that legitimately need the full unmasked shape (e.g. tests).
var AllEmployeeFieldPerms = EmployeeFieldPerms{true, true, true, true}

// maskAccountNumber renders a bank account number as "•••• 1234" (last 4),
// matching the Python read shape. nil/empty passes through.
func maskAccountNumber(acct *string) *string {
	if acct == nil || *acct == "" {
		return acct
	}
	s := *acct
	last := s
	if len(s) > 4 {
		last = s[len(s)-4:]
	}
	masked := "•••• " + last
	return &masked
}

// ApplyEmployeeFieldVisibility gates and masks the salary/banking sections of
// an employee read in place, per the caller's field perms (employees parity
// #6). On reads (unmask=false) the account number is masked even for banking
// viewers; on write echoes (unmask=true) it is returned in full. Sections the
// caller may neither view nor manage are stripped to nil.
func ApplyEmployeeFieldVisibility(view *dto.EmployeeRead, p EmployeeFieldPerms, unmask bool) {
	if view == nil {
		return
	}
	if !(p.SalaryView || p.SalaryManage) {
		view.BasicSalary = nil
		view.InsuranceSalary = nil
	}
	if !(p.BankingView || p.BankingManage) {
		view.BankAccount = nil
		view.BankName = nil
		view.BankHolderName = nil
		view.PaymentMethod = nil
	} else if !unmask {
		view.BankAccount = maskAccountNumber(view.BankAccount)
	}
}

// GuardSalaryWrite / GuardBankingWrite return a forbidden error when a payload
// sets salary or banking fields the caller may not manage (employees parity
// #6). Pure functions — called by the handler before Create/Update (the
// codebase's handler-level authorization pattern) and unit-tested directly.
func GuardSalaryWrite(set bool, p EmployeeFieldPerms) error {
	if set && !p.SalaryManage {
		return apperrors.ErrForbidden("You do not have permission to set salary fields")
	}
	return nil
}

func GuardBankingWrite(set bool, p EmployeeFieldPerms) error {
	if set && !p.BankingManage {
		return apperrors.ErrForbidden("You do not have permission to set banking fields")
	}
	return nil
}

// skillAssigner is the slice of SkillService that EmployeeService needs to
// apply inline skill_ids on create/update. Kept narrow for testability;
// satisfied by *SkillService.
type skillAssigner interface {
	ValidateSkillIDs(ctx context.Context, skillIDs []uuid.UUID) ([]uuid.UUID, error)
	ReplaceForEmployee(ctx context.Context, employeeID uuid.UUID, skillIDs []uuid.UUID) ([]dto.SkillRead, error)
}

// EmployeeService owns the HR-profile business logic. All repository fields
// use the repository INTERFACE types for mockability.
type EmployeeService struct {
	db      *gorm.DB
	emps    repositories.EmployeeRepository
	deps    repositories.DependentRepository
	users   repositories.UserRepository
	roles   repositories.RoleRepository
	quota   repositories.LeaveQuotaRepository
	uploads Uploader
	skills  skillAssigner
}

func NewEmployeeService(
	db *gorm.DB,
	emps repositories.EmployeeRepository,
	deps repositories.DependentRepository,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
	skills skillAssigner,
) *EmployeeService {
	return &EmployeeService{db: db, emps: emps, deps: deps, users: users, roles: roles, quota: quota, uploads: uploads, skills: skills}
}

// ---- Read ----

func (s *EmployeeService) Get(ctx context.Context, id uuid.UUID) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	return s.toRead(e), nil
}

func (s *EmployeeService) GetByUserID(ctx context.Context, userID uuid.UUID) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByUserIDWithFull(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee profile")
		}
		return nil, err
	}
	return s.toRead(e), nil
}

func (s *EmployeeService) List(ctx context.Context, q dto.EmployeeListQuery) ([]dto.EmployeeRead, int64, error) {
	emps, total, err := s.emps.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.EmployeeRead, 0, len(emps))
	for i := range emps {
		out = append(out, *s.toRead(&emps[i]))
	}
	return out, total, nil
}

func (s *EmployeeService) toRead(e *models.Employee) *dto.EmployeeRead {
	roles := make([]dto.RoleRead, 0)
	if e.User != nil {
		for _, r := range e.User.Roles {
			roles = append(roles, dto.RoleRead{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				IsSystem:    r.IsSystem,
				Permissions: []string(r.Permissions),
			})
		}
	}
	var mgr *dto.ManagerBrief
	if e.Manager != nil {
		mgr = &dto.ManagerBrief{
			ID:         e.Manager.ID,
			FullName:   e.Manager.FullName(),
			Position:   positionName(e.Manager),
			Department: departmentName(e.Manager),
		}
		if e.Manager.User != nil {
			mgr.IsActive = e.Manager.User.IsActive
		}
	}
	deps := make([]dto.DependentRead, 0, len(e.Dependents))
	for _, d := range e.Dependents {
		deps = append(deps, dto.DependentRead{
			ID:           d.ID,
			EmployeeID:   d.EmployeeID,
			FullName:     d.FullName,
			DOB:          d.DOB,
			Gender:       d.Gender,
			Relationship: d.Relationship,
			CreatedAt:    d.CreatedAt,
			UpdatedAt:    d.UpdatedAt,
		})
	}

	// Emergency contacts — 1-N list (employees parity #4).
	ec := make([]dto.EmergencyContactRead, 0, len(e.EmergencyContacts))
	for _, c := range e.EmergencyContacts {
		ec = append(ec, dto.EmergencyContactRead{
			ID:           c.ID,
			FullName:     c.FullName,
			Relationship: c.Relationship,
			PhoneNumber:  c.PhoneNumber,
		})
	}

	// Skills embed (employees parity #8) — hydrated from employee_skills.Skill.
	skills := make([]dto.RefRead, 0, len(e.EmployeeSkills))
	for _, es := range e.EmployeeSkills {
		if es.Skill != nil {
			skills = append(skills, dto.RefRead{ID: es.Skill.ID, Name: es.Skill.Name})
		}
	}

	// Leave quota (employees parity #5) — defaults mirror the seeded column
	// defaults (12 / 6) when the employee has no quota row yet.
	annualQuota := 12.0
	sickQuota := 6.0
	if e.LeaveQuota != nil {
		annualQuota = e.LeaveQuota.AnnualLeaveQuota
		sickQuota = e.LeaveQuota.SickLeaveQuota
	}

	// Reconciliation with the Phase 1 models.Employee declaration: several
	// fields are non-pointer scalars on the model (ContractType string,
	// ContractRenewal int, BasicSalary/InsuranceSalary float64, PaymentMethod
	// string) while the DTO carries pointers — convert verbatim here.
	contractType := e.ContractType
	paymentMethod := e.PaymentMethod
	basicSalary := e.BasicSalary
	insuranceSalary := e.InsuranceSalary
	contractRenewal := e.ContractRenewal > 0

	out := &dto.EmployeeRead{
		ID:                e.ID,
		UserID:            e.UserID,
		FirstName:         e.FirstName,
		LastName:          e.LastName,
		Phone:             e.Phone,
		PersonalEmail:     e.PersonalEmail,
		Gender:            e.Gender,
		DOB:               e.DOB,
		Nationality:       e.Nationality,
		IDNumber:          e.IDNumber,
		IDIssueDate:       e.IDIssueDate,
		IDFrontImage:      e.IDFrontImage,
		IDBackImage:       e.IDBackImage,
		PermanentAddress:  e.PermanentAddress,
		CurrentAddress:    e.CurrentAddress,
		Education:         e.Education,
		MaritalStatus:     e.MaritalStatus,
		ExperienceYear:    e.ExperienceYear,
		CVURL:             e.CVURL,
		AvatarURL:         e.AvatarURL,
		EmergencyContacts: ec,
		Manager:           mgr,
		Skills:            skills,
		JoinDate:          e.JoinDate,
		ContractType:      &contractType,
		ContractSignDate:  e.ContractSignDate,
		ContractEndDate:   e.ContractEndDate,
		ContractRenewal:   &contractRenewal,
		BasicSalary:       &basicSalary,
		InsuranceSalary:   &insuranceSalary,
		BankAccount:       e.BankAccount,
		BankName:          e.BankName,
		BankHolderName:    e.BankHolderName,
		PaymentMethod:     &paymentMethod,
		AnnualLeaveQuota:  annualQuota,
		SickLeaveQuota:    sickQuota,
		Roles:             roles,
		Dependents:        deps,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
	if e.User != nil {
		out.Email = e.User.Email
		out.IsActive = e.User.IsActive
	}
	out.Department = departmentRef(e)
	out.Position = positionRef(e)
	return out
}

func (s *EmployeeService) toSummary(e *models.Employee) *dto.EmployeeSummary {
	return &dto.EmployeeSummary{
		ID:           e.ID,
		FirstName:    e.FirstName,
		LastName:     e.LastName,
		AvatarURL:    e.AvatarURL,
		DepartmentID: e.DepartmentID,
		PositionID:   e.PositionID,
		ManagerID:    e.ManagerID,
		// Department/Position refs filled in Phase 3.
	}
}

// ---- Admin Create — single tx: user + employee + role assignment ----

func (s *EmployeeService) Create(ctx context.Context, in dto.EmployeeCreate) (*dto.EmployeeRead, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	exists, err := s.users.ExistsByEmail(ctx, email, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict("A user with this email already exists")
	}
	// Validate the line-manager assignment (employees parity #10). On create the
	// employee does not exist yet, so only existence + active are checked (no
	// cycle possible — nobody reports to a not-yet-created employee).
	if in.ManagerID != nil {
		if err := s.validateManagerAssignment(ctx, s.emps, *in.ManagerID, uuid.Nil); err != nil {
			return nil, err
		}
	}
	if err := validateExperienceYear(in.ExperienceYear); err != nil {
		return nil, err
	}
	if len(in.SkillIDs) > 0 {
		if _, err := s.skills.ValidateSkillIDs(ctx, in.SkillIDs); err != nil {
			return nil, err
		}
	}
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, apperrors.ErrBadRequest("failed to hash password")
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}

	var createdEmp models.Employee
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		u := &models.User{
			Email:        email,
			PasswordHash: hash,
			IsActive:     active,
		}
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		// The users.is_active column is `default:true`, so GORM drops a
		// zero-value (false) on INSERT and the DB default silently reactivates
		// an intentionally-inactive user. Force the value explicitly when the
		// caller asked for an inactive account.
		if !active {
			if err := tx.Model(&models.User{}).
				Where("id = ?", u.ID).
				Update("is_active", false).Error; err != nil {
				return err
			}
		}
		// Model/DTO reconciliation: ContractType, PaymentMethod are non-pointer
		// strings on the model (with NOT NULL defaults), salaries are float64,
		// ContractRenewal is an int. Map verbatim, falling back to model
		// defaults when the DTO pointer is nil.
		e := &models.Employee{
			UserID:           u.ID,
			FirstName:        strings.TrimSpace(in.FirstName),
			LastName:         strings.TrimSpace(in.LastName),
			Phone:            in.Phone,
			PersonalEmail:    in.PersonalEmail,
			Gender:           in.Gender,
			DOB:              in.DOB,
			Nationality:      in.Nationality,
			IDNumber:         in.IDNumber,
			IDIssueDate:      in.IDIssueDate,
			IDFrontImage:     in.IDFrontImage,
			IDBackImage:      in.IDBackImage,
			PermanentAddress: in.PermanentAddress,
			CurrentAddress:   in.CurrentAddress,
			Education:        in.Education,
			MaritalStatus:    in.MaritalStatus,
			ExperienceYear:   in.ExperienceYear,
			CVURL:            in.CVURL,
			DepartmentID:     in.DepartmentID,
			PositionID:       in.PositionID,
			ManagerID:        in.ManagerID,
			JoinDate:         in.JoinDate,
			ContractSignDate: in.ContractSignDate,
			ContractEndDate:  in.ContractEndDate,
			ContractRenewal:  boolToRenewal(in.ContractRenewal),
			BankAccount:      in.BankAccount,
			BankName:         in.BankName,
			BankHolderName:   in.BankHolderName,
		}
		if in.ContractType != nil && strings.TrimSpace(*in.ContractType) != "" {
			e.ContractType = strings.TrimSpace(*in.ContractType)
		} else {
			e.ContractType = "official"
		}
		if in.PaymentMethod != nil && strings.TrimSpace(*in.PaymentMethod) != "" {
			e.PaymentMethod = strings.TrimSpace(*in.PaymentMethod)
		} else {
			e.PaymentMethod = "bank_transfer"
		}
		if in.BasicSalary != nil {
			e.BasicSalary = *in.BasicSalary
		}
		if in.InsuranceSalary != nil {
			e.InsuranceSalary = *in.InsuranceSalary
		}
		if err := tx.Create(e).Error; err != nil {
			return err
		}
		// Resolve which roles to attach. When the admin supplies explicit
		// role_ids, honour them. Otherwise fall back to the system default
		// "Employee" role so every created employee is a usable self-service
		// account (carries auth:login) — without this, a freshly created
		// employee has zero permissions and cannot log in. The seed service
		// guarantees this role exists on every boot; if it is genuinely
		// missing we log and proceed role-less rather than failing the whole
		// creation (the admin can assign roles afterwards).
		roleIDs := in.RoleIDs
		if len(roleIDs) == 0 {
			defRole, derr := s.roles.FindByName(ctx, defaultEmployeeRoleName)
			switch {
			case derr == nil:
				roleIDs = []uuid.UUID{defRole.ID}
			case errors.Is(derr, gorm.ErrRecordNotFound):
				log.Printf("employees: default %q role not found; "+
					"creating user %q with no roles (cannot self-login until "+
					"a role is assigned)", defaultEmployeeRoleName, email)
			default:
				return derr
			}
		}
		if len(roleIDs) > 0 {
			rows := make([]map[string]any, 0, len(roleIDs))
			for _, rid := range roleIDs {
				rows = append(rows, map[string]any{
					"user_id": u.ID, "role_id": rid,
					"created_at": gorm.Expr("NOW()"), "updated_at": gorm.Expr("NOW()"),
					"is_deleted": false,
				})
			}
			if err := tx.Table("user_roles").Create(&rows).Error; err != nil {
				return err
			}
		}
		createdEmp = *e
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	// Persist emergency contacts (employees parity #4) once the user+employee
	// row is committed. Empty slice = none.
	if len(in.EmergencyContacts) > 0 {
		if err := s.emps.ReplaceEmergencyContacts(ctx, createdEmp.ID, toEmergencyModels(in.EmergencyContacts)); err != nil {
			return nil, err
		}
	}
	// Apply skills post-commit (same shape as emergency contacts). IDs were
	// pre-validated before the tx, so the common bad-id case is already ruled
	// out; a rare post-commit failure here (e.g. a concurrent skill delete)
	// returns an error to the caller, who can retry via PUT /employees/:id/skills
	// — the employee row stays, matching the emergency-contacts trade-off.
	if len(in.SkillIDs) > 0 {
		if _, err := s.skills.ReplaceForEmployee(ctx, createdEmp.ID, in.SkillIDs); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, createdEmp.ID)
}

// ---- Admin Update (anything allowed) ----

func (s *EmployeeService) Update(ctx context.Context, id uuid.UUID, in dto.EmployeeUpdate, callerUserID uuid.UUID) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}

	// Cheap input validation before any state-dependent guard.
	if err := validateExperienceYear(in.ExperienceYear); err != nil {
		return nil, err
	}

	// Cannot deactivate your own account (employees parity #12).
	if in.IsActive != nil && !*in.IsActive && e.UserID == callerUserID {
		return nil, apperrors.ErrBadRequest("You cannot deactivate your own account")
	}

	// Line-manager re-parent (employees parity #10). The authoritative
	// validation runs INSIDE the write tx under an advisory lock (below) to
	// close the cycle-check TOCTOU — a pre-tx check would race with a
	// concurrent reparent and could let two requests each commit half a cycle.
	setManager := (in.ClearManager == nil || !*in.ClearManager) && in.ManagerID != nil

	fields := map[string]any{}
	setIfNotNilStr := func(key string, v *string) {
		if v != nil {
			fields[key] = strings.TrimSpace(*v)
		}
	}
	setIfNotNilStr("first_name", in.FirstName)
	setIfNotNilStr("last_name", in.LastName)
	setIfNotNilStr("phone", in.Phone)
	setIfNotNilStr("personal_email", in.PersonalEmail)
	setIfNotNilStr("gender", in.Gender)
	setIfNotNilStr("nationality", in.Nationality)
	setIfNotNilStr("id_number", in.IDNumber)
	setIfNotNilStr("id_front_image", in.IDFrontImage)
	setIfNotNilStr("id_back_image", in.IDBackImage)
	setIfNotNilStr("permanent_address", in.PermanentAddress)
	setIfNotNilStr("current_address", in.CurrentAddress)
	setIfNotNilStr("education", in.Education)
	setIfNotNilStr("marital_status", in.MaritalStatus)
	setIfNotNilStr("contract_type", in.ContractType)
	setIfNotNilStr("bank_account", in.BankAccount)
	setIfNotNilStr("bank_name", in.BankName)
	setIfNotNilStr("bank_holder_name", in.BankHolderName)
	setIfNotNilStr("payment_method", in.PaymentMethod)
	if in.DOB != nil {
		fields["dob"] = *in.DOB
	}
	if in.IDIssueDate != nil {
		fields["id_issue_date"] = *in.IDIssueDate
	}
	if in.JoinDate != nil {
		fields["join_date"] = *in.JoinDate
	}
	if in.ContractSignDate != nil {
		fields["contract_sign_date"] = *in.ContractSignDate
	}
	if in.ContractEndDate != nil {
		fields["contract_end_date"] = *in.ContractEndDate
	}
	if in.ContractRenewal != nil {
		fields["contract_renewal"] = boolToRenewal(in.ContractRenewal)
	}
	if in.BasicSalary != nil {
		fields["basic_salary"] = *in.BasicSalary
	}
	if in.InsuranceSalary != nil {
		fields["insurance_salary"] = *in.InsuranceSalary
	}
	if in.ExperienceYear != nil {
		fields["experience_year"] = *in.ExperienceYear
	}
	setIfNotNilStr("cv_url", in.CVURL)
	// FK clearing wins over set.
	if in.ClearDept != nil && *in.ClearDept {
		fields["department_id"] = nil
	} else if in.DepartmentID != nil {
		fields["department_id"] = *in.DepartmentID
	}
	if in.ClearPos != nil && *in.ClearPos {
		fields["position_id"] = nil
	} else if in.PositionID != nil {
		fields["position_id"] = *in.PositionID
	}
	if in.ClearManager != nil && *in.ClearManager {
		fields["manager_id"] = nil
	} else if in.ManagerID != nil {
		fields["manager_id"] = *in.ManagerID
	}

	// Atomic: the employee-fields write and the user active toggle must
	// commit or roll back together so a partial admin update cannot leave
	// the employee row and its auth user in an inconsistent state. Mirrors
	// the transaction in SoftDelete; writes go through tx directly.
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reparenting: serialize via a tx advisory lock and re-validate against
		// committed state so a concurrent reparent cannot slip a cycle past the
		// check (employees parity #10 — closes the TOCTOU).
		if setManager {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", managerTreeLockKey).Error; err != nil {
				return err
			}
			if err := s.validateManagerAssignment(ctx, s.emps.WithTx(tx), *in.ManagerID, e.ID); err != nil {
				return err
			}
		}
		if len(fields) > 0 {
			if err := tx.Model(&models.Employee{}).
				Where("id = ? AND is_deleted = ?", e.ID, false).
				Updates(fields).Error; err != nil {
				return err
			}
		}
		if in.IsActive != nil {
			// UserID FK is always set — no nil guard needed.
			if err := tx.Model(&models.User{}).
				Where("id = ? AND is_deleted = ?", e.UserID, false).
				Update("is_active", *in.IsActive).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// Emergency contacts replace-set (employees parity #4):
	// nil = leave unchanged, [] = clear all, non-empty = replace the set.
	if in.EmergencyContacts != nil {
		if err := s.emps.ReplaceEmergencyContacts(ctx, e.ID, toEmergencyModels(*in.EmergencyContacts)); err != nil {
			return nil, err
		}
	}
	if in.SkillIDs != nil {
		if _, err := s.skills.ReplaceForEmployee(ctx, e.ID, *in.SkillIDs); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, id)
}

// ---- Self Update — HARD WHITELIST ----
//
// Enforced SERVER-SIDE: only phone, personal_email, permanent_address,
// current_address, marital_status, emergency_contact_* may be written. Even if
// a future DTO change widened EmployeeSelfUpdate, the field-by-field copy below
// makes it structurally impossible to mutate department_id, position_id,
// manager_id, basic_salary, insurance_salary, contract_*, role_ids, is_active.
func (s *EmployeeService) SelfUpdate(ctx context.Context, userID uuid.UUID, in dto.EmployeeSelfUpdate) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByUserIDWithFull(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee profile")
		}
		return nil, err
	}
	allowed := map[string]any{}
	// Identity fields — self-editable per audit decision #7.
	if in.FirstName != nil {
		allowed["first_name"] = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		allowed["last_name"] = strings.TrimSpace(*in.LastName)
	}
	if in.Gender != nil {
		allowed["gender"] = *in.Gender
	}
	if in.DOB != nil {
		allowed["dob"] = *in.DOB
	}
	if in.Phone != nil {
		allowed["phone"] = strings.TrimSpace(*in.Phone)
	}
	if in.PersonalEmail != nil {
		allowed["personal_email"] = strings.ToLower(strings.TrimSpace(*in.PersonalEmail))
	}
	if in.PermanentAddress != nil {
		allowed["permanent_address"] = strings.TrimSpace(*in.PermanentAddress)
	}
	if in.CurrentAddress != nil {
		allowed["current_address"] = strings.TrimSpace(*in.CurrentAddress)
	}
	if in.MaritalStatus != nil {
		allowed["marital_status"] = *in.MaritalStatus
	}
	if err := s.emps.UpdateFields(ctx, e.ID, allowed); err != nil {
		return nil, err
	}
	// Emergency contacts replace-set — self may manage their own (#4 + #7):
	// nil = leave unchanged, [] = clear all, non-empty = replace the set.
	if in.EmergencyContacts != nil {
		if err := s.emps.ReplaceEmergencyContacts(ctx, e.ID, toEmergencyModels(*in.EmergencyContacts)); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, e.ID)
}

// ---- Soft delete (cascading: soft-delete employee + deactivate user) ----

func (s *EmployeeService) SoftDelete(ctx context.Context, id uuid.UUID, callerUserID uuid.UUID) error {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Employee")
		}
		return err
	}
	// Cannot delete your own account (employees parity #12).
	if e.UserID == callerUserID {
		return apperrors.ErrBadRequest("You cannot delete your own account")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Employee{}).
			Where("id = ?", e.ID).
			Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error; err != nil {
			return err
		}
		// Clear manager_id on direct reports so no live employee points at a
		// now-deleted manager (employees parity #12). Use the map form (the
		// codebase's proven FK-clear pattern) — Update("col", nil) can be
		// dropped by GORM's zero-value handling.
		if err := tx.Model(&models.Employee{}).
			Where("manager_id = ? AND is_deleted = ?", e.ID, false).
			Updates(map[string]any{"manager_id": nil}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).
			Where("id = ?", e.UserID).
			Updates(map[string]any{"is_active": false}).Error
	})
}

// ---- Avatar ----

const (
	maxAvatarBytes = 5 * 1024 * 1024
	avatarSubdir   = "avatars"
)

// allowedAvatarMIME is the set of image content types permitted for an
// avatar. The authoritative check sniffs the actual file bytes rather than
// trusting the client-supplied Content-Type header.
var allowedAvatarMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func (s *EmployeeService) uploadAvatar(ctx context.Context, employeeID uuid.UUID, prev *string, content []byte, contentType, ext string) (*dto.EmployeeRead, error) {
	if len(content) > maxAvatarBytes {
		return nil, apperrors.ErrBadRequest("Avatar must not exceed 5MB")
	}
	// Authoritative content check: sniff the real bytes (RFC 2046 / WHATWG
	// mime-sniff via http.DetectContentType) instead of trusting the
	// client-supplied Content-Type header, which is attacker-controlled.
	sniffLen := len(content)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffed := http.DetectContentType(content[:sniffLen])
	if !allowedAvatarMIME[sniffed] {
		return nil, apperrors.ErrBadRequest("Avatar must be a valid image (PNG, JPEG, GIF, or WEBP)")
	}
	// Use the verified type for storage, not the client's header.
	contentType = sniffed
	url, err := s.uploads.Upload(ctx, avatarSubdir, ext, content, contentType)
	if err != nil {
		return nil, err
	}
	if err := s.emps.UpdateAvatarURL(ctx, employeeID, &url); err != nil {
		return nil, err
	}
	if prev != nil && *prev != "" {
		_ = s.uploads.Delete(ctx, *prev) // best-effort cleanup
	}
	return s.Get(ctx, employeeID)
}

func (s *EmployeeService) UpdateAvatarSelf(ctx context.Context, userID uuid.UUID, content []byte, contentType, ext string) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByUserIDWithFull(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee profile")
		}
		return nil, err
	}
	return s.uploadAvatar(ctx, e.ID, e.AvatarURL, content, contentType, ext)
}

func (s *EmployeeService) UpdateAvatarAdmin(ctx context.Context, id uuid.UUID, content []byte, contentType, ext string) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	return s.uploadAvatar(ctx, e.ID, e.AvatarURL, content, contentType, ext)
}

// ---- Leave quota (admin) ----

func (s *EmployeeService) UpdateLeaveQuota(ctx context.Context, id uuid.UUID, in dto.LeaveQuotaUpdateRequest) (*dto.EmployeeRead, error) {
	if _, err := s.emps.FindByIDWithFull(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	if err := s.quota.Upsert(ctx, id, in.AnnualLeaveQuota, in.SickLeaveQuota); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// ---- Line manager (employees parity #10) ----

// validateManagerAssignment enforces that a proposed line manager exists, is
// active, is not the target themselves, and is not within the target's
// subordinate chain (cycle prevention). managerID == uuid.Nil is a no-op
// (clearing the manager). targetID == uuid.Nil skips the self/cycle checks
// (used on create, before the employee exists).
// validateManagerAssignment runs against the supplied repo so it can be invoked
// either against the base DB (fast pre-check) or against a tx-bound repo for the
// authoritative in-transaction re-check (see Update's advisory-locked block).
func (s *EmployeeService) validateManagerAssignment(ctx context.Context, repo repositories.EmployeeRepository, managerID, targetID uuid.UUID) error {
	// Callers only invoke this when a manager id was actually supplied, so a
	// zero/Nil id here is a bad client value (not "no manager") — let the
	// existence check below reject it as a clean 400 rather than skipping.
	mgr, err := repo.FindByIDWithUser(ctx, managerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrBadRequest("Selected line manager does not exist")
		}
		return err
	}
	if mgr.User == nil || !mgr.User.IsActive {
		return apperrors.ErrBadRequest("Selected line manager is no longer active")
	}
	if targetID != uuid.Nil {
		if managerID == targetID {
			return apperrors.ErrBadRequest("Cannot set line manager to self")
		}
		chain, err := repo.SubordinateIDs(ctx, targetID)
		if err != nil {
			return err
		}
		if chain[managerID] {
			return apperrors.ErrBadRequest("Cannot assign — the selected manager is in this employee's reporting chain (would create a cycle)")
		}
	}
	return nil
}

// ManagerCandidates returns the line-manager picker options: active, non-deleted
// employees, excluding the target and its transitive subordinate chain (cycle
// prevention). When the target's currently-assigned manager is deactivated, it
// is kept in the list so the admin can preserve the historical assignment.
func (s *EmployeeService) ManagerCandidates(ctx context.Context, forEmployeeID *uuid.UUID, search string, limit int) ([]dto.ManagerCandidateRead, error) {
	var exclude []uuid.UUID
	var legacyManager *models.Employee
	if forEmployeeID != nil {
		target, err := s.emps.FindByIDWithUser(ctx, *forEmployeeID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if target != nil {
			exclude = append(exclude, target.ID)
			chain, err := s.emps.SubordinateIDs(ctx, target.ID)
			if err != nil {
				return nil, err
			}
			for id := range chain {
				exclude = append(exclude, id)
			}
			if target.ManagerID != nil {
				cm, err := s.emps.FindByIDWithOrg(ctx, *target.ManagerID)
				if err == nil && cm.User != nil && !cm.User.IsActive {
					legacyManager = cm
				}
			}
		}
	}
	emps, err := s.emps.ListManagerCandidates(ctx, exclude, search, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ManagerCandidateRead, 0, len(emps)+1)
	seen := make(map[uuid.UUID]bool, len(emps))
	for i := range emps {
		seen[emps[i].ID] = true
		out = append(out, toManagerCandidate(&emps[i]))
	}
	if legacyManager != nil && !seen[legacyManager.ID] {
		out = append(out, toManagerCandidate(legacyManager))
		// The appended legacy manager would otherwise sort last; re-sort so it
		// lands in its alphabetical slot (case-insensitive, matching the
		// LOWER(first_name), LOWER(last_name) ordering used in the repo query).
		sort.SliceStable(out, func(i, j int) bool {
			return strings.ToLower(out[i].FullName) < strings.ToLower(out[j].FullName)
		})
	}
	return out, nil
}

// DirectReports returns all live employees whose line manager is managerID
// (active AND inactive), sorted by name.
func (s *EmployeeService) DirectReports(ctx context.Context, managerID uuid.UUID) ([]dto.DirectReportRead, error) {
	if _, err := s.emps.FindByID(ctx, managerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	emps, err := s.emps.ListDirectReports(ctx, managerID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DirectReportRead, 0, len(emps))
	for i := range emps {
		out = append(out, toDirectReport(&emps[i]))
	}
	return out, nil
}

func toManagerCandidate(e *models.Employee) dto.ManagerCandidateRead {
	active := false
	if e.User != nil {
		active = e.User.IsActive
	}
	return dto.ManagerCandidateRead{
		ID:         e.ID,
		FullName:   e.FullName(),
		Position:   positionName(e),
		Department: departmentName(e),
		IsActive:   active,
	}
}

func toDirectReport(e *models.Employee) dto.DirectReportRead {
	active := false
	if e.User != nil {
		active = e.User.IsActive
	}
	return dto.DirectReportRead{
		ID:         e.ID,
		FullName:   e.FullName(),
		AvatarURL:  e.AvatarURL,
		Position:   positionName(e),
		Department: departmentName(e),
		IsActive:   active,
	}
}
