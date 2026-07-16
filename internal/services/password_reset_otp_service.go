package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// User-facing messages. The mobile app renders these verbatim (DR-001-001-02
// §2 Exit Points), so changing them changes the UI.
const (
	msgOTPEmailNotFound  = "No account found with this email address"
	msgOTPIncorrect      = "Incorrect code. Please check and try again."
	msgOTPExpired        = "This code has expired. Please request a new code."
	msgOTPRateLimited    = "Too many attempts. Please wait 15 minutes before trying again."
	msgOTPTooManyRetries = "Too many incorrect attempts. Please request a new code."
)

// PasswordResetOTPService implements the mobile forgot-password flow
// (DR-001-001-02). It is deliberately separate from PasswordResetService,
// which serves the web one-click-link flow — the two share the users table
// and the password_reset_tokens table but nothing else.
//
// Flow:
//  1. RequestOTP   — mails a 6-digit code (also used for Resend, Alt-4).
//  2. VerifyOTP    — consumes the code, returns a short-lived reset token.
//  3. The app then calls the existing POST /auth/reset-password with that
//     token, which is why no third method lives here.
//
// Divergences from the DR, decided with the project owner:
//   - Unknown / inactive email returns 404 (AC-05, SR-08) rather than the
//     enumerate-guarded 200 the web flow uses. This makes the endpoint an
//     account-existence oracle — accepted deliberately.
//   - SR-05 wins over Alt-3: the code is consumed at verify. Backing out of
//     the new-password screen and re-entering the same code fails; the app
//     must hold the reset token it was handed.
//   - SR-09 (kill sessions on every device) is NOT implemented — see
//     CHECKPOINT. UpdatePassword stamps users.password_reset_at but nothing
//     enforces it, so previously-issued access tokens live out their TTL.
type PasswordResetOTPService struct {
	users  repositories.UserRepository
	otps   repositories.PasswordResetOTPRepository
	tokens repositories.PasswordResetTokenRepository
	email  *EmailService
	cfg    *config.Config
}

// NewPasswordResetOTPService constructs the service.
func NewPasswordResetOTPService(
	users repositories.UserRepository,
	otps repositories.PasswordResetOTPRepository,
	tokens repositories.PasswordResetTokenRepository,
	email *EmailService,
	cfg *config.Config,
) *PasswordResetOTPService {
	return &PasswordResetOTPService{users: users, otps: otps, tokens: tokens, email: email, cfg: cfg}
}

// OTPRequestResult tells the app when the code dies and when the Resend link
// may be re-enabled (UX-05 countdown).
type OTPRequestResult struct {
	ExpiresAt         time.Time
	ResendAvailableAt time.Time
}

// OTPVerifyResult carries the single-use token that authorizes the actual
// password change on the next screen.
type OTPVerifyResult struct {
	ResetToken string
	ExpiresAt  time.Time
}

// RequestOTP issues a fresh 6-digit code for the email and mails it. It backs
// both "Send Code" (Screen 1) and "Resend Code" (Screen 2, Alt-4) — the same
// cooldown and rate limit apply to each, so the two need no separate paths.
//
// Errors: 404 when the email has no active account (AC-05); 429 when the
// resend cooldown or the per-window request limit is hit (AC-06, SR-04).
func (s *PasswordResetOTPService) RequestOTP(ctx context.Context, email string) (*OTPRequestResult, error) {
	user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFoundMsg(msgOTPEmailNotFound)
		}
		return nil, apperr.ErrInternal(err.Error())
	}
	// SR-01: a deactivated account is indistinguishable from a missing one.
	if !user.IsActive {
		return nil, apperr.ErrNotFoundMsg(msgOTPEmailNotFound)
	}

	now := time.Now().UTC()

	// SR-04 window limit. CountCreatedSince counts soft-deleted rows too, so
	// superseding a code does not refill the budget.
	windowStart := now.Add(-time.Duration(s.cfg.OTPRateLimitWindowMinutes) * time.Minute)
	issued, err := s.otps.CountCreatedSince(ctx, user.ID, windowStart)
	if err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}
	if issued >= int64(s.cfg.OTPMaxRequestsPerWindow) {
		return nil, apperr.ErrTooManyRequests(msgOTPRateLimited).WithDetails(map[string]any{
			"retry_after_seconds": s.cfg.OTPRateLimitWindowMinutes * 60,
		})
	}

	// Resend cooldown. Uses the latest row of any state: a burned or consumed
	// code still means an email was just sent to this address.
	if latest, err := s.otps.FindLatestForUser(ctx, user.ID); err == nil {
		cooldown := time.Duration(s.cfg.OTPResendCooldownSeconds) * time.Second
		if elapsed := now.Sub(latest.CreatedAt.UTC()); elapsed < cooldown {
			wait := int((cooldown - elapsed).Seconds()) + 1
			return nil, apperr.ErrTooManyRequests(
				fmt.Sprintf("Please wait %d seconds before requesting another code.", wait),
			).WithDetails(map[string]any{"retry_after_seconds": wait})
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrInternal(err.Error())
	}

	// SR-03: issuing a new code kills every earlier pending one.
	if err := s.otps.InvalidatePendingForUser(ctx, user.ID); err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}

	code, err := generateOTPCode()
	if err != nil {
		return nil, apperr.ErrInternal("Failed to generate verification code")
	}
	codeHash, err := utils.HashPassword(code)
	if err != nil {
		return nil, apperr.ErrInternal("Failed to hash verification code")
	}

	expiresAt := now.Add(time.Duration(s.cfg.OTPExpireMinutes) * time.Minute)
	otp := &models.PasswordResetOTP{
		UserID:    user.ID,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
	}
	if err := s.otps.Create(ctx, otp); err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}

	fullName := ""
	if user.Employee != nil {
		fullName = user.Employee.FullName()
	}

	// Email is best-effort: a dead SMTP must not fail the request, or the
	// user is stuck on a screen with no way forward. The error lands on the
	// row for operators (same contract as invites.last_email_error).
	if emailErr := s.email.SendPasswordResetOTP(ctx, user.Email, PasswordResetOTPEmailData{
		AppName:          s.cfg.AppName,
		FullName:         fullName,
		Code:             code,
		ExpiresAt:        expiresAt.Format("15:04 UTC"),
		ExpiresInMinutes: s.cfg.OTPExpireMinutes,
	}); emailErr != nil {
		if updateErr := s.otps.UpdateEmailError(ctx, otp.ID, emailErr.Error()); updateErr != nil {
			log.Printf("password_reset_otp: failed to store email error on otp %s: %v", otp.ID, updateErr)
		}
	}

	return &OTPRequestResult{
		ExpiresAt:         expiresAt,
		ResendAvailableAt: now.Add(time.Duration(s.cfg.OTPResendCooldownSeconds) * time.Second),
	}, nil
}

