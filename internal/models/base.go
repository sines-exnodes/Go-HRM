package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded in every entity model. The 4 audit columns are
// required by the migration design (see spec §5.2). Soft delete is NOT
// implemented via gorm.DeletedAt — instead, the explicit IsDeleted boolean
// and DeletedAt timestamp are managed by service-level Delete/Restore calls
// and queried through the NotDeleted scope.
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time  `gorm:"not null;default:now()"                          json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()"                          json:"updated_at"`
	IsDeleted bool       `gorm:"not null;default:false;index"                    json:"-"`
	DeletedAt *time.Time `                                                       json:"-"`
}

// NotDeleted is the default scope applied by repositories to every list,
// get-by-id, and count query. Callers that intentionally need to read
// soft-deleted rows (e.g. an admin "restore" flow) must opt out by NOT
// chaining this scope.
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}
