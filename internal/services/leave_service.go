package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// ---- Attachment upload constants ----

const (
	leaveAttachmentSubdir   = "leave-attachments"
	leaveAttachmentMaxBytes = 10 * 1024 * 1024 // 10 MB
)

// allowedAttachmentMIME is the set of content types accepted for a leave
// attachment (images + PDF — matches the Python source). The authoritative
// check sniffs the file bytes via http.DetectContentType — the client's
// Content-Type header is treated as a hint only (review-fix #2 pattern,
// already battle-tested in Phase 2 avatars and Phase 4 skill icons).
var allowedAttachmentMIME = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// AttachmentUpload bundles the multipart attachment fields. Nil means
// "no attachment supplied in this request" (Update preserves the existing
// URL on the row).
type AttachmentUpload struct {
	Content     []byte
	ContentType string // client-supplied hint
	Ext         string // ".pdf" / ".png" / etc — lowercased by caller
}

// ---- Service ----

// LeaveService owns the leave_requests aggregate and the cross-aggregate
// reads (employee profile + department + position projections, quota
// lookup). Permission resolution lives in the handler (matches
// DependentService's owner-or-admin shape): the handler precomputes an
// `asAdmin` bool from the user's effective permission set and passes it
// down. The service enforces ownership + status invariants.
type LeaveService struct {
	leaves  repositories.LeaveRequestRepository
	emps    repositories.EmployeeRepository
	depts   repositories.DepartmentRepository
	pos     repositories.PositionRepository
	quota   repositories.LeaveQuotaRepository
	uploads Uploader // optional; nil means attachment upload is unavailable
}

// NewLeaveService constructs a LeaveService. Pass nil for `uploads` if the
// storage backend is unconfigured — attachment endpoints will then return
// a 500 with a clear message but the rest of the API stays usable.
func NewLeaveService(
	leaves repositories.LeaveRequestRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
) *LeaveService {
	return &LeaveService{
		leaves:  leaves,
		emps:    emps,
		depts:   depts,
		pos:     pos,
		quota:   quota,
		uploads: uploads,
	}
}

// ---- Pure helpers ----

// truncateToDate strips the time component and pins the timestamp to UTC.
// Postgres DATE columns are date-only; using a midnight-UTC timestamp
// avoids off-by-one bugs when comparing across time zones.
func truncateToDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// calculateTotalDays mirrors the Python contract: inclusive calendar day
// count (no business-day skip), multiplied by 0.5 when the period is a
// half-day variant. Callers must have already validated the date range
// (to >= from) and the half-day single-day rule (from == to when half).
func calculateTotalDays(from, to time.Time, period models.LeavePeriod) float64 {
	from = truncateToDate(from)
	to = truncateToDate(to)
	days := float64(int(to.Sub(from).Hours()/24)) + 1
	if models.IsHalfDayPeriod(period) {
		return days * 0.5
	}
	return days
}

// validateDateInputs enforces the date-range invariants common to Create
// and Update: to >= from, and half-day periods must span a single day.
// Returns the computed total_days on success.
func validateDateInputs(from, to time.Time, period models.LeavePeriod) (float64, error) {
	from = truncateToDate(from)
	to = truncateToDate(to)
	if to.Before(from) {
		return 0, apperrors.ErrBadRequest("To Date must be on or after From Date")
	}
	if models.IsHalfDayPeriod(period) && !from.Equal(to) {
		return 0, apperrors.ErrBadRequest("Half-day leave must be a single day (from_date must equal to_date)")
	}
	return calculateTotalDays(from, to, period), nil
}

// validateLeaveType checks the enum domain explicitly. The DB has the
// same CHECK constraint, but failing here yields a 400 with a clear
// message instead of a 500-on-insert.
func validateLeaveType(t models.LeaveType) error {
	switch t {
	case models.LeaveTypeAnnual, models.LeaveTypeSick, models.LeaveTypePersonal,
		models.LeaveTypeMaternity, models.LeaveTypeUnpaid:
		return nil
	}
	return apperrors.ErrBadRequest(fmt.Sprintf("invalid leave_type: %q", string(t)))
}

