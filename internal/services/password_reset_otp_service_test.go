package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// otpTestConfig mirrors the shipped defaults. SMTP_HOST is empty, so the
// EmailService is in disabled mode throughout — the dispatched code never
// leaves the process, which is exactly why the verify tests seed their own
// row with a known code instead of scraping one out of RequestOTP.
func otpTestConfig() *config.Config {
	return &config.Config{
		AppName:                    "Exnodes HRM Test",
		OTPExpireMinutes:           10,
		OTPResendCooldownSeconds:   60,
		OTPMaxRequestsPerWindow:    3,
		OTPRateLimitWindowMinutes:  15,
		OTPMaxVerifyAttempts:       5,
		OTPResetTokenExpireMinutes: 10,
	}
}

type otpFixture struct {
	svc       *services.PasswordResetOTPService
	resetSvc  *services.PasswordResetService
	otpRepo   repositories.PasswordResetOTPRepository
	tokenRepo repositories.PasswordResetTokenRepository
	cfg       *config.Config
}

func newOTPFixture(t *testing.T) *otpFixture {
	t.Helper()
	cfg := otpTestConfig()
	emailSvc, err := services.NewEmailService(cfg)
	require.NoError(t, err)
	require.False(t, emailSvc.IsConfigured(), "test SMTP must stay disabled")

	otpRepo := repositories.NewPasswordResetOTPRepository(testDB)
	tokenRepo := repositories.NewPasswordResetTokenRepository(testDB)
	return &otpFixture{
		svc:       services.NewPasswordResetOTPService(testUserRepo, otpRepo, tokenRepo, emailSvc, cfg),
		resetSvc:  services.NewPasswordResetService(testUserRepo, tokenRepo, emailSvc, cfg),
		otpRepo:   otpRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

// seedOTP inserts a code with a known plaintext so verification can be
// exercised. Mirrors what RequestOTP persists.
func seedOTP(t *testing.T, userID uuid.UUID, code string, expiresAt time.Time) *models.PasswordResetOTP {
	t.Helper()
	hash, err := utils.HashPassword(code)
	require.NoError(t, err)
	o := &models.PasswordResetOTP{UserID: userID, CodeHash: hash, ExpiresAt: expiresAt}
	require.NoError(t, testDB.Create(o).Error)
	return o
}

// backdateOTPs rewinds created_at for every code belonging to the user, so a
// test can step past the resend cooldown without sleeping.
func backdateOTPs(t *testing.T, userID uuid.UUID, d time.Duration) {
	t.Helper()
	err := testDB.Exec(
		`UPDATE password_reset_otps SET created_at = created_at - $1::interval WHERE user_id = $2`,
		d.String(), userID,
	).Error
	require.NoError(t, err)
}

func latestOTPRow(t *testing.T, userID uuid.UUID) *models.PasswordResetOTP {
	t.Helper()
	var o models.PasswordResetOTP
	require.NoError(t, testDB.Where("user_id = ?", userID).Order("created_at DESC").First(&o).Error)
	return &o
}

func requireAppError(t *testing.T, err error, wantHTTP int, wantMsg string) *apperr.AppError {
	t.Helper()
	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok, "expected *AppError, got %T", err)
	require.Equal(t, wantHTTP, ae.HTTP)
	if wantMsg != "" {
		require.Equal(t, wantMsg, ae.Message)
	}
	return ae
}

// AC-05 / SR-08: the mobile flow deliberately drops the web flow's enumerate
// guard so the app can show an inline "No account found" error. If this ever
// flips back to a 200, the app silently sends users to an OTP screen for an
// address that will never receive one.
func TestOTPRequest_UnknownEmail_404(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	_, err := f.svc.RequestOTP(context.Background(), "nobody@example.com")
	requireAppError(t, err, 404, "No account found with this email address")
}

// SR-01: a deactivated account must be indistinguishable from a missing one,
// or the endpoint reports employment status to anyone with an email address.
func TestOTPRequest_InactiveUser_404(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "inactive@example.com", "pw-Aa123456")
	require.NoError(t, testUserRepo.ToggleActive(context.Background(), u.ID, false))

	_, err := f.svc.RequestOTP(context.Background(), "inactive@example.com")
	requireAppError(t, err, 404, "No account found with this email address")
}

