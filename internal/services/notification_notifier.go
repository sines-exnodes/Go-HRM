package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// leaveDecisionDateFormat renders dates as "10 June, 2026" — the D MMMM, YYYY
// style DR-MOB-005-001-01 uses for user-facing dates (see AC-16). The day is
// unpadded, so 3 August reads "3 August, 2026", not "03 August, 2026".
//
// This is the one place notification dates are formatted, and unlike the
// timestamp on the row (which the client formats from RFC3339), this text is
// baked into the stored body at creation time — a snapshot, per Rule 12.
// Changing it does NOT rewrite notifications that already exist.
const leaveDecisionDateFormat = "2 January, 2006"

// leaveNotifier writes an in-app notification when a leave request is
// approved or rejected (DR Rule 4).
//
// The leave aggregate keys on employees(id) but notifications key on
// users(id), so this type owns the employee → user resolution. That
// translation is exactly why the notifier is a separate collaborator rather
// than a method on LeaveService.
type leaveNotifier struct {
	notifs *NotificationService
	emps   repositories.EmployeeRepository
	push   *PushNotificationService // optional — nil disables OS-level push
}

// NewLeaveNotifier constructs the concrete LeaveNotifier. Pass nil for `push`
// to write only the in-app feed row (what most service tests do).
func NewLeaveNotifier(
	notifs *NotificationService,
	emps repositories.EmployeeRepository,
	push *PushNotificationService,
) LeaveNotifier {
	return &leaveNotifier{notifs: notifs, emps: emps, push: push}
}

func (n *leaveNotifier) NotifyLeaveDecision(
	ctx context.Context,
	employeeID, leaveID uuid.UUID,
	approved bool,
	from, to time.Time,
) {
	if n.notifs == nil || n.emps == nil {
		return
	}

	emp, err := n.emps.FindByID(ctx, employeeID)
	if err != nil {
		log.Printf("leave: resolve employee %s for notification: %v", employeeID, err)
		return
	}

	title, body := leaveDecisionCopy(approved, from, to)

	err = n.notifs.CreateMany(ctx, []models.Notification{{
		UserID:   emp.UserID,
		Type:     models.NotificationTypeLeaveRequest,
		Title:    title,
		Body:     body,
		SourceID: leaveID,
	}})
	if err != nil {
		log.Printf("leave: create notification for leave %s: %v", leaveID, err)
	}

	// OS-level push is best-effort and deliberately detached.
	//
	// Unlike the announcement path, NotifyLeaveDecision runs SYNCHRONOUSLY
	// inside Approve/Reject. Calling FCM inline would put network latency on
	// the approval response and make approving fail when FCM is unreachable —
	// the in-app row above is the durable surface and must not depend on it.
	//
	// context.Background() rather than ctx: the request context is cancelled
	// the moment the handler returns, which would abort the push mid-flight.
	if n.push != nil {
		go n.sendDecisionPush(context.Background(), emp.UserID, leaveID, title, body)
	}
}

func (n *leaveNotifier) sendDecisionPush(ctx context.Context, userID, leaveID uuid.UUID, title, body string) {
	result, err := n.push.SendToUser(ctx, userID, dto.NotificationTestRequest{
		Title: title,
		Body:  body,
		// Mirrors the announcement payload so the app can deep-link from the
		// OS notification straight to the leave request detail screen.
		Data: map[string]any{"type": string(models.NotificationTypeLeaveRequest), "id": leaveID.String()},
	})
	if err != nil {
		log.Printf("leave: push for leave %s to user %s: %v", leaveID, userID, err)
		return
	}
	log.Printf("leave: push user=%s leave=%s sent=%d skipped=%d errors=%v",
		userID, leaveID, result.Sent, result.Skipped, result.Errors)
}

// leaveDecisionCopy renders the DR section 3 / AC-09 copy. Extracted so the
// exact wording is testable without a database.
func leaveDecisionCopy(approved bool, from, to time.Time) (title, body string) {
	verb := "rejected"
	title = "Leave Request Rejected"
	if approved {
		verb = "approved"
		title = "Leave Request Approved"
	}
	body = fmt.Sprintf(
		"Your leave request from %s to %s has been %s.",
		from.Format(leaveDecisionDateFormat),
		to.Format(leaveDecisionDateFormat),
		verb,
	)
	return title, body
}
