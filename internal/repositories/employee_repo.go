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
	// FindByIDWithOrg preloads User + Department + Position (for line-manager
	// candidate/brief rows that need the org context — employees parity #10).
	FindByIDWithOrg(ctx context.Context, id uuid.UUID) (*models.Employee, error)

	// Phase 2 admin queries.
	FindByIDWithFull(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	FindByUserIDWithFull(ctx context.Context, userID uuid.UUID) (*models.Employee, error)
	List(ctx context.Context, q dto.EmployeeListQuery) ([]models.Employee, int64, error)
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, url *string) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]any) error
	// ReplaceEmergencyContacts replaces the live emergency-contact set for an
	// employee: soft-deletes all current live rows, then inserts the new set
	// fresh (UUID PKs, so no reactivation is needed). Empty slice = clear all.
	ReplaceEmergencyContacts(ctx context.Context, employeeID uuid.UUID, contacts []models.EmployeeEmergencyContact) error

	// Line-manager suite (employees parity #10).
	// SubordinateIDs returns the transitive set of employees reporting (directly
	// or via chain) to rootEmployeeID. The root itself is NOT included.
	SubordinateIDs(ctx context.Context, rootEmployeeID uuid.UUID) (map[uuid.UUID]bool, error)
	// ListManagerCandidates returns active, non-deleted employees not in
	// excludeIDs, with User/Department/Position preloaded. Optional search
	// matches full_name / position name / department name.
	ListManagerCandidates(ctx context.Context, excludeIDs []uuid.UUID, search string, limit int) ([]models.Employee, error)
	// ListDirectReports returns live employees whose manager_id = managerID
	// (active AND inactive), with User/Department/Position preloaded.
	ListDirectReports(ctx context.Context, managerID uuid.UUID) ([]models.Employee, error)

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

func (r *employeeRepository) FindByIDWithOrg(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("User", notDeleted).
		Preload("Department", notDeleted).
		Preload("Position", notDeleted).
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
		Preload("Manager.User", notDeleted).
		Preload("Manager.Department", notDeleted).
		Preload("Manager.Position", notDeleted).
		Preload("Dependents", "is_deleted = ?", false).
		Preload("EmergencyContacts", "is_deleted = ?", false).
		Preload("EmployeeSkills", "is_deleted = ?", false).
		Preload("EmployeeSkills.Skill", "is_deleted = ?", false).
		Preload("LeaveQuota", "is_deleted = ?", false).
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
		Preload("Manager.User", notDeleted).
		Preload("Manager.Department", notDeleted).
		Preload("Manager.Position", notDeleted).
		Preload("Dependents", "is_deleted = ?", false).
		Preload("EmergencyContacts", "is_deleted = ?", false).
		Preload("EmployeeSkills", "is_deleted = ?", false).
		Preload("EmployeeSkills.Skill", "is_deleted = ?", false).
		Preload("LeaveQuota", "is_deleted = ?", false).
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
		Preload("Manager").
		Preload("Manager.User", notDeleted).
		Preload("Manager.Department", notDeleted).
		Preload("Manager.Position", notDeleted).
		Preload("EmergencyContacts", "is_deleted = ?", false).
		Preload("EmployeeSkills", "is_deleted = ?", false).
		Preload("EmployeeSkills.Skill", "is_deleted = ?", false).
		Preload("LeaveQuota", "is_deleted = ?", false).
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

func (r *employeeRepository) ReplaceEmergencyContacts(ctx context.Context, employeeID uuid.UUID, contacts []models.EmployeeEmergencyContact) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft-delete the current live set.
		if err := tx.Model(&models.EmployeeEmergencyContact{}).
			Where("employee_id = ? AND is_deleted = ?", employeeID, false).
			Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error; err != nil {
			return err
		}
		if len(contacts) == 0 {
			return nil
		}
		// Insert the new set fresh. Leave ID/CreatedAt/UpdatedAt zero so the
		// DB defaults (gen_random_uuid(), NOW()) apply.
		rows := make([]models.EmployeeEmergencyContact, 0, len(contacts))
		for _, c := range contacts {
			rows = append(rows, models.EmployeeEmergencyContact{
				EmployeeID:   employeeID,
				FullName:     c.FullName,
				Relationship: c.Relationship,
				PhoneNumber:  c.PhoneNumber,
			})
		}
		return tx.Create(&rows).Error
	})
}

func (r *employeeRepository) SubordinateIDs(ctx context.Context, rootEmployeeID uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool)
	frontier := []uuid.UUID{rootEmployeeID}
	for len(frontier) > 0 {
		var children []uuid.UUID
		if err := r.db.WithContext(ctx).Model(&models.Employee{}).
			Where("manager_id IN ? AND is_deleted = ?", frontier, false).
			Pluck("id", &children).Error; err != nil {
			return nil, err
		}
		next := make([]uuid.UUID, 0, len(children))
		for _, c := range children {
			// Guard against self-reference and re-visits (cycle-safe BFS).
			if c == rootEmployeeID || result[c] {
				continue
			}
			result[c] = true
			next = append(next, c)
		}
		frontier = next
	}
	return result, nil
}

func (r *employeeRepository) ListManagerCandidates(ctx context.Context, excludeIDs []uuid.UUID, search string, limit int) ([]models.Employee, error) {
	q := r.db.WithContext(ctx).Model(&models.Employee{}).
		Preload("User", notDeleted).
		Preload("Department", notDeleted).
		Preload("Position", notDeleted).
		Joins("JOIN users ON users.id = employees.user_id").
		Where("employees.is_deleted = ? AND users.is_active = ?", false, true)
	if len(excludeIDs) > 0 {
		q = q.Where("employees.id NOT IN ?", excludeIDs)
	}
	if search != "" {
		p := utils.BuildILIKEPattern(search)
		// LEFT JOIN only LIVE org rows so a soft-deleted position/department
		// name cannot drive a search match (matches the NotDeleted convention).
		q = q.Joins("LEFT JOIN positions ON positions.id = employees.position_id AND positions.is_deleted = false").
			Joins("LEFT JOIN departments ON departments.id = employees.department_id AND departments.is_deleted = false").
			Where("employees.full_name ILIKE ? OR positions.name ILIKE ? OR departments.name ILIKE ?", p, p, p)
	}
	if limit < 1 {
		limit = 50
	}
	var emps []models.Employee
	err := q.Order("LOWER(employees.full_name) ASC").Limit(limit).Find(&emps).Error
	return emps, err
}

func (r *employeeRepository) ListDirectReports(ctx context.Context, managerID uuid.UUID) ([]models.Employee, error) {
	var emps []models.Employee
	err := r.db.WithContext(ctx).
		Preload("User", notDeleted).
		Preload("Department", notDeleted).
		Preload("Position", notDeleted).
		Where("manager_id = ? AND is_deleted = ?", managerID, false).
		Order("LOWER(full_name) ASC").
		Find(&emps).Error
	return emps, err
}

// WithTx returns a repository bound to the given transaction handle.
func (r *employeeRepository) WithTx(tx *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: tx}
}

func (r *employeeRepository) DB() *gorm.DB { return r.db }