// validateLeavePeriod checks the enum domain explicitly.
func validateLeavePeriod(p models.LeavePeriod) error {
	switch p {
	case models.LeavePeriodFullDay, models.LeavePeriodMorningHalf, models.LeavePeriodAfternoonHalf:
		return nil
	}
	return apperrors.ErrBadRequest(fmt.Sprintf("invalid leave_period: %q", string(p)))
}

// ---- Attachment upload ----

// uploadAttachment validates and stores a leave attachment, returning the
// public URL. Content type is determined by sniffing the actual bytes,
// NOT trusting the client header (review-fix #2 pattern). Allowed MIME:
// PDF + common image types.
func (s *LeaveService) uploadAttachment(ctx context.Context, att AttachmentUpload) (string, error) {
	if s.uploads == nil {
		return "", apperrors.ErrInternal("Storage is not configured; cannot upload attachment")
	}
	if len(att.Content) == 0 {
		return "", apperrors.ErrBadRequest("Attachment file is empty")
	}
	if len(att.Content) > leaveAttachmentMaxBytes {
		return "", apperrors.ErrBadRequest("Attachment must not exceed 10MB")
	}
	sniffLen := len(att.Content)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffed := http.DetectContentType(att.Content[:sniffLen])
	if !allowedAttachmentMIME[sniffed] {
		return "", apperrors.ErrBadRequest("Attachment must be a PDF or image (PNG, JPEG, GIF, WEBP)")
	}
	url, err := s.uploads.Upload(ctx, leaveAttachmentSubdir, att.Ext, att.Content, sniffed)
	if err != nil {
		return "", err
	}
	return url, nil
}

// ---- Read projection ----

// populateRead is the per-row builder shared by Get/List/Dashboard/History.
// Inflates employee/department/position into the LeaveRefRead projection
// when the referenced rows are live. Missing refs are silently omitted —
// the FE renders the row without the embedded sub-objects.
//
// NOTE on perf: this is N+1 for list paths. The current pattern matches
// Phase 4's ListForEmployee and is acceptable at the typical page sizes
// (<= 100). A future optimization would batch-load employees/departments
// /positions by ID set. Flagged for the post-Phase-5 review.
func (s *LeaveService) populateRead(ctx context.Context, lr *models.LeaveRequest) (dto.LeaveRequestRead, error) {
	out := dto.LeaveRequestRead{
		ID:            lr.ID.String(),
		FromDate:      lr.FromDate,
		ToDate:        lr.ToDate,
		LeavePeriod:   lr.LeavePeriod,
		LeaveType:     lr.LeaveType,
		TotalDays:     lr.TotalDays,
		Reason:        lr.Reason,
		AttachmentURL: lr.AttachmentURL,
		Status:        lr.Status,
		CreatedBy:     lr.CreatedBy.String(),
		CreatedAt:     lr.CreatedAt,
		UpdatedAt:     lr.UpdatedAt,
	}
	emp, err := s.emps.FindByID(ctx, lr.EmployeeID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return out, err
		}
		// Soft-deleted/missing employee: leave the embedded refs empty.
		return out, nil
	}
	out.Employee = &dto.LeaveRefRead{ID: emp.ID.String(), Name: emp.FullName}
	if emp.DepartmentID != nil {
		if d, derr := s.depts.FindByID(ctx, *emp.DepartmentID, false); derr == nil && d != nil {
			out.Department = &dto.LeaveRefRead{ID: d.ID.String(), Name: d.Name}
		}
	}
	if emp.PositionID != nil {
		if p, perr := s.pos.FindByID(ctx, *emp.PositionID); perr == nil && p != nil {
			out.Position = &dto.LeaveRefRead{ID: p.ID.String(), Name: p.Name}
		}
	}
	return out, nil
}

