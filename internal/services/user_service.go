package services

import (
	"context"
	"errors"
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
	emps     repositories.EmployeeRepository
	tokens   *repositories.DeviceTokenRepository
	settings *repositories.NotificationSettingsRepository
	empSvc   *EmployeeService
}

func NewUserService(
	users repositories.UserRepository,
	emps repositories.EmployeeRepository,
	tokens *repositories.DeviceTokenRepository,
	settings *repositories.NotificationSettingsRepository,
	empSvc *EmployeeService,
) *UserService {
	return &UserService{users: users, emps: emps, tokens: tokens, settings: settings, empSvc: empSvc}
}

// ---- GET /users/me ----

func (s *UserService) GetMe(ctx context.Context, u *models.User) (*dto.UserMeRead, error) {
	roles := make([]dto.RoleRead, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, dto.RoleRead{
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
	return s.users.AssignRoles(ctx, id, roleIDs)
}

// ---- Admin user patch (toggle is_active) ----

func (s *UserService) AdminPatch(ctx context.Context, id uuid.UUID, in dto.AdminUserPatch) error {
	if in.IsActive == nil {
		return nil
	}
	return s.users.ToggleActive(ctx, id, *in.IsActive)
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
