package models

import (
	"time"

	"github.com/google/uuid"
)

// ContractType is the employment contract category. Only labour_contract in v1.
type ContractType string

const ContractTypeLabour ContractType = "labour_contract"

// UserContract maps to the user_contracts table.
// ExpiryDate is nil when IsEndless is true — never store a sentinel far-future date.
type UserContract struct {
	BaseModel
	EmployeeID    uuid.UUID    `gorm:"type:uuid;not null;index"`
	ContractType  ContractType `gorm:"type:text;not null"`
	SignedDate    time.Time    `gorm:"type:date;not null"`
	ExpiryDate    *time.Time   `gorm:"type:date"`
	IsEndless     bool         `gorm:"not null;default:false"`
	AttachmentURL *string      `gorm:"type:text"`
}

func (UserContract) TableName() string { return "user_contracts" }