// populateReadList is the bulk wrapper. Build-error on a single row
// short-circuits the whole list (a DB error during populate is a real
// problem). Refs that 404 silently drop to nil per populateRead.
func (s *LeaveService) populateReadList(ctx context.Context, rows []models.LeaveRequest) ([]dto.LeaveRequestRead, error) {
	out := make([]dto.LeaveRequestRead, 0, len(rows))
	for i := range rows {
		r, err := s.populateRead(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// ---- Balance ----

// computeBalance reads the employee's stored quota and subtracts the
// approved-leave aggregate for the calendar year. Missing quota row
// (admin never set one) yields zero quotas, which is the safe default.
func (s *LeaveService) computeBalance(ctx context.Context, employeeID uuid.UUID, year int) (dto.LeaveBalanceSummary, error) {
	grouped, err := s.leaves.SumApprovedDays(ctx, employeeID, year)
	if err != nil {
		return dto.LeaveBalanceSummary{}, err
	}
	var (
		annualQuota float64
		sickQuota   float64
	)
	q, err := s.quota.GetByEmployee(ctx, employeeID)
	if err != nil {
		return dto.LeaveBalanceSummary{}, err
	}
	if q != nil {
		annualQuota = q.AnnualLeaveQuota
		sickQuota = q.SickLeaveQuota
	}

	annualUsed := grouped[models.LeaveTypeAnnual].Days
	sickUsed := grouped[models.LeaveTypeSick].Days
	var totalCount int64
	for _, v := range grouped {
		totalCount += v.Count
	}
	return dto.LeaveBalanceSummary{
		Year:            year,
		AnnualQuota:     annualQuota,
		AnnualUsed:      annualUsed,
		AnnualRemaining: annualQuota - annualUsed,
		SickQuota:       sickQuota,
		SickUsed:        sickUsed,
		SickRemaining:   sickQuota - sickUsed,
		LeavesThisYear:  int(totalCount),
	}, nil
}

// ---- Warning computation ----

// computeWarnings builds the non-blocking warnings list for Create/Update.
// Insufficient quota only triggers for annual/sick (the quota types).
// Overlap inspects pending+approved rows for the same employee, excluding
// the current row on Update.
func (s *LeaveService) computeWarnings(ctx context.Context, employeeID uuid.UUID, leaveType models.LeaveType, from, to time.Time, totalDays float64, excludeID *uuid.UUID) ([]string, error) {
	warnings := []string{}

	if models.IsQuotaLeaveType(leaveType) {
		balance, err := s.computeBalance(ctx, employeeID, from.UTC().Year())
		if err != nil {
			return nil, err
		}
		remaining, ok := remainingForType(balance, leaveType)
		if ok && remaining < totalDays {
			warnings = append(warnings, fmt.Sprintf(
				"Insufficient %s leave balance: requested %.1f day(s), remaining %.1f day(s)",
				string(leaveType), totalDays, remaining,
			))
		}
	}

	overlaps, err := s.leaves.Overlapping(ctx, employeeID, from, to, excludeID)
	if err != nil {
		return nil, err
	}
	if len(overlaps) > 0 {
		ranges := make([]string, 0, len(overlaps))
		for _, o := range overlaps {
			ranges = append(ranges, fmt.Sprintf("%s..%s (%s)",
				o.FromDate.Format("2006-01-02"), o.ToDate.Format("2006-01-02"), string(o.Status),
			))
		}
		warnings = append(warnings, fmt.Sprintf(
			"Date range overlaps existing leave request(s): %s", strings.Join(ranges, ", "),
		))
	}

	return warnings, nil
}

// remainingForType maps the leave type to the matching quota remainder.
// Returns ok=false for non-quota types (personal/maternity/unpaid),
// signalling the caller to skip the insufficient-quota warning entirely.
func remainingForType(b dto.LeaveBalanceSummary, t models.LeaveType) (float64, bool) {
	switch t {
	case models.LeaveTypeAnnual:
		return b.AnnualRemaining, true
	case models.LeaveTypeSick:
		return b.SickRemaining, true
	default:
		return 0, false
	}
}

// ---- Authorization helpers ----

// resolveCurrentEmployee returns the employee row for the authenticated
// user. Missing employee record (user with no HR profile) yields a 403
// — only employees can act on leave requests.
func (s *LeaveService) resolveCurrentEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
	emp, err := s.emps.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden("No employee record for current user")
		}
		return nil, err
	}
	return emp, nil
}

