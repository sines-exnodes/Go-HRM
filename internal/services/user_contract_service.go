package services

import (
	"context"
	"errors"
	"math"
	"net/http"
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

// List returns a paginated, optionally filtered list of contracts for the user.
func (s *UserContractService) List(ctx context.Context, userID uuid.UUID, q dto.UserContractListQuery) (dto.PaginatedData[dto.UserContractRead], *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return dto.PaginatedData[dto.UserContractRead]{}, aerr
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 50 {
		q.PageSize = 10
	}
	contracts, total, err := s.repo.List(ctx, emp.ID, q)
	if err != nil {
		return dto.PaginatedData[dto.UserContractRead]{}, apperrors.ErrInternal(err.Error())
	}
	items := make([]dto.UserContractRead, len(contracts))
	for i, c := range contracts {
		items[i] = toUserContractRead(c)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(q.PageSize)))
	}
	return dto.PaginatedData[dto.UserContractRead]{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update applies a partial PATCH. Only non-nil DTO fields are applied.
func (s *UserContractService) Update(ctx context.Context, userID, contractID uuid.UUID, req dto.UserContractUpdate) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}
	if req.ContractType != nil {
		c.ContractType = models.ContractType(*req.ContractType)
	}
	if req.SignedDate != nil {
		c.SignedDate = *req.SignedDate
	}
	if req.IsEndless != nil {
		c.IsEndless = *req.IsEndless
	}
	if req.ExpiryDate != nil {
		c.ExpiryDate = req.ExpiryDate
	}
	if req.AttachmentURL != nil {
		if *req.AttachmentURL == "" {
			c.AttachmentURL = nil
		} else {
			c.AttachmentURL = req.AttachmentURL
		}
	}
	// If endless, always clear expiry.
	if c.IsEndless {
		c.ExpiryDate = nil
	}
	if aerr := validateContractDates(c.IsEndless, c.SignedDate, c.ExpiryDate); aerr != nil {
		return nil, aerr
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := toUserContractRead(*c)
	return &out, nil
}

// UploadAttachment validates, uploads, and stores the attachment URL on the contract.
// ext is the lowercase file extension (e.g. ".pdf").
func (s *UserContractService) UploadAttachment(ctx context.Context, userID, contractID uuid.UUID, content []byte, ext string) (*dto.UserContractAttachmentResponse, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}
	if s.uploads == nil {
		return nil, apperrors.ErrInternal("storage is not configured; cannot upload attachment")
	}
	if len(content) == 0 {
		return nil, apperrors.ErrBadRequest("attachment file is empty")
	}
	if len(content) > contractAttachmentMaxBytes {
		return nil, apperrors.ErrBadRequest("attachment must not exceed 5 MB")
	}
	sniffLen := len(content)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffed := http.DetectContentType(content[:sniffLen])
	if sniffed == "application/zip" && ext == ".docx" {
		sniffed = contractDocxMIME
	}
	if !allowedContractMIME[sniffed] {
		return nil, apperrors.ErrBadRequest("attachment must be PDF, PNG, JPG, or DOCX")
	}
	url, err := s.uploads.Upload(ctx, contractAttachmentSubdir, ext, content, sniffed)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	c.AttachmentURL = &url
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	return &dto.UserContractAttachmentResponse{AttachmentURL: url}, nil
}
