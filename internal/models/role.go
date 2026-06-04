package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringSlice is a JSONB-backed []string for GORM.
type StringSlice []string

// Value implements driver.Valuer for GORM JSONB serialization.
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for GORM JSONB deserialization.
func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}
	var bytes []byte
	switch v := src.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("StringSlice: unsupported Scan source type")
	}
	if len(bytes) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Role maps to the roles table.
type Role struct {
	BaseModel
	Name        string      `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Description string      `gorm:"type:text;not null;default:''" json:"description"`
	Level       int         `gorm:"not null;default:100" json:"level"`
	IsSystem    bool        `gorm:"not null;default:false" json:"is_system"`
	Permissions StringSlice `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"permissions"`
}

func (Role) TableName() string { return "roles" }
