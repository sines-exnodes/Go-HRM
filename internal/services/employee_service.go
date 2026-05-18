package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

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
	uploads *UploadService
}

func NewEmployeeService(
	db *gorm.DB,
	emps repositories.EmployeeRepository,
	deps *repositories.DependentRepository,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	quota *repositories.LeaveQuotaRepository,
	uploads *UploadService,
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
