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

// OrganizationSettingsService owns the singleton system_config row.
// Permission gating happens upstream (RequirePerms(PermOrgSettings) on
// the admin routes; GET /company-profile is JWT-only by design).
type OrganizationSettingsService struct {
	repo repositories.SystemConfigRepository
	emps repositories.EmployeeRepository
}

// NewOrganizationSettingsService constructs the service. emps is used
// only by GET /company-profile to resolve the updated_by_name field;
// pass nil to skip that projection (the field stays nil).
func NewOrganizationSettingsService(
	repo repositories.SystemConfigRepository,
	emps repositories.EmployeeRepository,
) *OrganizationSettingsService {
	return &OrganizationSettingsService{repo: repo, emps: emps}
}

// EnsureExists is the seed-time entry point. Idempotent — safe to call
// on every boot.
func (s *OrganizationSettingsService) EnsureExists(ctx context.Context) error {
	return s.repo.EnsureExists(ctx)
}

// resolveCurrentEmployee mirrors the LeaveService / AttendanceService /
// AnnouncementService helper. Used by UpdateCompanyProfile to record
// `company_address_updated_by`.
func (s *OrganizationSettingsService) resolveCurrentEmployee(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	if s.emps == nil {
		return uuid.Nil, apperrors.ErrInternal("employee repo not configured")
	}
	emp, err := s.emps.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.ErrForbidden("No employee record for current user")
		}
		return uuid.Nil, err
	}
	return emp.ID, nil
}

// loadCfg returns the singleton row, or 500 if seed didn't run.
func (s *OrganizationSettingsService) loadCfg(ctx context.Context) (*models.SystemConfig, error) {
	cfg, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, apperrors.ErrInternal("system_config singleton missing — seed did not run")
	}
	return cfg, nil
}

// ---- Attendance ----

// GetAttendance returns the four threshold fields.
func (s *OrganizationSettingsService) GetAttendance(ctx context.Context) (*dto.AttendanceSettingsRead, error) {
	cfg, err := s.loadCfg(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.AttendanceSettingsRead{
		LateThresholdHour:       int(cfg.LateThresholdHour),
		LateThresholdMinute:     int(cfg.LateThresholdMinute),
		CheckoutThresholdHour:   int(cfg.CheckoutThresholdHour),
		CheckoutThresholdMinute: int(cfg.CheckoutThresholdMinute),
	}, nil
}

// UpdateAttendance applies a partial patch and returns the new state.
// Empty patch (every pointer nil) is allowed — returns the current row
// untouched.
func (s *OrganizationSettingsService) UpdateAttendance(ctx context.Context, in dto.AttendanceSettingsUpdate) (*dto.AttendanceSettingsRead, error) {
	fields := map[string]any{}
	if in.LateThresholdHour != nil {
		fields["late_threshold_hour"] = *in.LateThresholdHour
	}
	if in.LateThresholdMinute != nil {
		fields["late_threshold_minute"] = *in.LateThresholdMinute
	}
	if in.CheckoutThresholdHour != nil {
		fields["checkout_threshold_hour"] = *in.CheckoutThresholdHour
	}
	if in.CheckoutThresholdMinute != nil {
		fields["checkout_threshold_minute"] = *in.CheckoutThresholdMinute
	}
	if err := s.repo.UpdateFields(ctx, fields); err != nil {
		return nil, err
	}
	return s.GetAttendance(ctx)
}

// ---- Company profile ----

// GetCompanyProfile returns the address + lat/lng plus a best-effort
// resolution of the updater's full name.
func (s *OrganizationSettingsService) GetCompanyProfile(ctx context.Context) (*dto.CompanyProfileRead, error) {
	cfg, err := s.loadCfg(ctx)
	if err != nil {
		return nil, err
	}
	out := &dto.CompanyProfileRead{
		CompanyAddress:          cfg.CompanyAddress,
		CompanyLatitude:         cfg.CompanyLatitude,
		CompanyLongitude:        cfg.CompanyLongitude,
		CompanyAddressUpdatedAt: cfg.CompanyAddressUpdatedAt,
		CompanyAddressUpdatedBy: cfg.CompanyAddressUpdatedBy,
	}
	if cfg.CompanyAddressUpdatedBy != nil && s.emps != nil {
		emp, err := s.emps.FindByID(ctx, *cfg.CompanyAddressUpdatedBy)
		if err == nil && emp != nil {
			name := emp.FullName
			out.UpdatedByName = &name
		}
	}
	return out, nil
}

// UpdateCompanyProfile patches the address + lat/lng + audit columns.
// Stamps company_address_updated_at = NOW() and updates the
// updated_by employee FK whenever ANY of the three address fields are
// supplied (matches the Python contract).
func (s *OrganizationSettingsService) UpdateCompanyProfile(ctx context.Context, currentUserID uuid.UUID, in dto.CompanyProfileUpdate) (*dto.CompanyProfileRead, error) {
	fields := map[string]any{}
	dirty := false
	if in.CompanyAddress != nil {
		fields["company_address"] = *in.CompanyAddress
		dirty = true
	}
	if in.CompanyLatitude != nil {
		fields["company_latitude"] = *in.CompanyLatitude
		dirty = true
	}
	if in.CompanyLongitude != nil {
		fields["company_longitude"] = *in.CompanyLongitude
		dirty = true
	}
	if dirty {
		empID, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		fields["company_address_updated_at"] = time.Now().UTC()
		fields["company_address_updated_by"] = empID
	}
	if err := s.repo.UpdateFields(ctx, fields); err != nil {
		return nil, err
	}
	return s.GetCompanyProfile(ctx)
}
