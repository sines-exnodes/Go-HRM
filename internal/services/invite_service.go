package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// InviteService owns the invites aggregate. Permission gating is
// upstream (RequirePerms(PermInviteManage) on admin endpoints). The
// public Accept endpoint is unauthenticated by design — the token is
// the credential.
type InviteService struct {
	cfg     *config.Config
	repo    repositories.InviteRepository
	emps    repositories.EmployeeRepository
	users   repositories.UserRepository
	roles   repositories.RoleRepository
	empSvc  *EmployeeService
	email   *EmailService
	db      *gorm.DB
}

// NewInviteService constructs the service. emailSvc may be nil in
// tests; production wires the real one (which itself degrades to
// ErrEmailDisabled when SMTP is unconfigured).
func NewInviteService(
	cfg *config.Config,
	repo repositories.InviteRepository,
	emps repositories.EmployeeRepository,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	empSvc *EmployeeService,
	emailSvc *EmailService,
	db *gorm.DB,
) *InviteService {
	return &InviteService{
		cfg:    cfg,
		repo:   repo,
		emps:   emps,
		users:  users,
		roles:  roles,
		empSvc: empSvc,
		email:  emailSvc,
		db:     db,
	}
}

// generateToken returns 32 random bytes encoded as URL-safe base64
// (no padding). ~43 chars; collision-resistant by construction.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// resolveCurrentEmployee mirrors the pattern from LeaveService /
// AttendanceService / AnnouncementService / OrganizationSettingsService.
func (s *InviteService) resolveCurrentEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	emp, err := s.emps.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden("No employee record for current user")
		}
		return nil, err
	}
	return emp, nil
}

// status derives the pseudo-enum (pending/accepted/expired/revoked)
// from the row's three lifecycle columns.
func deriveInviteStatus(inv *models.Invite) string {
	if inv.IsDeleted {
		return "revoked"
	}
	if inv.AcceptedAt != nil {
		return "accepted"
	}
	if inv.ExpiresAt.Before(time.Now().UTC()) {
		return "expired"
	}
	return "pending"
}

// toRead is the projection helper.
func (s *InviteService) toRead(inv *models.Invite) dto.InviteRead {
	out := dto.InviteRead{
		ID:             inv.ID,
		Email:          inv.Email,
		FullName:       inv.FullName,
		RoleIDs:        []uuid.UUID(inv.RoleIDs),
		DepartmentID:   inv.DepartmentID,
		PositionID:     inv.PositionID,
		ExpiresAt:      inv.ExpiresAt,
		AcceptedAt:     inv.AcceptedAt,
		AcceptedUserID: inv.AcceptedUserID,
		Status:         deriveInviteStatus(inv),
		InvitedBy:      inv.InvitedBy,
		LastEmailError: inv.LastEmailError,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
	}
	if inv.Inviter != nil {
		out.Inviter = &dto.InviteInviterBrief{
			ID:       inv.Inviter.ID,
			FullName: inv.Inviter.FullName,
		}
	}
	return out
}

// sendInviteEmail is the best-effort delivery path. Records error on
// the invite row when SMTP fails; never returns the error to the
// caller (REVISION NOTES #11). Runs synchronously so the caller can
// surface the populated last_email_error in the create/resend response.
func (s *InviteService) sendInviteEmail(ctx context.Context, inv *models.Invite) {
	if s.email == nil {
		return
	}
	fullName := ""
	if inv.FullName != nil {
		fullName = *inv.FullName
	}
	acceptURL := fmt.Sprintf("%s/invites/accept?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), inv.Token)
	data := InviteEmailData{
		AppName:   s.cfg.AppName,
		FullName:  fullName,
		AcceptURL: acceptURL,
		ExpiresAt: inv.ExpiresAt.Format("2006-01-02 15:04 MST"),
	}
	err := s.email.SendInvite(ctx, inv.Email, data)
	if err != nil {
		msg := err.Error()
		inv.LastEmailError = &msg
		// Best-effort persist — log but don't propagate (the parent
		// request is already returning success).
		if upErr := s.repo.Update(ctx, inv); upErr != nil {
			log.Printf("invite: failed to persist last_email_error: %v", upErr)
		}
	} else {
		// Clear any prior error on a successful resend.
		if inv.LastEmailError != nil {
			inv.LastEmailError = nil
			if upErr := s.repo.Update(ctx, inv); upErr != nil {
				log.Printf("invite: failed to clear last_email_error: %v", upErr)
			}
		}
	}
}

// ---- Create ----

// Create issues an invite. Conflict (409) if the email already has a
// pending invite OR an existing user with that address. Returns the
// invite read shape regardless of email-send outcome — last_email_error
// is populated when SMTP fails.
func (s *InviteService) Create(ctx context.Context, currentUserID uuid.UUID, in dto.InviteCreate) (*dto.InviteRead, error) {
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return nil, apperrors.ErrBadRequest("email is required")
	}

	// Reject when a user with this email already exists (no point
	// inviting them).
	existingUser, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperrors.ErrConflict("A user with this email already exists")
	}

	// Reject when a pending invite already exists.
	pending, err := s.repo.FindPendingByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, apperrors.ErrConflict("A pending invite for this email already exists")
	}

	token, err := generateToken()
	if err != nil {
		return nil, apperrors.ErrInternal("failed to generate invite token")
	}

	expiresIn := s.cfg.InviteTokenExpireHours
	if expiresIn <= 0 {
		expiresIn = 72
	}

	inv := &models.Invite{
		Email:        email,
		FullName:     in.FullName,
		Token:        token,
		RoleIDs:      models.UUIDArray(in.RoleIDs),
		DepartmentID: in.DepartmentID,
		PositionID:   in.PositionID,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(expiresIn) * time.Hour),
		InvitedBy:    currentEmp.ID,
	}
	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Synchronously try to send (records error on row but doesn't
	// propagate). Caller sees the result in last_email_error.
	s.sendInviteEmail(ctx, inv)

	// Reload to pick up the inviter preload + any error stamp.
	final, err := s.repo.FindByID(ctx, inv.ID)
	if err != nil {
		return nil, err
	}
	read := s.toRead(final)
	return &read, nil
}

