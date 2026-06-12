package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

const (
	contractAttachmentSubdir   = "contracts"
	contractAttachmentMaxBytes = 5 * 1024 * 1024
	contractDocxMIME           = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

var allowedContractMIME = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	contractDocxMIME:  true,
}

// UserContractService manages employment contracts per user.
type UserContractService struct {
	repo         repositories.UserContractRepository
	employeeRepo repositories.EmployeeRepository
	uploads      Uploader
}

// NewUserContractService constructs a UserContractService.
// Pass nil for uploads to disable attachment upload.
func NewUserContractService(
	repo repositories.UserContractRepository,
	employeeRepo repositories.EmployeeRepository,
	uploads Uploader,
) *UserContractService {
	return &UserContractService{repo: repo, employeeRepo: employeeRepo, uploads: uploads}
}

// ---- private helpers ----

func (s *UserContractService) resolveEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, *apperrors.AppError) {
	emp, err := s.employeeRepo.FindByUserID(ctx, userID)
	if err != nil || emp == nil {
		return nil, apperrors.ErrNotFound("employee profile not found")
	}
	return emp, nil
}

func (s *UserContractService) fetchAndCheckOwnership(ctx context.Context, contractID, employeeID uuid.UUID) (*models.UserContract, *apperrors.AppError) {
	c, err := s.repo.Get(ctx, contractID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("contract not found")
		}
		return nil, apperrors.ErrInternal(err.Error())
	}
	if c.EmployeeID != employeeID {
		return nil, apperrors.ErrNotFound("contract not found")
	}
	return c, nil
}

func validateContractDates(isEndless bool, signedDate time.Time, expiryDate *time.Time) *apperrors.AppError {
	if isEndless {
		return nil
	}
	if expiryDate == nil {
		return apperrors.ErrBadRequest("expiry date is required")
	}
	if !expiryDate.After(signedDate) {
		return apperrors.ErrBadRequest("expiry date must be after signed date")
	}
	return nil
}

func toUserContractRead(c models.UserContract) dto.UserContractRead {
	return dto.UserContractRead{
		ID:            c.ID,
		ContractType:  string(c.ContractType),
		SignedDate:    c.SignedDate,
		ExpiryDate:    c.ExpiryDate,
		IsEndless:     c.IsEndless,
		AttachmentURL: c.AttachmentURL,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ---- public methods ----

// Create adds a new contract for the user identified by userID.
func (s *UserContractService) Create(ctx context.Context, userID uuid.UUID, req dto.UserContractCreate) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	expiry := req.ExpiryDate
	if req.IsEndless {
		expiry = nil
	}
	if aerr := validateContractDates(req.IsEndless, req.SignedDate, expiry); aerr != nil {
		return nil, aerr
	}
	c := &models.UserContract{
		EmployeeID:    emp.ID,
		ContractType:  models.ContractType(req.ContractType),
		SignedDate:    req.SignedDate,
		ExpiryDate:    expiry,
		IsEndless:     req.IsEndless,
		AttachmentURL: req.AttachmentURL,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := toUserContractRead(*c)
	return &out, nil
}

// Get returns a single contract. Returns 404 if not found or belongs to a different user.
func (s *UserContractService) Get(ctx context.Context, userID, contractID uuid.UUID) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}
	out := toUserContractRead(*c)
	return &out, nil
}

// Delete soft-deletes a contract.
func (s *UserContractService) Delete(ctx context.Context, userID, contractID uuid.UUID) *apperrors.AppError {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return aerr
	}
	if _, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID); aerr != nil {
		return aerr
	}
	if err := s.repo.Delete(ctx, contractID); err != nil {
		return apperrors.ErrInternal(err.Error())
	}
	return nil
}

// List — stubbed, implemented in a follow-up task.
func (s *UserContractService) List(_ context.Context, _ uuid.UUID, _ dto.UserContractListQuery) (dto.PaginatedData[dto.UserContractRead], *apperrors.AppError) {
	return dto.PaginatedData[dto.UserContractRead]{}, nil
}

// Update — stubbed, implemented in a follow-up task.
func (s *UserContractService) Update(_ context.Context, _, _ uuid.UUID, _ dto.UserContractUpdate) (*dto.UserContractRead, *apperrors.AppError) {
	return nil, nil
}

// UploadAttachment — stubbed, implemented in a follow-up task.
func (s *UserContractService) UploadAttachment(_ context.Context, _, _ uuid.UUID, _ []byte, _ string) (*dto.UserContractAttachmentResponse, *apperrors.AppError) {
	return nil, nil
}
