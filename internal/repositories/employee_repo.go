package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// EmployeeRepository defines data access for the HR profile of a user.
type EmployeeRepository interface {
	Create(ctx context.Context, e *models.Employee) error
	Update(ctx context.Context, e *models.Employee) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error)
	FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error)

	// Phase 2 admin queries.
	FindByIDWithFull(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	FindByUserIDWithFull(ctx context.Context, userID uuid.UUID) (*models.Employee, error)
	List(ctx context.Context, q dto.EmployeeListQuery) ([]models.Employee, int64, error)
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, url *string) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]any) error
	WithTx(tx *gorm.DB) EmployeeRepository
	DB() *gorm.DB
}

type employeeRepository struct{ db *gorm.DB }

// NewEmployeeRepository constructs a Postgres-backed EmployeeRepository.
func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, e *models.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *employeeRepository) Update(ctx context.Context, e *models.Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *employeeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("employee not found or already deleted")
	}
	return nil
}

func (r *employeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employeeRepository) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("User", notDeleted).
		First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FindByIDWithFull preloads user, roles, manager, dependents.
// Department/Position preloads are deferred to Phase 3 (those tables land
// later), so they are intentionally not preloaded here.
func (r *employeeRepository) FindByIDWithFull(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Roles").
		Preload("Manager").
		Preload("Dependents", "is_deleted = ?", false).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &e, err
}

func (r *employeeRepository) FindByUserIDWithFull(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Roles").
		Preload("Manager").
		Preload("Dependents", "is_deleted = ?", false).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &e, err
}

func (r *employeeRepository) List(ctx context.Context, q dto.EmployeeListQuery) ([]models.Employee, int64, error) {
	tx := r.db.WithContext(ctx).Model(&models.Employee{}).
		Preload("User").
		Preload("User.Roles").
		Joins("JOIN users ON users.id = employees.user_id").
		Where("employees.is_deleted = ?", false)

	if q.Search != "" {
		p := utils.BuildILIKEPattern(q.Search)
		tx = tx.Where(
			"employees.full_name ILIKE ? OR employees.phone ILIKE ? OR employees.personal_email ILIKE ? OR users.email ILIKE ?",
			p, p, p, p,
		)
	}
	if q.DepartmentID != nil {
		tx = tx.Where("employees.department_id = ?", *q.DepartmentID)
	}
	if q.PositionID != nil {
		tx = tx.Where("employees.position_id = ?", *q.PositionID)
	}
	if q.ManagerID != nil {
		tx = tx.Where("employees.manager_id = ?", *q.ManagerID)
	}
	if q.IsActive != nil {
		tx = tx.Where("users.is_active = ?", *q.IsActive)
	}
	if q.RoleID != nil {
		tx = tx.Where(
			"EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = employees.user_id AND ur.role_id = ? AND ur.is_deleted = false)",
			*q.RoleID,
		)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}

	var emps []models.Employee
	if err := tx.
		Order("employees.full_name ASC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&emps).Error; err != nil {
		return nil, 0, err
	}
	return emps, total, nil
}

func (r *employeeRepository) UpdateAvatarURL(ctx context.Context, id uuid.UUID, url *string) error {
	return r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Update("avatar_url", url).Error
}

func (r *employeeRepository) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(fields).Error
}

// WithTx returns a repository bound to the given transaction handle.
func (r *employeeRepository) WithTx(tx *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: tx}
}

func (r *employeeRepository) DB() *gorm.DB { return r.db }