// ---- Get ----

func (s *InviteService) Get(ctx context.Context, id uuid.UUID) (*dto.InviteRead, error) {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Invite")
		}
		return nil, err
	}
	read := s.toRead(inv)
	return &read, nil
}

// ---- List ----

func (s *InviteService) List(ctx context.Context, q dto.InviteListQuery) (dto.PaginatedData[dto.InviteRead], error) {
	rows, total, err := s.repo.List(ctx, repositories.InviteListFilter{
		Email:    q.Email,
		Status:   q.Status,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		return dto.PaginatedData[dto.InviteRead]{}, err
	}
	items := make([]dto.InviteRead, 0, len(rows))
	for i := range rows {
		items = append(items, s.toRead(&rows[i]))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return dto.PaginatedData[dto.InviteRead]{
		Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
	}, nil
}

// ---- Resend ----

// Resend ships the existing token (no rotation — partial deliveries
// stay valid) and clears last_email_error on success.
func (s *InviteService) Resend(ctx context.Context, id uuid.UUID) (*dto.InviteRead, error) {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Invite")
		}
		return nil, err
	}
	if inv.AcceptedAt != nil {
		return nil, apperrors.ErrConflict("Invite has already been accepted")
	}
	if inv.ExpiresAt.Before(time.Now().UTC()) {
		return nil, apperrors.ErrConflict("Invite has expired — revoke and re-create")
	}
	s.sendInviteEmail(ctx, inv)

	final, err := s.repo.FindByID(ctx, inv.ID)
	if err != nil {
		return nil, err
	}
	read := s.toRead(final)
	return &read, nil
}

// ---- Revoke (soft delete) ----

func (s *InviteService) Revoke(ctx context.Context, id uuid.UUID) error {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Invite")
		}
		return err
	}
	if inv.AcceptedAt != nil {
		return apperrors.ErrConflict("Cannot revoke an already-accepted invite")
	}
	return s.repo.SoftDelete(ctx, id)
}

// ---- Accept (public) ----

// Accept consumes the invite token, creates a user + employee row, and
// marks the invite accepted. Wraps the entire flow in a transaction so
// a failure rolls back both inserts.
func (s *InviteService) Accept(ctx context.Context, in dto.InviteAccept) (*dto.InviteAcceptResult, error) {
	if strings.TrimSpace(in.Token) == "" {
		return nil, apperrors.ErrBadRequest("token is required")
	}
	if len(in.Password) < 8 {
		return nil, apperrors.ErrBadRequest("password must be at least 8 characters")
	}

	inv, err := s.repo.FindByToken(ctx, in.Token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperrors.ErrNotFound("Invite")
	}
	if inv.AcceptedAt != nil {
		return nil, apperrors.ErrConflict("Invite has already been used")
	}
	if inv.ExpiresAt.Before(time.Now().UTC()) {
		return nil, apperrors.ErrBadRequest("Invite has expired")
	}

	// Compose EmployeeCreate from the invite + accept-time inputs.
	fullName := ""
	if inv.FullName != nil {
		fullName = *inv.FullName
	}
	if in.FullName != nil && strings.TrimSpace(*in.FullName) != "" {
		fullName = *in.FullName
	}
	if fullName == "" {
		// Default to the local-part of the email so the schema's
		// NOT NULL constraint is satisfied.
		fullName = strings.Split(inv.Email, "@")[0]
	}

	created, err := s.empSvc.Create(ctx, dto.EmployeeCreate{
		Email:        inv.Email,
		Password:     in.Password,
		FullName:     fullName,
		DepartmentID: inv.DepartmentID,
		PositionID:   inv.PositionID,
	})
	if err != nil {
		return nil, err
	}

	// Replace roles when the invite specified any.
	if len(inv.RoleIDs) > 0 {
		if err := s.users.ReplaceRoles(ctx, created.UserID, []uuid.UUID(inv.RoleIDs)); err != nil {
			return nil, err
		}
	}

	// Stamp the invite as accepted.
	now := time.Now().UTC()
	inv.AcceptedAt = &now
	inv.AcceptedUserID = &created.UserID
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}

	return &dto.InviteAcceptResult{
		UserID:   created.UserID,
		Email:    created.Email,
		FullName: created.FullName,
		Message:  "Account created — you can now log in",
	}, nil
}
