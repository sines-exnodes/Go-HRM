package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// boolToRenewal converts the DTO *bool contract-renewal flag to the model's
// integer column (1 = renew, 0 = do not renew).
func boolToRenewal(b *bool) int {
	if b != nil && !*b {
		return 0
	}
	return 1
}

// EmployeeService owns the HR-profile business logic. Repository fields use the
// repository INTERFACE types where the repo exposes one (employee/user/role)
// and concrete struct pointers for the struct-only repos (dependent/quota).
type EmployeeService struct {
	db      *gorm.DB
	emps    repositories.EmployeeRepository
	deps    *repositories.DependentRepository
	users   repositories.UserRepository
	roles   repositories.RoleRepository
	quota   *repositories.LeaveQuotaRepository
	uploads Uploader
}

func NewEmployeeService(
	db *gorm.DB,
	emps repositories.EmployeeRepository,
	deps *repositories.DependentRepository,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	quota *repositories.LeaveQuotaRepository,
	uploads Uploader,
) *EmployeeService {
	return &EmployeeService{db: db, emps: emps, deps: deps, users: users, roles: roles, quota: quota, uploads: uploads}
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
	var mgr *dto.RefRead
	if e.Manager != nil {
		mgr = &dto.RefRead{ID: e.Manager.ID, Name: e.Manager.FullName}
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
		ID:                       e.ID,
		UserID:                   e.UserID,
		FullName:                 e.FullName,
		Phone:                    e.Phone,
		PersonalEmail:            e.PersonalEmail,
		Gender:                   e.Gender,
		DOB:                      e.DOB,
		Nationality:              e.Nationality,
		IDNumber:                 e.IDNumber,
		IDIssueDate:              e.IDIssueDate,
		IDFrontImage:             e.IDFrontImage,
		IDBackImage:              e.IDBackImage,
		PermanentAddress:         e.PermanentAddress,
		CurrentAddress:           e.CurrentAddress,
		Education:                e.Education,
		MaritalStatus:            e.MaritalStatus,
		EmergencyContactName:     e.EmergencyContactName,
		EmergencyContactRelation: e.EmergencyContactRelation,
		EmergencyContactPhone:    e.EmergencyContactPhone,
		AvatarURL:                e.AvatarURL,
		Manager:                  mgr,
		JoinDate:                 e.JoinDate,
		ContractType:             &contractType,
		ContractSignDate:         e.ContractSignDate,
		ContractEndDate:          e.ContractEndDate,
		ContractRenewal:          &contractRenewal,
		BasicSalary:              &basicSalary,
		InsuranceSalary:          &insuranceSalary,
		BankAccount:              e.BankAccount,
		BankName:                 e.BankName,
		BankHolderName:           e.BankHolderName,
		PaymentMethod:            &paymentMethod,
		Roles:                    roles,
		Dependents:               deps,
		CreatedAt:                e.CreatedAt,
		UpdatedAt:                e.UpdatedAt,
	}
	if e.User != nil {
		out.Email = e.User.Email
		out.IsActive = e.User.IsActive
	}
	// Department/Position refs preloaded in Phase 3; intentionally nil until then.
	return out
}

func (s *EmployeeService) toSummary(e *models.Employee) *dto.EmployeeSummary {
	return &dto.EmployeeSummary{
		ID:           e.ID,
		FullName:     e.FullName,
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
			UserID:                   u.ID,
			FullName:                 strings.TrimSpace(in.FullName),
			Phone:                    in.Phone,
			PersonalEmail:            in.PersonalEmail,
			Gender:                   in.Gender,
			DOB:                      in.DOB,
			Nationality:              in.Nationality,
			IDNumber:                 in.IDNumber,
			IDIssueDate:              in.IDIssueDate,
			IDFrontImage:             in.IDFrontImage,
			IDBackImage:              in.IDBackImage,
			PermanentAddress:         in.PermanentAddress,
			CurrentAddress:           in.CurrentAddress,
			Education:                in.Education,
			MaritalStatus:            in.MaritalStatus,
			EmergencyContactName:     in.EmergencyContactName,
			EmergencyContactRelation: in.EmergencyContactRelation,
			EmergencyContactPhone:    in.EmergencyContactPhone,
			DepartmentID:             in.DepartmentID,
			PositionID:               in.PositionID,
			ManagerID:                in.ManagerID,
			JoinDate:                 in.JoinDate,
			ContractSignDate:         in.ContractSignDate,
			ContractEndDate:          in.ContractEndDate,
			ContractRenewal:          boolToRenewal(in.ContractRenewal),
			BankAccount:              in.BankAccount,
			BankName:                 in.BankName,
			BankHolderName:           in.BankHolderName,
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
		if len(in.RoleIDs) > 0 {
			rows := make([]map[string]any, 0, len(in.RoleIDs))
			for _, rid := range in.RoleIDs {
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
	return s.Get(ctx, createdEmp.ID)
}

// ---- Admin Update (anything allowed) ----

func (s *EmployeeService) Update(ctx context.Context, id uuid.UUID, in dto.EmployeeUpdate) (*dto.EmployeeRead, error) {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}

	fields := map[string]any{}
	setIfNotNilStr := func(key string, v *string) {
		if v != nil {
			fields[key] = strings.TrimSpace(*v)
		}
	}
	setIfNotNilStr("full_name", in.FullName)
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
	setIfNotNilStr("emergency_contact_name", in.EmergencyContactName)
	setIfNotNilStr("emergency_contact_relation", in.EmergencyContactRelation)
	setIfNotNilStr("emergency_contact_phone", in.EmergencyContactPhone)
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

	if err := s.emps.UpdateFields(ctx, e.ID, fields); err != nil {
		return nil, err
	}
	if in.IsActive != nil && e.User != nil {
		if err := s.users.ToggleActive(ctx, e.UserID, *in.IsActive); err != nil {
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
	if in.EmergencyContactName != nil {
		allowed["emergency_contact_name"] = strings.TrimSpace(*in.EmergencyContactName)
	}
	if in.EmergencyContactRelation != nil {
		allowed["emergency_contact_relation"] = strings.TrimSpace(*in.EmergencyContactRelation)
	}
	if in.EmergencyContactPhone != nil {
		allowed["emergency_contact_phone"] = strings.TrimSpace(*in.EmergencyContactPhone)
	}
	if err := s.emps.UpdateFields(ctx, e.ID, allowed); err != nil {
		return nil, err
	}
	return s.Get(ctx, e.ID)
}

// ---- Soft delete (cascading: soft-delete employee + deactivate user) ----

func (s *EmployeeService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	e, err := s.emps.FindByIDWithFull(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Employee")
		}
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Employee{}).
			Where("id = ?", e.ID).
			Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error; err != nil {
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

func (s *EmployeeService) uploadAvatar(ctx context.Context, employeeID uuid.UUID, prev *string, content []byte, contentType, ext string) (*dto.EmployeeRead, error) {
	if len(content) > maxAvatarBytes {
		return nil, apperrors.ErrBadRequest("Avatar must not exceed 5MB")
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, apperrors.ErrBadRequest("Avatar must be an image (PNG, JPEG, or WEBP)")
	}
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