// Happy path. The SMTP failure must be recorded on the row and NOT surfaced:
// a dead mail server cannot be allowed to strand the user on Screen 1 with a
// 500 (the invites.last_email_error contract, applied to codes).
func TestOTPRequest_CreatesHashedCodeAndSwallowsEmailFailure(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")
	makeEmployee(t, u, "Test User")

	res, err := f.svc.RequestOTP(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().UTC().Add(10*time.Minute), res.ExpiresAt, time.Minute)
	require.WithinDuration(t, time.Now().UTC().Add(60*time.Second), res.ResendAvailableAt, time.Minute)

	row := latestOTPRow(t, u.ID)
	require.NotEmpty(t, row.CodeHash)
	require.NotRegexp(t, `^\d{6}$`, row.CodeHash, "code must be hashed at rest, never stored plaintext")
	require.Nil(t, row.ConsumedAt)
	require.Equal(t, 0, row.AttemptCount)
	require.Contains(t, row.LastEmailError, "SMTP not configured")
}

// SR-03: only the newest code may work. Without this, every code issued in
// the 15-minute window would stay live and multiply the guessing surface.
func TestOTPRequest_SupersedesPreviousCode(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")

	_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
	require.NoError(t, err)
	first := latestOTPRow(t, u.ID)

	backdateOTPs(t, u.ID, 2*time.Minute) // step past the resend cooldown
	_, err = f.svc.RequestOTP(context.Background(), "user@example.com")
	require.NoError(t, err)

	var old models.PasswordResetOTP
	require.NoError(t, testDB.Where("id = ?", first.ID).First(&old).Error)
	require.True(t, old.IsDeleted, "issuing a new code must kill the previous one")

	active, err := f.otpRepo.FindLatestActiveForUser(context.Background(), u.ID)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, active.ID)
}

// UX-05 / Alt-4: the resend cooldown is enforced server-side. The countdown in
// the app is a hint, not the control.
func TestOTPRequest_ResendCooldown_429(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	makeUser(t, "user@example.com", "pw-Aa123456")

	_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
	require.NoError(t, err)

	_, err = f.svc.RequestOTP(context.Background(), "user@example.com")
	ae := requireAppError(t, err, 429, "")
	require.Contains(t, ae.Message, "seconds before requesting another code")
	require.Contains(t, ae.Details, "retry_after_seconds")
}

// AC-06 / SR-04: 3 requests per 15 minutes.
//
// The regression this pins: superseding a code soft-deletes its row, so a
// rate-limit count that filtered out deleted rows would reset the budget on
// every request and the limit would never fire. Each request below supersedes
// the last, so the 4th only fails if deleted rows are still counted.
func TestOTPRequest_RateLimitCountsSupersededCodes_429(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")

	for i := 0; i < 3; i++ {
		_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
		require.NoErrorf(t, err, "request %d should be allowed", i+1)
		backdateOTPs(t, u.ID, 2*time.Minute) // clear the cooldown, stay inside the 15m window
	}

	_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
	ae := requireAppError(t, err, 429, "Too many attempts. Please wait 15 minutes before trying again.")
	require.Contains(t, ae.Details, "retry_after_seconds")
}

// The window must actually roll: once the oldest requests age out, the user
// gets a fresh budget. A limit that never releases locks people out for good.
func TestOTPRequest_RateLimitWindowRolls(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")

	for i := 0; i < 3; i++ {
		_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
		require.NoError(t, err)
		backdateOTPs(t, u.ID, 2*time.Minute)
	}
	_, err := f.svc.RequestOTP(context.Background(), "user@example.com")
	requireAppError(t, err, 429, "")

	backdateOTPs(t, u.ID, 16*time.Minute) // every prior request now outside the window
	_, err = f.svc.RequestOTP(context.Background(), "user@example.com")
	require.NoError(t, err)
}

