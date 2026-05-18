package models

import "github.com/google/uuid"

// DeviceToken is a registered FCM/APNs token for push notifications.
type DeviceToken struct {
	BaseModel
	UserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token    string    `gorm:"type:text;not null" json:"token"`
	DeviceID string    `gorm:"type:text;not null" json:"device_id"`
	Platform string    `gorm:"type:text;not null;default:'unknown'" json:"platform"`
}

func (DeviceToken) TableName() string { return "device_tokens" }