// ---- Create ----

// Create inserts a new leave_request. EmployeeID in the DTO is the
// subject — admin-only when it differs from the current employee.
// Returns the created row plus any non-blocking warnings (insufficient
// quota, date overlap). The DB row is created regardless of warnings.
func (s *LeaveService) Create(ctx context.Context, currentUserID uuid.UUID, asAdmin bool, in dto.LeaveRequestCreate, attachment *AttachmentUpload) (*dto.LeaveRequestWriteResult, error) {
	if err := validateLeaveType(in.LeaveType); err != nil {
		return nil, err
	}
	if err := validateLeavePeriod(in.LeavePeriod); err != nil {
		return nil, err
	}
	totalDays, err := validateDateInputs(in.FromDate, in.ToDate, in.LeavePeriod)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Reason) == "" {
		return nil, apperrors.ErrBadRequest("Reason cannot be blank")
	}

	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	// Resolve the subject (employee_id on the row).
	subject := currentEmp.ID
	if in.EmployeeID != nil && *in.EmployeeID != uuid.Nil && *in.EmployeeID != currentEmp.ID {
		if !asAdmin {
			return nil, apperrors.ErrForbidden("Only admins can create a leave request on behalf of another employee")
		}
		// Verify the target employee exists and is live.
		if _, ferr := s.emps.FindByID(ctx, *in.EmployeeID); ferr != nil {
			if errors.Is(ferr, gorm.ErrRecordNotFound) {
				return nil, apperrors.ErrNotFound("Employee")
			}
			return nil, ferr
		}
		subject = *in.EmployeeID
	}

	warnings, err := s.computeWarnings(ctx, subject, in.LeaveType, in.FromDate, in.ToDate, totalDays, nil)
	if err != nil {
		return nil, err
	}

	// Upload attachment BEFORE the DB write so a Postgres failure can't
	// leave an orphaned object (matches the skill-icon pattern).
	var attachmentURL *string
	if attachment != nil {
		url, uerr := s.uploadAttachment(ctx, *attachment)
		if uerr != nil {
			return nil, uerr
		}
		attachmentURL = &url
	}

	row := &models.LeaveRequest{
		EmployeeID:    subject,
		FromDate:      truncateToDate(in.FromDate),
		ToDate:        truncateToDate(in.ToDate),
		LeavePeriod:   in.LeavePeriod,
		LeaveType:     in.LeaveType,
		TotalDays:     totalDays,
		Reason:        strings.TrimSpace(in.Reason),
		AttachmentURL: attachmentURL,
		Status:        models.LeaveStatusPending,
		CreatedBy:     currentEmp.ID,
	}
	if err := s.leaves.Create(ctx, row); err != nil {
		if attachmentURL != nil && s.uploads != nil {
			_ = s.uploads.Delete(ctx, *attachmentURL)
		}
		return nil, err
	}

	read, err := s.populateRead(ctx, row)
	if err != nil {
		return nil, err
	}
	return &dto.LeaveRequestWriteResult{Request: read, Warnings: warnings}, nil
}

// ---- Update / state transitions / delete ----

