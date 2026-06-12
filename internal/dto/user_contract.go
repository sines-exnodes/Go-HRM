package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserContractCreate is the create request body.
type UserContractCreate struct {
	ContractType  string     `json:"contract_type"  binding:"required,oneof=labour_contract"`
	SignedDate    time.Time  `json:"signed_date"    binding:"required"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     bool       `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
}

// UserContractUpdate is the partial-PATCH request body.
// nil pointer = leave field unchanged.
// AttachmentURL non-nil empty string ("") = remove the attachment.
type UserContractUpdate struct {
	ContractType  *string    `json:"contract_type"  binding:"omitempty,oneof=labour_contract"`
	SignedDate    *time.Time `json:"signed_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     *bool      `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
}

// UserContractListQuery holds the list filter and pagination params.
type UserContractListQuery struct {
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
	SignedFrom *time.Time `form:"signed_from"`
	SignedTo   *time.Time `form:"signed_to"`
	ExpiryFrom *time.Time `form:"expiry_from"`
	ExpiryTo   *time.Time `form:"expiry_to"`
}

// UserContractRead is the API response shape for a single contract.
type UserContractRead struct {
	ID            uuid.UUID  `json:"id"`
	ContractType  string     `json:"contract_type"`
	SignedDate    time.Time  `json:"signed_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     bool       `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UserContractAttachmentResponse is returned after a successful attachment upload.
type UserContractAttachmentResponse struct {
	AttachmentURL string `json:"attachment_url"`
}
