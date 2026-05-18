package dto

import (
	"time"

	"github.com/google/uuid"
)

type DependentRead struct {
	ID           uuid.UUID  `json:"id"`
	EmployeeID   uuid.UUID  `json:"employee_id"`
	FullName     string     `json:"full_name"`
	DOB          *time.Time `json:"dob,omitempty"`
	Gender       *string    `json:"gender,omitempty"`
	Relationship string     `json:"relationship"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DependentCreate struct {
	FullName     string     `json:"full_name"    binding:"required,min=1,max=200"`
	DOB          *time.Time `json:"dob,omitempty"`
	Gender       *string    `json:"gender,omitempty"       binding:"omitempty,oneof=male female other"`
	Relationship string     `json:"relationship" binding:"required,oneof=child parent spouse sibling other"`
}

type DependentUpdate struct {
	FullName     *string    `json:"full_name,omitempty"`
	DOB          *time.Time `json:"dob,omitempty"`
	Gender       *string    `json:"gender,omitempty"       binding:"omitempty,oneof=male female other"`
	Relationship *string    `json:"relationship,omitempty" binding:"omitempty,oneof=child parent spouse sibling other"`
}

type DependentListQuery struct {
	Page     int `form:"page,default=1"       binding:"gte=1"`
	PageSize int `form:"page_size,default=50" binding:"gte=1,lte=200"`
}
