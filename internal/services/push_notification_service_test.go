package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// fakePushClient is a test double for services.PushClient. Records
// every Send call for assertions; controllable via Configured + ErrOn.
type fakePushClient struct {
	Configured bool
	ErrOn      map[string]error // per-token error map
	Sent       []services.PushMessage
}

func (f *fakePushClient) Send(ctx context.Context, msg services.PushMessage) error {
	f.Sent = append(f.Sent, msg)
	if e, ok := f.ErrOn[msg.Token]; ok {
		return e
	}
	return nil
}

func (f *fakePushClient) IsConfigured() bool { return f.Configured }

// ---- Tests ----

func TestPush_SendToUser_NoConfig_AllSkipped(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	u, _ := makeEmpUser(t, "pushuser@example.com", "Push")
	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-a", Token: "token-a", Platform: "ios"}))
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-b", Token: "token-b", Platform: "android"}))

	fake := &fakePushClient{Configured: false}
	svc := services.NewPushNotificationService(fake, tokenRepo)

	out, err := svc.SendToUser(ctx, u.ID, dto.NotificationTestRequest{Title: "hi", Body: "there"})
	require.NoError(t, err)
	assert.Equal(t, 0, out.Sent)
	assert.Equal(t, 2, out.Skipped)
	assert.Empty(t, fake.Sent, "no-op client must not be invoked when IsConfigured=false")
}

func TestPush_SendToUser_AllDelivered(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	u, _ := makeEmpUser(t, "pushuser@example.com", "Push")
	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-a", Token: "token-a", Platform: "ios"}))
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-b", Token: "token-b", Platform: "android"}))

	fake := &fakePushClient{Configured: true}
	svc := services.NewPushNotificationService(fake, tokenRepo)

	out, err := svc.SendToUser(ctx, u.ID, dto.NotificationTestRequest{Title: "hi", Body: "there", Data: map[string]any{"k": "v"}})
	require.NoError(t, err)
	assert.Equal(t, 2, out.Sent)
	assert.Equal(t, 0, out.Skipped)
	require.Len(t, fake.Sent, 2)
	assert.Equal(t, "hi", fake.Sent[0].Title)
	assert.Equal(t, "v", fake.Sent[0].Data["k"])
}

func TestPush_SendToUser_PartialFailures_AggregatesErrors(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	u, _ := makeEmpUser(t, "pushuser@example.com", "Push")
	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-good", Token: "token-good", Platform: "ios"}))
	require.NoError(t, tokenRepo.Upsert(ctx, &models.DeviceToken{UserID: u.ID, DeviceID: "dev-bad", Token: "token-bad", Platform: "android"}))

	fake := &fakePushClient{
		Configured: true,
		ErrOn:      map[string]error{"token-bad": errors.New("invalid token")},
	}
	svc := services.NewPushNotificationService(fake, tokenRepo)

	out, err := svc.SendToUser(ctx, u.ID, dto.NotificationTestRequest{Title: "hi", Body: "there"})
	require.NoError(t, err)
	assert.Equal(t, 1, out.Sent)
	assert.Equal(t, 1, out.Skipped)
	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "invalid token")
}

func TestPush_SendToUser_NoDevices_EmptyResult(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	u, _ := makeEmpUser(t, "nodevices@example.com", "NoD")
	tokenRepo := repositories.NewDeviceTokenRepository(testDB)
	fake := &fakePushClient{Configured: true}
	svc := services.NewPushNotificationService(fake, tokenRepo)

	out, err := svc.SendToUser(ctx, u.ID, dto.NotificationTestRequest{Title: "hi", Body: "there"})
	require.NoError(t, err)
	assert.Equal(t, 0, out.Sent)
	assert.Equal(t, 0, out.Skipped)
	assert.Empty(t, out.Errors)
}

// ---- EmailService ----

func TestEmail_IsConfigured_FalseWhenHostEmpty(t *testing.T) {
	cfg := inviteTestConfig()
	svc, err := services.NewEmailService(cfg)
	require.NoError(t, err)
	assert.False(t, svc.IsConfigured())
}

func TestEmail_SendInvite_ReturnsErrEmailDisabled_WhenSMTPEmpty(t *testing.T) {
	cfg := inviteTestConfig()
	svc, err := services.NewEmailService(cfg)
	require.NoError(t, err)

	err = svc.SendInvite(context.Background(), "to@example.com", services.InviteEmailData{
		AppName:   "Test",
		FullName:  "Newbie",
		AcceptURL: "http://example.com/accept?token=x",
		ExpiresAt: "soon",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, services.ErrEmailDisabled))
}

// Ensure unused import warning is silenced — uuid is referenced via
// the device token UserID assignments above.
var _ = uuid.Nil