// Update patches an existing leave_request. Status transitions:
//   - rejected/cancelled  -> 409 (terminal, no edits)
//   - approved (admin)    -> reverted to pending after the patch
//   - approved (owner)    -> 403 (owner can only edit pending)
//   - pending             -> stays pending
func (s *LeaveService) Update(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool, in dto.LeaveRequestUpdate, attachment *AttachmentUpload) (*dto.LeaveRequestWriteResult, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}

	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	isOwner := row.EmployeeID == currentEmp.ID
	if !asAdmin && !isOwner {
		return nil, apperrors.ErrForbidden("You do not own this leave request")
	}

	switch row.Status {
	case models.LeaveStatusRejected, models.LeaveStatusCancelled:
		return nil, apperrors.ErrConflict(fmt.Sprintf("Cannot edit a %s leave request", string(row.Status)))
	case models.LeaveStatusApproved:
		if !asAdmin {
			return nil, apperrors.ErrForbidden("Only an admin can edit an approved leave request")
		}
	}

	// Apply patches. Pointer types preserve "not provided" semantics.
	if in.FromDate != nil {
		row.FromDate = *in.FromDate
	}
	if in.ToDate != nil {
		row.ToDate = *in.ToDate
	}
	if in.LeavePeriod != nil {
		if err := validateLeavePeriod(*in.LeavePeriod); err != nil {
			return nil, err
		}
		row.LeavePeriod = *in.LeavePeriod
	}
	if in.LeaveType != nil {
		if err := validateLeaveType(*in.LeaveType); err != nil {
			return nil, err
		}
		row.LeaveType = *in.LeaveType
	}
	if in.Reason != nil {
		r := strings.TrimSpace(*in.Reason)
		if r == "" {
			return nil, apperrors.ErrBadRequest("Reason cannot be blank")
		}
		row.Reason = r
	}

	// Recompute total_days whenever dates or period might have changed.
	// Cheaper to recompute unconditionally than to track a dirty flag.
	totalDays, err := validateDateInputs(row.FromDate, row.ToDate, row.LeavePeriod)
	if err != nil {
		return nil, err
	}
	row.TotalDays = totalDays
	row.FromDate = truncateToDate(row.FromDate)
	row.ToDate = truncateToDate(row.ToDate)

	// Status: approved + admin patch -> revert to pending (Python contract).
	if row.Status == models.LeaveStatusApproved {
		row.Status = models.LeaveStatusPending
	}

	warnings, err := s.computeWarnings(ctx, row.EmployeeID, row.LeaveType, row.FromDate, row.ToDate, row.TotalDays, &row.ID)
	if err != nil {
		return nil, err
	}

	// Attachment swap (upload new BEFORE writing the row, best-effort
	// delete old AFTER a successful write — mirror of skill-icon Update).
	prevAttachment := row.AttachmentURL
	if attachment != nil {
		url, uerr := s.uploadAttachment(ctx, *attachment)
		if uerr != nil {
			return nil, uerr
		}
		row.AttachmentURL = &url
	}

	if err := s.leaves.Update(ctx, row); err != nil {
		if attachment != nil && row.AttachmentURL != nil && s.uploads != nil {
			_ = s.uploads.Delete(ctx, *row.AttachmentURL)
		}
		return nil, err
	}
	if attachment != nil && prevAttachment != nil && *prevAttachment != "" && s.uploads != nil {
		_ = s.uploads.Delete(ctx, *prevAttachment)
	}

	read, err := s.populateRead(ctx, row)
	if err != nil {
		return nil, err
	}
	return &dto.LeaveRequestWriteResult{Request: read, Warnings: warnings}, nil
}

// Approve transitions a pending request to approved. Permission gate is
// applied upstream (RequirePerms(PermLeaveApprove)).
func (s *LeaveService) Approve(ctx context.Context, id uuid.UUID) (*dto.LeaveRequestRead, error) {
	return s.transitionStatus(ctx, id, models.LeaveStatusApproved, []models.LeaveStatus{models.LeaveStatusPending})
}

// Reject transitions a pending request to rejected. Permission gate is
// applied upstream (RequirePerms(PermLeaveApprove) — Python uses the
// same permission for approve and reject).
func (s *LeaveService) Reject(ctx context.Context, id uuid.UUID) (*dto.LeaveRequestRead, error) {
	return s.transitionStatus(ctx, id, models.LeaveStatusRejected, []models.LeaveStatus{models.LeaveStatusPending})
}

