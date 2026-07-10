package services

import (
	"context"
	"log"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
)

type announcementNotifier struct {
	push  *PushNotificationService
	email *EmailService
	users repositories.UserRepository
}

func NewAnnouncementNotifier(
	push *PushNotificationService,
	email *EmailService,
	users repositories.UserRepository,
) AnnouncementNotifier {
	return &announcementNotifier{push: push, email: email, users: users}
}

func (n *announcementNotifier) NotifyAnnouncement(ctx context.Context, userIDs []uuid.UUID, id uuid.UUID, title, description string) {
	for _, uid := range userIDs {
		if n.push != nil {
			req := dto.NotificationTestRequest{
				Title: title,
				Body:  plainTextPreview(description, 128),
				Data:  map[string]any{"type": "announcement", "id": id.String()},
			}
			result, err := n.push.SendToUser(ctx, uid, req)
			if err != nil {
				log.Printf("announcements: push to user %s: %v", uid, err)
			} else {
				log.Printf("announcements: push user=%s sent=%d skipped=%d errors=%v", uid, result.Sent, result.Skipped, result.Errors)
			}
		}
		if n.email != nil {
			user, err := n.users.FindByID(ctx, uid)
			if err != nil {
				log.Printf("announcements: lookup user %s for email: %v", uid, err)
				continue
			}
			if user.Email == "" {
				continue
			}
			if err := n.email.SendAnnouncementNotification(ctx, user.Email, title, description); err != nil {
				log.Printf("announcements: email to %s: %v", user.Email, err)
			}
		}
	}
}
