package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// SeedConfig configures the seed service.
type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string // default "Super Admin" if blank
}

// SeedService creates the 5 system roles, 1 super admin user, and the
// matching employee row on boot. Safe to run repeatedly — operations are
// merge/upsert and never overwrite manually-edited records.
type SeedService struct {
	db        *gorm.DB
	users     repositories.UserRepository
	roles     repositories.RoleRepository
	employees repositories.EmployeeRepository
	cfg       SeedConfig
}

// NewSeedService constructs a SeedService.
func NewSeedService(
	db *gorm.DB,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	employees repositories.EmployeeRepository,
	cfg SeedConfig,
) *SeedService {
	return &SeedService{db: db, users: users, roles: roles, employees: employees, cfg: cfg}
}

type roleSeed struct {
	Name        string
	Description string
	Permissions []permissions.Permission
}

func defaultRoles() []roleSeed {
	return []roleSeed{
		{
			Name:        "Super Admin",
			Description: "Full system access with all permissions",
			Permissions: []permissions.Permission{permissions.PermAll},
		},
		{
			Name:        "Admin",
			Description: "Administrative access for user and role management",
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersDelete,
				permissions.PermUsersManageRoles, permissions.PermUsersChangePwd,
				permissions.PermRolesRead, permissions.PermRolesCreate, permissions.PermRolesUpdate,
				permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate, permissions.PermDepartmentsDelete,
				permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate, permissions.PermPositionsDelete,
				permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate, permissions.PermSkillsDelete,
				permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
				permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermLeaveQuotaManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
				permissions.PermOrgSettings,
			},
		},
		{
			Name:        "HR Manager",
			Description: "Human resources management access",
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersChangePwd,
				permissions.PermRolesRead,
				permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate,
				permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate,
				permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate,
				permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
				permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermLeaveQuotaManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
				permissions.PermOrgSettings,
			},
		},
		{
			Name:        "Manager",
			Description: "Team management access with user visibility",
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead,
				permissions.PermDepartmentsRead,
				permissions.PermPositionsRead,
				permissions.PermSkillsRead,
				permissions.PermLeaveRead, permissions.PermLeaveCreate,
				permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
			},
		},
		{
			Name:        "Employee",
			Description: "Basic employee access (own profile only)",
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermLeaveRead, permissions.PermLeaveCreate,
				permissions.PermAttendanceRead,
			},
		},
	}
}

// Seed creates/updates the 5 system roles and the configured super admin.
func (s *SeedService) Seed(ctx context.Context) error {
	if err := s.seedRoles(ctx); err != nil {
		return err
	}
	if err := s.seedSuperAdmin(ctx); err != nil {
		return err
	}
	return nil
}

func (s *SeedService) seedRoles(ctx context.Context) error {
	for _, rs := range defaultRoles() {
		existing, err := s.roles.FindByName(ctx, rs.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		desired := make(models.StringSlice, 0, len(rs.Permissions))
		for _, p := range rs.Permissions {
			desired = append(desired, string(p))
		}
		if existing == nil {
			r := &models.Role{
				Name:        rs.Name,
				Description: rs.Description,
				IsSystem:    true,
				Permissions: desired,
			}
			if err := s.roles.Create(ctx, r); err != nil {
				return err
			}
			log.Printf("seed: created role %q", rs.Name)
			continue
		}
		// Merge: only add missing perms, never remove manually-added ones.
		if !existing.IsSystem {
			existing.IsSystem = true
		}
		present := map[string]bool{}
		for _, p := range existing.Permissions {
			present[p] = true
		}
		changed := false
		for _, p := range desired {
			if !present[p] {
				existing.Permissions = append(existing.Permissions, p)
				changed = true
			}
		}
		if changed {
			if err := s.roles.Update(ctx, existing); err != nil {
				return err
			}
			log.Printf("seed: merged permissions into role %q", rs.Name)
		}
	}
	return nil
}

func (s *SeedService) seedSuperAdmin(ctx context.Context) error {
	if s.cfg.SuperAdminEmail == "" || s.cfg.SuperAdminPassword == "" {
		log.Printf("seed: SUPER_ADMIN_EMAIL/PASSWORD not set, skipping super admin user")
		return nil
	}
	saRole, err := s.roles.FindByName(ctx, "Super Admin")
	if err != nil {
		return err
	}

	adminName := s.cfg.SuperAdminName
	if adminName == "" {
		adminName = "Super Admin"
	}

	existing, err := s.users.FindByEmailWithRolesAndEmployee(ctx, s.cfg.SuperAdminEmail)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var userID uuid.UUID
	if existing != nil {
		userID = existing.ID

		// Ensure role linkage; never touch password.
		ids := []uuid.UUID{}
		hasSA := false
		for _, r := range existing.Roles {
			ids = append(ids, r.ID)
			if r.ID == saRole.ID {
				hasSA = true
			}
		}
		if !hasSA {
			ids = append(ids, saRole.ID)
			if err := s.users.ReplaceRoles(ctx, existing.ID, ids); err != nil {
				return err
			}
			log.Printf("seed: linked super admin role to existing user %q", existing.Email)
		}
	} else {
		hash, err := utils.HashPassword(s.cfg.SuperAdminPassword)
		if err != nil {
			return err
		}
		u := &models.User{
			Email:        s.cfg.SuperAdminEmail,
			PasswordHash: hash,
			IsActive:     true,
		}
		if err := s.users.Create(ctx, u); err != nil {
			return err
		}
		if err := s.users.ReplaceRoles(ctx, u.ID, []uuid.UUID{saRole.ID}); err != nil {
			return err
		}
		log.Printf("seed: created super admin user %q", u.Email)
		userID = u.ID
	}

	// Ensure the matching employee row exists (idempotent — never overwrite
	// a manually-edited employee record).
	_, err = s.employees.FindByUserID(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	today := time.Now().UTC()
	emp := &models.Employee{
		UserID:          userID,
		FullName:        adminName,
		ContractType:    "official",
		ContractRenewal: 1,
		PaymentMethod:   "bank_transfer",
		JoinDate:        &today,
	}
	if err := s.employees.Create(ctx, emp); err != nil {
		return err
	}
	log.Printf("seed: created super admin employee profile for %q", s.cfg.SuperAdminEmail)
	return nil
}
