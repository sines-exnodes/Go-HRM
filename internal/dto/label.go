package dto

import (
	"time"

	"github.com/google/uuid"
)

// LabelCreate is the body for POST /api/v1/announcement-labels. The
// endpoint has get-or-create semantics: if a live label with the same
// case-insensitive name exists, it is returned with HTTP 200; otherwise
// a new row is inserted and returned with HTTP 201.
type LabelCreate struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// LabelRead is the wire shape returned by every label endpoint.
type LabelRead struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
