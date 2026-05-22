package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// PushNotificationService composes the PushClient + DeviceTokenRepository
// — looks up the user's tokens, dispatches a message to each, and
// aggregates the per-token outcomes.
type PushNotificationService struct {
	client PushClient
	tokens *repositories.DeviceTokenRepository
}

func NewPushNotificationService(client PushClient, tokens *repositories.DeviceTokenRepository) *PushNotificationService {
	return &PushNotificationService{client: client, tokens: tokens}
}

// SendToUser delivers a push to every registered device for userID.
// Returns a result envelope describing the outcome (sent / skipped /
// errors). Never returns a top-level error for transport failures —
// the per-device errors are aggregated in the result instead.
func (s *PushNotificationService) SendToUser(ctx context.Context, userID uuid.UUID, req dto.NotificationTestRequest) (*dto.NotificationTestResult, error) {
	tokens, err := s.tokens.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := &dto.NotificationTestResult{Errors: []string{}}
	if !s.client.IsConfigured() {
		out.Skipped = len(tokens)
		return out, nil
	}

	for _, t := range tokens {
		if err := s.client.Send(ctx, PushMessage{
			Token: t.Token,
			Title: req.Title,
			Body:  req.Body,
			Data:  req.Data,
		}); err != nil {
			out.Errors = append(out.Errors, err.Error())
			out.Skipped++
			continue
		}
		out.Sent++
	}
	return out, nil
}
