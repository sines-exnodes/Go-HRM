package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

const (
	notificationDefaultPageSize = 50
	notificationMaxPageSize     = 100
)

// NotificationService owns the in-app notification feed. Every read and
// write is scoped to a single user's ID — there is no unscoped path, which
// is how DR AC-01 (an employee never sees another's notifications) is
// enforced server-side.
type NotificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// List returns the user's notifications, newest first.
func (s *NotificationService) List(
	ctx context.Context,
	userID uuid.UUID,
	q dto.NotificationListQuery,
) (dto.PaginatedData[dto.NotificationRead], *apperrors.AppError) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = notificationDefaultPageSize
	}
	if pageSize > notificationMaxPageSize {
		pageSize = notificationMaxPageSize
	}

	rows, total, err := s.repo.List(ctx, repositories.NotificationListQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return dto.PaginatedData[dto.NotificationRead]{}, apperrors.ErrInternal(err.Error())
	}

	items := make([]dto.NotificationRead, 0, len(rows))
	for i := range rows {
		items = append(items, notificationToRead(&rows[i]))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return dto.PaginatedData[dto.NotificationRead]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UnreadCount backs the dashboard header bell.
func (s *NotificationService) UnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (dto.NotificationUnreadCountRead, *apperrors.AppError) {
	n, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return dto.NotificationUnreadCountRead{}, apperrors.ErrInternal(err.Error())
	}
	return dto.NotificationUnreadCountRead{UnreadCount: n}, nil
}

// MarkRead stamps the notification read for this user and returns it.
//
// Marking an already-read notification is a 200 no-op, not a 409: DR Rule 8
// makes read terminal, so a repeat is a successful arrival at the intended
// state, and the mobile client may legitimately retry after a dropped
// response.
func (s *NotificationService) MarkRead(
	ctx context.Context,
	id, userID uuid.UUID,
) (*dto.NotificationRead, *apperrors.AppError) {
	row, err := s.repo.MarkRead(ctx, id, userID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Notification")
		}
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := notificationToRead(row)
	return &out, nil
}

// CreateMany is the producer-facing entry point. Collisions on
// uq_notifications_user_source are skipped, so callers get DR Rule 5
// (one notification per event) for free on retry.
func (s *NotificationService) CreateMany(ctx context.Context, rows []models.Notification) error {
	return s.repo.CreateMany(ctx, rows)
}

func notificationToRead(m *models.Notification) dto.NotificationRead {
	return dto.NotificationRead{
		ID:        m.ID,
		Type:      string(m.Type),
		Title:     m.Title,
		Body:      m.Body,
		SourceID:  m.SourceID,
		IsRead:    m.ReadAt != nil,
		ReadAt:    m.ReadAt,
		CreatedAt: m.CreatedAt,
	}
}