// VerifyOTP checks the submitted code and, on success, consumes it (SR-05)
// and mints a single-use reset token the app hands to POST
// /auth/reset-password. The token reuses the web flow's
// password_reset_tokens table with a short TTL.
//
// Errors: 404 for an unknown/inactive email; 400 for a wrong, expired, or
// already-consumed code. Wrong-code responses carry a remaining_attempts
// detail so the app can render the AC-11 hint.
func (s *PasswordResetOTPService) VerifyOTP(ctx context.Context, email, code string) (*OTPVerifyResult, error) {
	user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFoundMsg(msgOTPEmailNotFound)
		}
		return nil, apperr.ErrInternal(err.Error())
	}
	if !user.IsActive {
		return nil, apperr.ErrNotFoundMsg(msgOTPEmailNotFound)
	}

	otp, err := s.otps.FindLatestActiveForUser(ctx, user.ID)
	if err != nil {
		// No live code — consumed, burned, or never issued. All are "expired"
		// to the user: the fix is the same in every case (request a new one).
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrBadRequest(msgOTPExpired)
		}
		return nil, apperr.ErrInternal(err.Error())
	}

	now := time.Now().UTC()
	if now.After(otp.ExpiresAt) {
		return nil, apperr.ErrBadRequest(msgOTPExpired)
	}

	if !utils.CheckPassword(code, otp.CodeHash) {
		attempts, incErr := s.otps.IncrementAttempts(ctx, otp.ID)
		if incErr != nil {
			return nil, apperr.ErrInternal(incErr.Error())
		}
		// Bound the online guessing budget — 6 digits is only 10^6 wide, and
		// the per-window request limit does not cap verify attempts.
		if attempts >= s.cfg.OTPMaxVerifyAttempts {
			if burnErr := s.otps.Burn(ctx, otp.ID); burnErr != nil {
				log.Printf("password_reset_otp: failed to burn otp %s after %d attempts: %v", otp.ID, attempts, burnErr)
			}
			return nil, apperr.ErrBadRequest(msgOTPTooManyRetries).WithDetails(map[string]any{
				"remaining_attempts": 0,
			})
		}
		return nil, apperr.ErrBadRequest(msgOTPIncorrect).WithDetails(map[string]any{
			"remaining_attempts": s.cfg.OTPMaxVerifyAttempts - attempts,
		})
	}

	// SR-05: consume on verify, before the token exists. If token creation
	// then fails the user must request a new code — the safe direction.
	if err := s.otps.MarkConsumed(ctx, otp.ID, now); err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return nil, apperr.ErrInternal("Failed to generate reset token")
	}
	// Drop any web-flow link token still pending for this user: two live
	// paths to the same password change is one more than we want.
	if err := s.tokens.InvalidatePendingForUser(ctx, user.ID); err != nil {
		log.Printf("password_reset_otp: failed to invalidate pending reset tokens for %s: %v", user.ID, err)
	}

	tokenExpiresAt := now.Add(time.Duration(s.cfg.OTPResetTokenExpireMinutes) * time.Minute)
	prt := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     rawToken,
		ExpiresAt: tokenExpiresAt,
	}
	if err := s.tokens.Create(ctx, prt); err != nil {
		return nil, apperr.ErrInternal(err.Error())
	}

	return &OTPVerifyResult{ResetToken: rawToken, ExpiresAt: tokenExpiresAt}, nil
}

// generateOTPCode returns a uniformly-distributed 6-digit code drawn from
// crypto/rand. Zero-padded, so "000042" is a valid code and the space is the
// full 10^6.
func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
