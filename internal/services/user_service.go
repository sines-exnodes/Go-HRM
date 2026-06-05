package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// UserService owns auth-side user operations. users/emps are reconciled to
// their repository INTERFACE types; tokens/settings are concrete structs.
type UserService struct {
	users    repositories.UserRepository
	roles    repositories.RoleRepository
	emps     repositories.EmployeeRepository
	tokens   *repositories.DeviceTokenRepository
	settings *repositories.NotificationSettingsRepository
	empSvc   *EmployeeService
}

func NewUserService(
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	emps repositories.EmployeeRepository,
	tokens *repositories.DeviceTokenRepository,
	settings *repositories.NotificationSettingsRepository,
	empSvc *EmployeeService,
) *UserService {
	return &UserService{users: users, roles: roles, emps: emps, tokens: tokens, settings: settings, empSvc: empSvc}
}

// ---- GET /users/me ----

func (s *UserService) GetMe(ctx context.Context, u *models.User) (*dto.UserMeRead, error) {
	roles := make([]dto.RoleRef, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, dto.RoleRef{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
			Permissions: []string(r.Permissions),
		})
	}
	notif := true
	if s.settings != nil {
		ns, err := s.settings.Get(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		if ns != nil {
			notif = ns.NotificationsEnabled
		}
	}
	out := &dto.UserMeRead{
		ID:                   u.ID,
		Email:                u.Email,
		IsActive:             u.IsActive,
		Roles:                roles,
		NotificationsEnabled: notif,
		CreatedAt:            u.CreatedAt,
		UpdatedAt:            u.UpdatedAt,
	}
	// Embed employee summary if present.
	e, err := s.emps.FindByUserIDWithFull(ctx, u.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if e != nil {
		out.Employee = s.empSvc.toSummary(e)
	}
	return out, nil
}

// ---- Admin list / get ----

// toAdminRead maps a user (with Roles preloaded and Employee optionally
// preloaded) to the admin read shape, embedding the employee summary.
func (s *UserService) toAdminRead(u *models.User) *dto.UserAdminRead {
	roles := make([]dto.RoleRef, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, dto.RoleRef{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
			Permissions: []string(r.Permissions),
		})
	}
	out := &dto.UserAdminRead{
		ID:        u.ID,
		Email:     u.Email,
		IsActive:  u.IsActive,
		Roles:     roles,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.Employee != nil {
		out.Employee = s.empSvc.toSummary(u.Employee)
	}
	return out
}

// List returns a paginated slice of users with roles + employee summary.
func (s *UserService) List(ctx context.Context, q dto.UserListQuery) (*dto.PaginatedData[dto.UserAdminRead], error) {
	users, total, err := s.users.List(ctx, repositories.UserListFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
		IsActive: q.IsActive,
	})
	if err != nil {
		return nil, err
	}
	reads := make([]dto.UserAdminRead, 0, len(users))
	for i := range users {
		reads = append(reads, *s.toAdminRead(&users[i]))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return &dto.PaginatedData[dto.UserAdminRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}

// Get returns a single user (roles + employee summary) by ID.
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*dto.UserAdminRead, error) {
	u, err := s.users.FindByIDWithRolesAndEmployee(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("User")
		}
		return nil, err
	}
	return s.toAdminRead(u), nil
}

// ---- Change password (self) ----

func (s *UserService) ChangePassword(ctx context.Context, user *models.User, current, next string) error {
	if !utils.CheckPassword(current, user.PasswordHash) {
		return apperrors.ErrBadRequest("Current password is incorrect")
	}
	hash, err := utils.HashPassword(next)
	if err != nil {
		return apperrors.ErrBadRequest("failed to hash password")
	}
	return s.users.UpdatePassword(ctx, user.ID, hash)
}

// ---- Admin change password ----

func (s *UserService) AdminChangePassword(ctx context.Context, id uuid.UUID, next string) error {
	hash, err := utils.HashPassword(next)
	if err != nil {
		return apperrors.ErrBadRequest("failed to hash password")
	}
	return s.users.UpdatePassword(ctx, id, hash)
}

// ---- Change email (self) ----