// AC-10 + AC-20 end to end: verify hands back a reset token, that token drives
// the shared /auth/reset-password path, and the new password actually logs in.
// This is the whole point of the feature — if it passes nothing else matters,
// and if it fails nothing else does.
func TestOTPVerify_HappyPath_TokenResetsPassword(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "old-Pw123456")
	seedOTP(t, u.ID, "482913", time.Now().UTC().Add(10*time.Minute))

	res, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "482913")
	require.NoError(t, err)
	require.NotEmpty(t, res.ResetToken)
	require.WithinDuration(t, time.Now().UTC().Add(10*time.Minute), res.ExpiresAt, time.Minute)

	// SR-05: consumed at verify, not at reset.
	row := latestOTPRow(t, u.ID)
	require.NotNil(t, row.ConsumedAt)

	require.NoError(t, f.resetSvc.ResetWithToken(context.Background(), res.ResetToken, "new-Pw123456"))

	updated, err := testUserRepo.FindByID(context.Background(), u.ID)
	require.NoError(t, err)
	require.True(t, utils.CheckPassword("new-Pw123456", updated.PasswordHash), "new password must authenticate")
	require.False(t, utils.CheckPassword("old-Pw123456", updated.PasswordHash), "old password must stop working")
	require.NotNil(t, updated.PasswordResetAt, "reset must stamp password_reset_at")
}

// SR-05, stated as the behaviour the app must cope with: the code dies at
// verify. Alt-3's "back out of Screen 3 and the OTP still works" is NOT
// implemented — the app has to hold the reset_token. If this test starts
// failing, the Alt-3 reading has been restored and the DR conflict is live again.
func TestOTPVerify_ReplayAfterConsume_Rejected(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")
	seedOTP(t, u.ID, "482913", time.Now().UTC().Add(10*time.Minute))

	_, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "482913")
	require.NoError(t, err)

	_, err = f.svc.VerifyOTP(context.Background(), "user@example.com", "482913")
	requireAppError(t, err, 400, "This code has expired. Please request a new code.")
}

// AC-11: wrong code keeps the user on the screen with a countdown of what's
// left, rather than silently burning the code on the first typo.
func TestOTPVerify_WrongCode_ReportsRemainingAttempts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")
	seedOTP(t, u.ID, "482913", time.Now().UTC().Add(10*time.Minute))

	_, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "000000")
	ae := requireAppError(t, err, 400, "Incorrect code. Please check and try again.")
	require.Equal(t, 4, ae.Details["remaining_attempts"])

	_, err = f.svc.VerifyOTP(context.Background(), "user@example.com", "111111")
	ae = requireAppError(t, err, 400, "Incorrect code. Please check and try again.")
	require.Equal(t, 3, ae.Details["remaining_attempts"], "attempts must accumulate across requests")
}

// The online-guessing bound. A 6-digit code is only 10^6 wide and the
// per-window request limit does not cap verify attempts, so without this an
// attacker holding one issued code could walk the whole space in 10 minutes.
// Asserts the code is dead afterwards even for the CORRECT value.
func TestOTPVerify_BurnsCodeAfterMaxAttempts(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")
	seedOTP(t, u.ID, "482913", time.Now().UTC().Add(10*time.Minute))

	for i := 0; i < 4; i++ {
		_, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "000000")
		requireAppError(t, err, 400, "Incorrect code. Please check and try again.")
	}

	_, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "000000")
	ae := requireAppError(t, err, 400, "Too many incorrect attempts. Please request a new code.")
	require.Equal(t, 0, ae.Details["remaining_attempts"])

	_, err = f.svc.VerifyOTP(context.Background(), "user@example.com", "482913")
	requireAppError(t, err, 400, "This code has expired. Please request a new code.")
}

// AC-12: an expired code reports as expired (with the Resend affordance),
// not as incorrect — the user must not retype a code that can never work.
func TestOTPVerify_ExpiredCode_400(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	u := makeUser(t, "user@example.com", "pw-Aa123456")
	seedOTP(t, u.ID, "482913", time.Now().UTC().Add(-1*time.Minute))

	_, err := f.svc.VerifyOTP(context.Background(), "user@example.com", "482913")
	requireAppError(t, err, 400, "This code has expired. Please request a new code.")
}

func TestOTPVerify_UnknownEmail_404(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	_, err := f.svc.VerifyOTP(context.Background(), "nobody@example.com", "482913")
	requireAppError(t, err, 404, "No account found with this email address")
}

// A code issued for one account must never verify against another, even when
// the digits happen to match.
func TestOTPVerify_CodeIsScopedToItsAccount(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	f := newOTPFixture(t)

	victim := makeUser(t, "victim@example.com", "pw-Aa123456")
	makeUser(t, "attacker@example.com", "pw-Aa123456")
	seedOTP(t, victim.ID, "482913", time.Now().UTC().Add(10*time.Minute))

	_, err := f.svc.VerifyOTP(context.Background(), "attacker@example.com", "482913")
	requireAppError(t, err, 400, "This code has expired. Please request a new code.")
}
