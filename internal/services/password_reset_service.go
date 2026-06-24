package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// PasswordResetService implements the self-service forgot-password flow:
//  1. RequestReset — always returns nil (enumerate guard); creates a token and
//     sends an email best-effort, recording send errors on the token row.
//  2. ResetWithToken — validates the token and applies the new password hash.
type PasswordResetService struct {
	users  repositories.UserRepository
	tokens repositories.PasswordResetTokenRepository
	email  *EmailService
	cfg    *config.Config
}

// NewPasswordResetService constructs the service.
func NewPasswordResetService(
	users repositories.UserRepository,
	tokens repositories.PasswordResetTokenRepository,
	email *EmailService,
	cfg *config.Config,
) *PasswordResetService {
	return &PasswordResetService{users: users, tokens: tokens, email: email, cfg: cfg}
}

// RequestReset initiates a password reset for the given email address.
// It always returns nil regardless of whether the email exists — callers
// must never expose a different response for known vs unknown emails.
func (s *PasswordResetService) RequestReset(ctx context.Context, email string) error {
	user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("password_reset: no user for %q — returning 200 (enumerate guard)", email)
			return nil
		}
		log.Printf("password_reset: db error for %q: %v — returning 200 (enumerate guard)", email, err)
		return nil
	}
	if !user.IsActive {
		log.Printf("password_reset: user %s inactive — returning 200 (enumerate guard)", user.ID)
		return nil
	}

	// Invalidate any pending tokens so only one is active at a time.
	if err := s.tokens.InvalidatePendingForUser(ctx, user.ID); err != nil {
		log.Printf("password_reset: failed to invalidate old tokens for %s: %v", user.ID, err)
	}

	rawToken, err := generateResetToken()
	if err != nil {
		log.Printf("password_reset: token generation failed: %v", err)
		return nil
	}

	expireHours := s.cfg.PasswordResetTokenExpireHours
	if expireHours <= 0 {
		expireHours = 1
	}
	prt := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().UTC().Add(time.Duration(expireHours) * time.Hour),
	}
	if err := s.tokens.Create(ctx, prt); err != nil {
		log.Printf("password_reset: failed to persist token: %v", err)
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, rawToken)

	fullName := ""
	if user.Employee != nil {
		fullName = user.Employee.FullName()
	}

	emailErr := s.email.SendPasswordReset(ctx, email, PasswordResetEmailData{
		AppName:   s.cfg.AppName,
		FullName:  fullName,
		ResetURL:  resetURL,
		ExpiresAt: prt.ExpiresAt.UTC().Format("02 Jan 2006 15:04 UTC"),
	})
	if emailErr != nil {
		if updateErr := s.tokens.UpdateEmailError(ctx, prt.ID, emailErr.Error()); updateErr != nil {
			log.Printf("password_reset: failed to store email error on token %s: %v", prt.ID, updateErr)
		}
	}

	return nil
}

// VerifyToken checks whether a reset token is valid (exists, not used, not expired).
// Returns the associated User so the caller can show their name/email on the set-password page.
// Does NOT consume the token.
func (s *PasswordResetService) VerifyToken(ctx context.Context, token string) (*models.User, error) {
	prt, err := s.tokens.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrBadRequest("Invalid or expired reset token")
		}
		return nil, apperr.ErrInternal(err.Error())
	}
	if prt.UsedAt != nil {
		return nil, apperr.ErrBadRequest("Reset token has already been used")
	}
	if time.Now().UTC().After(prt.ExpiresAt) {
		return nil, apperr.ErrBadRequest("Reset token has expired")
	}
	user, err := s.users.FindByIDWithRolesAndEmployee(ctx, prt.UserID)
	if err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}
	return user, nil
}

// ResetWithToken validates the token and applies the new hashed password.
// Returns a 400 AppError for invalid/expired/used tokens.
func (s *PasswordResetService) ResetWithToken(ctx context.Context, token, newPassword string) error {
	prt, err := s.tokens.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrBadRequest("Invalid or expired reset token")
		}
		return apperr.ErrInternal(err.Error())
	}

	if prt.UsedAt != nil {
		return apperr.ErrBadRequest("Reset token has already been used")
	}
	if time.Now().UTC().After(prt.ExpiresAt) {
		return apperr.ErrBadRequest("Reset token has expired")
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return apperr.ErrInternal("Failed to hash password")
	}

	if err := s.users.UpdatePassword(ctx, prt.UserID, hashed); err != nil {
		return apperr.ErrInternal(err.Error())
	}

	now := time.Now().UTC()
	if err := s.tokens.MarkUsed(ctx, prt.ID, now); err != nil {
		log.Printf("password_reset: failed to mark token %s used: %v", prt.ID, err)
	}

	return nil
}

// generateResetToken returns 32 bytes of crypto-random data encoded as
// URL-safe base64 (no padding). The result is safe to embed in a URL
// query parameter without further encoding.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
