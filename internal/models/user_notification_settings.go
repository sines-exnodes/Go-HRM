package models

import "github.com/google/uuid"

// UserNotificationSettings is a 1-1 per-user push notification toggle.
type UserNotificationSettings struct {
	BaseModel
	UserID               uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	NotificationsEnabled bool      `gorm:"not null;default:true" json:"notifications_enabled"`
}

func (UserNotificationSettings) TableName() string { return "user_notification_settings" }
