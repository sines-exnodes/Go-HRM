package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/exnodes/hrm-api/internal/models"
)

// UserListFilter is the querystring-derived filter for the admin user list.
type UserListFilter struct {
	Page     int
	PageSize int
	Search   string // substring match on email (ILIKE)
	IsActive *bool  // optional is_active filter
}

// UserRepository defines data access for users.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error)
	List(ctx context.Context, f UserListFilter) ([]models.User, int64, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error)
	FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error

	// Phase 2 auth-side admin queries.
	ExistsByEmail(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, newEmail string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashed string) error
	ToggleActive(ctx context.Context, id uuid.UUID, active bool) error
	AssignRoles(ctx context.Context, id uuid.UUID, roleIDs []uuid.UUID) error
	CreateTx(ctx context.Context, tx *gorm.DB, u *models.User) error

	// SetLoginAttempts persists the brute-force-protection counters.
	// A nil lockedUntil clears the lock; a non-nil value stamps it. Used
	// by AuthService.Login on bad password, threshold-hit, and successful
	// login (which calls it with attempts=0, lockedUntil=nil).
	SetLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error
}

type userRepository struct{ db *gorm.DB }

// NewUserRepository constructs a Postgres-backed UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// loadActiveRoles populates u.Roles with only the roles whose user_roles
// membership AND role row are both not soft-deleted. GORM's many2many preload
// fetches the join rows and the roles in separate queries, so the join-table
// soft-delete flag cannot be expressed as a preload condition; an explicit
// join query is used instead so a soft-deleted membership does not grant the
// role.
func loadActiveRoles(db *gorm.DB, ctx context.Context, u *models.User) error {
	var roles []models.Role
	err := db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ?", u.ID).
		Where("ur.is_deleted = ?", false).
		Where("roles.is_deleted = ?", false).
		Find(&roles).Error
	if err != nil {
		return err
	}
	u.Roles = roles
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	if err := loadActiveRoles(r.db, ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	if err := loadActiveRoles(r.db, ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Employee", notDeleted).
		First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	if err := loadActiveRoles(r.db, ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Employee", notDeleted).
		First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	if err := loadActiveRoles(r.db, ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// List returns a paginated slice of non-deleted users (ordered by created_at
// DESC) with active roles loaded, plus the total count. An optional email
// substring and is_active filter narrow the result.
func (r *userRepository) List(ctx context.Context, f UserListFilter) ([]models.User, int64, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 10
	}
	q := r.db.WithContext(ctx).Model(&models.User{}).Scopes(notDeleted)
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("email ILIKE ?", "%"+s+"%")
	}
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []models.User
	if err := q.
		Preload("Employee", notDeleted).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	for i := range users {
		if err := loadActiveRoles(r.db, ctx, &users[i]); err != nil {
			return nil, 0, err
		}
	}
	return users, total, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now, "is_active": false})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}
	return nil
}

// ReplaceRoles atomically replaces the user's role set. Memberships are
// soft-deleted (audit trail preserved) rather than physically removed; a
// previously soft-deleted (user_id, role_id) pair is revived in place to
// respect the (user_id, role_id) primary key.
func (r *userRepository) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceUserRolesTx(tx, userID, roleIDs)
	})
}

// replaceUserRolesTx soft-deletes the user's current memberships then revives
// or inserts the desired set within the supplied transaction.
func replaceUserRolesTx(tx *gorm.DB, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if err := tx.Exec(
		"UPDATE user_roles SET is_deleted = TRUE, deleted_at = NOW() WHERE user_id = ? AND is_deleted = FALSE",
		userID,
	).Error; err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	// Upsert on the (user_id, role_id) PK: a soft-deleted row is revived
	// (is_deleted=FALSE, deleted_at=NULL); a fresh pair is inserted.
	rows := make([]map[string]any, 0, len(roleIDs))
	for _, rid := range roleIDs {
		rows = append(rows, map[string]any{
			"user_id":    userID,
			"role_id":    rid,
			"created_at": gorm.Expr("NOW()"),
			"updated_at": gorm.Expr("NOW()"),
			"is_deleted": false,
			"deleted_at": gorm.Expr("NULL"),
		})
	}
	return tx.Table("user_roles").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"is_deleted": false,
				"deleted_at": nil,
				"updated_at": gorm.Expr("NOW()"),
			}),
		}).
		Create(&rows).Error
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ? AND is_deleted = ?", email, false)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, id uuid.UUID, newEmail string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{
			"email":            newEmail,
			"email_changed_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashed string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{
			"password_hash":     hashed,
			"password_reset_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *userRepository) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Update("is_active", active).Error
}

func (r *userRepository) AssignRoles(ctx context.Context, id uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceUserRolesTx(tx, id, roleIDs)
	})
}

// CreateTx is used by employee.Create within a transaction.
func (r *userRepository) CreateTx(ctx context.Context, tx *gorm.DB, u *models.User) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(u).Error
}

func (r *userRepository) SetLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{
			"failed_login_attempts": attempts,
			"locked_until":          lockedUntil,
		}).Error
}