func (s *UserService) ChangeEmail(ctx context.Context, user *models.User, newEmail, currentPassword string) (*dto.UserMeRead, error) {
	if !utils.CheckPassword(currentPassword, user.PasswordHash) {
		return nil, apperrors.ErrBadRequest("Current password is incorrect")
	}
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == user.Email {
		return nil, apperrors.ErrBadRequest("New email must be different from the current email")
	}
	exists, err := s.users.ExistsByEmail(ctx, newEmail, &user.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict("This email is already in use")
	}
	if err := s.users.UpdateEmail(ctx, user.ID, newEmail); err != nil {
		return nil, err
	}
	user.Email = newEmail
	return s.GetMe(ctx, user)
}

// ---- Admin role assignment ----

func (s *UserService) AssignRoles(ctx context.Context, id uuid.UUID, roleIDs []uuid.UUID, admin *models.User) error {
	if admin.ID == id {
		return apperrors.ErrBadRequest("You cannot change your own role")
	}
	if err := s.checkRoleAssignmentAuthority(ctx, admin, roleIDs); err != nil {
		return err
	}
	return s.users.AssignRoles(ctx, id, roleIDs)
}

// checkRoleAssignmentAuthority ports the Python rule (app/services/role.py):
// an assigner may only grant roles whose level is <= the assigner's own max
// role level. admin.Roles is preloaded by the JWT middleware on the request
// path; tests must pass a *models.User whose Roles slice is populated.
func (s *UserService) checkRoleAssignmentAuthority(ctx context.Context, admin *models.User, roleIDs []uuid.UUID) error {
	if len(roleIDs) == 0 {
		return nil
	}
	assignerMax := 0
	for _, r := range admin.Roles {
		if r.Level > assignerMax {
			assignerMax = r.Level
		}
	}
	targetRoles, err := s.roles.FindByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	for _, r := range targetRoles {
		if r.Level > assignerMax {
			return apperrors.ErrForbidden(fmt.Sprintf(
				"Cannot assign role '%s' (level %d): exceeds your authority level (%d)", r.Name, r.Level, assignerMax))
		}
	}
	return nil
}

// ---- Admin user patch (toggle is_active) ----

func (s *UserService) AdminPatch(ctx context.Context, id uuid.UUID, in dto.AdminUserPatch, admin *models.User) error {
	if in.IsActive == nil {
		return nil
	}
	// Cannot deactivate your own account (employees parity #12).
	if !*in.IsActive && admin.ID == id {
		return apperrors.ErrBadRequest("You cannot deactivate your own account")
	}
	return s.users.ToggleActive(ctx, id, *in.IsActive)
}

// ---- Admin change email (employees parity #13) ----

// AdminChangeEmail changes another user's email. No password required — the
// route is gated by users:update. UpdateEmail stamps email_changed_at, which
// invalidates the target's existing access tokens on their next request.
func (s *UserService) AdminChangeEmail(ctx context.Context, id uuid.UUID, newEmail string) (*dto.UserAdminRead, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("User")
		}
		return nil, err
	}
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == u.Email {
		return nil, apperrors.ErrBadRequest("New email must be different from the current email")
	}
	exists, err := s.users.ExistsByEmail(ctx, newEmail, &u.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict("This email is already in use")
	}
	if err := s.users.UpdateEmail(ctx, u.ID, newEmail); err != nil {
		return nil, err
	}
	return s.Get(ctx, u.ID)
}

// ---- Admin delete (with reauth) ----

func (s *UserService) AdminDelete(ctx context.Context, id uuid.UUID, admin *models.User, currentPassword string) error {
	if admin.ID == id {
		return apperrors.ErrBadRequest("You cannot delete your own account")
	}
	if !utils.CheckPassword(currentPassword, admin.PasswordHash) {
		return apperrors.ErrBadRequest("Incorrect password. Please try again.")
	}
	return s.users.SoftDelete(ctx, id)
}

// ---- Device tokens ----

func (s *UserService) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, in dto.FcmTokenRequest) error {
	platform := in.Platform
	if platform == "" {
		platform = "unknown"
	}
	return s.tokens.Upsert(ctx, &models.DeviceToken{
		UserID:   userID,
		Token:    in.Token,
		DeviceID: in.DeviceID,
		Platform: platform,
	})
}

func (s *UserService) RemoveDeviceToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.tokens.DeleteByToken(ctx, userID, token)
}

// ---- Notification settings ----

func (s *UserService) UpdateNotificationSettings(ctx context.Context, userID uuid.UUID, enabled bool) error {
	return s.settings.Upsert(ctx, userID, enabled)
}
