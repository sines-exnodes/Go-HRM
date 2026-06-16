package services

import (
	"context"
	"errors"
	"log"
	"strings"
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
	db           *gorm.DB
	users        repositories.UserRepository
	roles        repositories.RoleRepository
	employees    repositories.EmployeeRepository
	systemConfig repositories.SystemConfigRepository
	cfg          SeedConfig
}

// NewSeedService constructs a SeedService.
func NewSeedService(
	db *gorm.DB,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	employees repositories.EmployeeRepository,
	systemConfig repositories.SystemConfigRepository,
	cfg SeedConfig,
) *SeedService {
	return &SeedService{db: db, users: users, roles: roles, employees: employees, systemConfig: systemConfig, cfg: cfg}
}

type roleSeed struct {
	Name        string
	Description string
	Level       int
	Permissions []permissions.Permission
}

func defaultRoles() []roleSeed {
	return []roleSeed{
		{
			Name:        "Super Admin",
			Description: "Full system access with all permissions",
			Level:       100,
			Permissions: []permissions.Permission{permissions.PermAll},
		},
		{
			Name:        "Admin",
			Description: "Administrative access for user and role management",
			Level:       90,
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersDelete,
				permissions.PermUsersManageRoles, permissions.PermUsersChangePwd,
				// Salary/banking field-level perms (employees parity #6) — Admin
				// manages payroll, so holds both view and manage on each.
				permissions.PermUsersSalaryView, permissions.PermUsersSalaryManage,
				permissions.PermUsersBankingView, permissions.PermUsersBankingManage,
				permissions.PermUsersContractsView, permissions.PermUsersContractsManage,
				permissions.PermRolesRead, permissions.PermRolesCreate, permissions.PermRolesUpdate,
				// Employee HR profile management — added in the full-verify fix
				// (full-verify.md). Pre-existing seed gap: Admin could not GET /
				// employees, POST /employees, etc. because PermEmployees* was
				// only available to Super Admin's wildcard. Adding here matches
				// the role name's intent (Admin manages the HR aggregate too).
				permissions.PermEmployeesRead, permissions.PermEmployeesCreate, permissions.PermEmployeesUpdate, permissions.PermEmployeesDelete,
				permissions.PermDependentsManage,
				permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate, permissions.PermDepartmentsDelete,
				permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate, permissions.PermPositionsDelete,
				permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate, permissions.PermSkillsDelete,
				permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
				permissions.PermLeaveApproveAll, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermLeaveQuotaManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
				// Phase 4: announcement-label endpoints are gated by
				// PermAnnounceManage. Without this line, only Super Admin's
				// wildcard would reach the labels API — labels are admin-
				// managed, so Admin must hold the perm directly.
				permissions.PermAnnounceManage,
				permissions.PermOrgSettings,
				permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
				permissions.PermOrgWorkdaysView,
				permissions.PermInviteManage,
			},
		},
		{
			Name:        "HR Manager",
			Description: "Human resources management access",
			Level:       80,
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersChangePwd,
				// Salary/banking field-level perms (employees parity #6) — HR
				// manages payroll, so holds both view and manage on each.
				permissions.PermUsersSalaryView, permissions.PermUsersSalaryManage,
				permissions.PermUsersBankingView, permissions.PermUsersBankingManage,
				permissions.PermUsersContractsView, permissions.PermUsersContractsManage,
				permissions.PermRolesRead,
				// Employee HR profile management — same fix as Admin (above).
				// HR mirrors the rest of HR's perm shape: Read/Create/Update but
				// NOT Delete (Admin owns the destructive op on users/employees).
				permissions.PermEmployeesRead, permissions.PermEmployeesCreate, permissions.PermEmployeesUpdate,
				permissions.PermDependentsManage,
				permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate,
				permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate,
				permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate,
				permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
				permissions.PermLeaveApproveAll, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermLeaveQuotaManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
				// Phase 4: HR managers publish announcements in the Python
				// system and therefore also manage announcement labels.
				permissions.PermAnnounceManage,
				permissions.PermOrgSettings,
				permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
				permissions.PermOrgWorkdaysView,
				permissions.PermInviteManage,
			},
		},
		{
			Name:        "Manager",
			Description: "Team management access with user visibility",
			Level:       50,
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersRead, permissions.PermUsersContractsView,
				permissions.PermDepartmentsRead,
				permissions.PermPositionsRead,
				permissions.PermSkillsRead,
				permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
				permissions.PermLeaveApproveTeam, permissions.PermLeaveCancel, permissions.PermLeaveManage,
				permissions.PermAttendanceRead, permissions.PermAttendanceManage,
				permissions.PermOrgHolidaysView,
				permissions.PermOrgWorkdaysView,
			},
		},
		{
			Name:        "Employee",
			Description: "Basic employee access (own profile only)",
			Level:       10,
			Permissions: []permissions.Permission{
				permissions.PermAuthLogin,
				permissions.PermUsersContractsView,
				// Self-service on own leave requests: Read+Create were already
				// here; Update/Cancel/Delete are the load-bearing fix surfaced
				// by Phase 5 live verification (REVISION NOTES #4 had claimed
				// "no seed gap" but the Employee role was missing the perms an
				// owner needs to act on their own pending request). The
				// service enforces ownership in every branch, so granting
				// these to Employee cannot leak cross-employee writes.
				permissions.PermLeaveRead, permissions.PermLeaveCreate,
				permissions.PermLeaveUpdate, permissions.PermLeaveCancel,
				permissions.PermLeaveDelete,
				permissions.PermAttendanceRead,
				permissions.PermOrgHolidaysView,
				permissions.PermOrgWorkdaysView,
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
	if err := s.seedOrgDefaults(ctx); err != nil {
		return err
	}
	if err := s.seedSystemConfig(ctx); err != nil {
		return err
	}
	return nil
}

// seedSystemConfig is the Phase-8 idempotent INSERT of the system_config
// singleton row. Safe to call on every boot — uses ON CONFLICT DO NOTHING.
func (s *SeedService) seedSystemConfig(ctx context.Context) error {
	if s.systemConfig == nil {
		return nil
	}
	if err := s.systemConfig.EnsureExists(ctx); err != nil {
		return err
	}
	return nil
}

// seedOrgDefaults inserts a small default department/position tree the first
// time the departments table is empty. Idempotent: a non-empty table is left
// untouched so manual edits are never clobbered.
func (s *SeedService) seedOrgDefaults(ctx context.Context) error {
	var deptCount int64
	if err := s.db.WithContext(ctx).
		Model(&models.Department{}).
		Where("is_deleted = ?", false).
		Count(&deptCount).Error; err != nil {
		return err
	}
	if deptCount > 0 {
		return nil
	}

	eng := &models.Department{Name: "Engineering"}
	hr := &models.Department{Name: "Human Resources"}
	if err := s.db.WithContext(ctx).Create(eng).Error; err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Create(hr).Error; err != nil {
		return err
	}
	backend := &models.Department{Name: "Backend", ParentID: &eng.ID}
	mobile := &models.Department{Name: "Mobile", ParentID: &eng.ID}
	if err := s.db.WithContext(ctx).Create(backend).Error; err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Create(mobile).Error; err != nil {
		return err
	}

	// Positions are a flat global catalog post-000014 — no department FK.
	// Seeded names are picked to be unambiguous on their own since the
	// previously per-department uniqueness is now global.
	positions := []*models.Position{
		{Name: "Software Engineer"},
		{Name: "Mobile Engineer"},
		{Name: "HR Specialist"},
	}
	for _, p := range positions {
		if err := s.db.WithContext(ctx).Create(p).Error; err != nil {
			return err
		}
	}
	log.Printf("seed: created default org tree (4 departments, 3 positions)")
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
				Level:       rs.Level,
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

	adminFirst, adminLast := "Super", "Admin"
	if s.cfg.SuperAdminName != "" {
		parts := strings.SplitN(strings.TrimSpace(s.cfg.SuperAdminName), " ", 2)
		adminFirst = parts[0]
		if len(parts) > 1 {
			adminLast = strings.TrimSpace(parts[1])
		} else {
			adminLast = ""
		}
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
		FirstName:       adminFirst,
		LastName:        adminLast,
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