// Cancel transitions pending or approved to cancelled. Owner can cancel
// their own; admin can cancel anyone's. Permission gate is applied
// upstream (RequirePerms(PermLeaveCancel)); ownership is enforced here.
func (s *LeaveService) Cancel(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.LeaveRequestRead, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}
	currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if !asAdmin && row.EmployeeID != currentEmp.ID {
		return nil, apperrors.ErrForbidden("You do not own this leave request")
	}
	return s.transitionStatus(ctx, id, models.LeaveStatusCancelled, []models.LeaveStatus{models.LeaveStatusPending, models.LeaveStatusApproved})
}

// transitionStatus is the shared finalizer for Approve/Reject/Cancel.
// Re-fetches inside the call so a concurrent update is reflected in the
// 409 message (rather than blindly overwriting).
func (s *LeaveService) transitionStatus(ctx context.Context, id uuid.UUID, to models.LeaveStatus, fromAllowed []models.LeaveStatus) (*dto.LeaveRequestRead, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}
	allowed := false
	for _, s := range fromAllowed {
		if row.Status == s {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, apperrors.ErrConflict(fmt.Sprintf(
			"Cannot transition leave request from %s to %s", string(row.Status), string(to),
		))
	}
	row.Status = to
	if err := s.leaves.Update(ctx, row); err != nil {
		return nil, err
	}
	read, err := s.populateRead(ctx, row)
	if err != nil {
		return nil, err
	}
	return &read, nil
}

// Delete soft-deletes a leave request.
//   - Admin (asAdmin): may delete any status.
//   - Non-admin owner: only `pending` and only their own.
//   - Anyone else: 403.
func (s *LeaveService) Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) error {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Leave request")
		}
		return err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return err
		}
		if row.EmployeeID != currentEmp.ID {
			return apperrors.ErrForbidden("You do not own this leave request")
		}
		if row.Status != models.LeaveStatusPending {
			return apperrors.ErrForbidden("Only pending leave requests can be deleted by their owner")
		}
	}
	return s.leaves.SoftDelete(ctx, id)
}

// ---- Read endpoints ----

// Get returns a single leave request. Owner or admin only; everyone else
// gets a 403. The PermLeaveRead gate is applied upstream.
func (s *LeaveService) Get(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.LeaveRequestRead, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}
	if !asAdmin {
		currentEmp, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		if row.EmployeeID != currentEmp.ID {
			return nil, apperrors.ErrForbidden("You do not own this leave request")
		}
	}
	read, err := s.populateRead(ctx, row)
	if err != nil {
		return nil, err
	}
	return &read, nil
}

// List returns paginated leave requests for the admin/manager listing
// page. Search/department/position filters are resolved by enumerating
// matching employees first, then passing the employee_id allowlist to
// the leave repo. The PermLeaveRead gate is applied upstream.
func (s *LeaveService) List(ctx context.Context, q dto.LeaveListQuery) (dto.PaginatedData[dto.LeaveRequestRead], error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 10
	}

	// Resolve the employee-ID allowlist when any employee-side filter is set.
	var employeeIDs []uuid.UUID
	needFilter := strings.TrimSpace(q.Search) != "" || q.DepartmentID != "" || q.PositionID != ""
	if needFilter {
		empQuery := dto.EmployeeListQuery{
			Page:     1,
			PageSize: 1000, // upper bound — phase-5 listing pages are admin-facing
			Search:   q.Search,
		}
		if q.DepartmentID != "" {
			deptID, perr := uuid.Parse(q.DepartmentID)
			if perr != nil {
				return dto.PaginatedData[dto.LeaveRequestRead]{}, apperrors.ErrBadRequest("invalid department_id")
			}
			empQuery.DepartmentID = &deptID
		}
		if q.PositionID != "" {
			posID, perr := uuid.Parse(q.PositionID)
			if perr != nil {
				return dto.PaginatedData[dto.LeaveRequestRead]{}, apperrors.ErrBadRequest("invalid position_id")
			}
			empQuery.PositionID = &posID
		}
		emps, _, eerr := s.emps.List(ctx, empQuery)
		if eerr != nil {
			return dto.PaginatedData[dto.LeaveRequestRead]{}, eerr
		}
		employeeIDs = make([]uuid.UUID, 0, len(emps))
		for _, e := range emps {
			employeeIDs = append(employeeIDs, e.ID)
		}
		// Empty allowlist => zero results (the repo short-circuits on len==0).
		if len(employeeIDs) == 0 {
			return dto.PaginatedData[dto.LeaveRequestRead]{
				Items: []dto.LeaveRequestRead{}, Total: 0, Page: page, PageSize: size, TotalPages: 0,
			}, nil
		}
	}

	rows, total, err := s.leaves.List(ctx, employeeIDs, q.Status, page, size)
	if err != nil {
		return dto.PaginatedData[dto.LeaveRequestRead]{}, err
	}
	items, err := s.populateReadList(ctx, rows)
	if err != nil {
		return dto.PaginatedData[dto.LeaveRequestRead]{}, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return dto.PaginatedData[dto.LeaveRequestRead]{
		Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
	}, nil
}

// GetBalance returns the quota summary for the given employee + year.
// year <= 0 defaults to the current calendar year (UTC).
func (s *LeaveService) GetBalance(ctx context.Context, employeeID uuid.UUID, year int) (*dto.LeaveBalanceSummary, error) {
	if _, err := s.emps.FindByID(ctx, employeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Employee")
		}
		return nil, err
	}
	if year <= 0 {
		year = time.Now().UTC().Year()
	}
	b, err := s.computeBalance(ctx, employeeID, year)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetMyDashboard composes balance + upcoming + history for the current
// employee. The endpoint is JWT-only (no extra perm gate).
func (s *LeaveService) GetMyDashboard(ctx context.Context, currentUserID uuid.UUID) (*dto.LeaveDashboardRead, error) {
	emp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	year := time.Now().UTC().Year()
	balance, err := s.computeBalance(ctx, emp.ID, year)
	if err != nil {
		return nil, err
	}
	today := truncateToDate(time.Now().UTC())
	upcomingRows, err := s.leaves.Upcoming(ctx, emp.ID, today, 5)
	if err != nil {
		return nil, err
	}
	historyRows, err := s.leaves.History(ctx, emp.ID, today, 5)
	if err != nil {
		return nil, err
	}
	upcoming, err := s.populateReadList(ctx, upcomingRows)
	if err != nil {
		return nil, err
	}
	history, err := s.populateReadList(ctx, historyRows)
	if err != nil {
		return nil, err
	}
	return &dto.LeaveDashboardRead{Balance: balance, Upcoming: upcoming, History: history}, nil
}

// ListMyHistory returns the paginated /history/me listing: rows where
// to_date is past OR status is terminal (rejected/cancelled).
func (s *LeaveService) ListMyHistory(ctx context.Context, currentUserID uuid.UUID, q dto.LeaveHistoryQuery) (dto.PaginatedData[dto.LeaveRequestRead], error) {
	emp, err := s.resolveCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return dto.PaginatedData[dto.LeaveRequestRead]{}, err
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 10
	}
	rows, total, err := s.leaves.ListByEmployee(ctx, emp.ID, repositories.ListByEmployeeFilter{
		Statuses:  q.Status,
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		Page:      page,
		PageSize:  size,
	})
	if err != nil {
		return dto.PaginatedData[dto.LeaveRequestRead]{}, err
	}
	items, err := s.populateReadList(ctx, rows)
	if err != nil {
		return dto.PaginatedData[dto.LeaveRequestRead]{}, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return dto.PaginatedData[dto.LeaveRequestRead]{
		Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
	}, nil
}
