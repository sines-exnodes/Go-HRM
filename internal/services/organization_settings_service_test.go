package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Helpers ----

func newOrgSettingsSvc(t *testing.T) *services.OrganizationSettingsService {
	t.Helper()
	repo := repositories.NewSystemConfigRepository(testDB)
	// EnsureExists upfront — truncateAll wipes the sentinel.
	require.NoError(t, repo.EnsureExists(context.Background()))
	return services.NewOrganizationSettingsService(
		repo,
		repositories.NewEmployeeRepository(testDB),
	)
}

// ---- EnsureExists ----

func TestOrgSettings_EnsureExists_Idempotent(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	repo := repositories.NewSystemConfigRepository(testDB)
	require.NoError(t, repo.EnsureExists(ctx))
	require.NoError(t, repo.EnsureExists(ctx), "second EnsureExists must be a no-op")
	require.NoError(t, repo.EnsureExists(ctx))

	// Verify exactly one row, with the sentinel UUID.
	var count int64
	require.NoError(t, testDB.Model(&models.SystemConfig{}).Count(&count).Error)
	require.Equal(t, int64(1), count)

	cfg, err := repo.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, models.SystemConfigSingletonID, cfg.ID)
}

// ---- Attendance ----

func TestOrgSettings_GetAttendance_DefaultsAfterSeed(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)
	out, err := svc.GetAttendance(ctx)
	require.NoError(t, err)
	// DB defaults.
	assert.Equal(t, 9, out.LateThresholdHour)
	assert.Equal(t, 0, out.LateThresholdMinute)
	assert.Equal(t, 18, out.CheckoutThresholdHour)
	assert.Equal(t, 0, out.CheckoutThresholdMinute)
}

func TestOrgSettings_UpdateAttendance_PartialPatch(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)

	// Only late_threshold_hour supplied — others must stay defaults.
	newHour := 10
	out, err := svc.UpdateAttendance(ctx, dto.AttendanceSettingsUpdate{LateThresholdHour: &newHour})
	require.NoError(t, err)
	assert.Equal(t, 10, out.LateThresholdHour)
	assert.Equal(t, 0, out.LateThresholdMinute)
	assert.Equal(t, 18, out.CheckoutThresholdHour)
}

func TestOrgSettings_UpdateAttendance_EmptyPatchNoop(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)

	out, err := svc.UpdateAttendance(ctx, dto.AttendanceSettingsUpdate{})
	require.NoError(t, err, "empty patch must succeed")
	assert.Equal(t, 9, out.LateThresholdHour)
}

// ---- Company profile ----

func TestOrgSettings_GetCompanyProfile_NullByDefault(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)
	out, err := svc.GetCompanyProfile(ctx)
	require.NoError(t, err)
	assert.Nil(t, out.CompanyAddress)
	assert.Nil(t, out.CompanyLatitude)
	assert.Nil(t, out.CompanyLongitude)
	assert.Nil(t, out.CompanyAddressUpdatedAt)
	assert.Nil(t, out.CompanyAddressUpdatedBy)
	assert.Nil(t, out.UpdatedByName)
}

func TestOrgSettings_UpdateCompanyProfile_StampsUpdatedByAndAt(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)
	admin, adminEmp := makeEmpUser(t, "orgadmin@example.com", "Org Admin")

	addr := "1 HRM Lane"
	lat := 21.0285
	lng := 105.8542
	out, err := svc.UpdateCompanyProfile(ctx, admin.ID, dto.CompanyProfileUpdate{
		CompanyAddress:   &addr,
		CompanyLatitude:  &lat,
		CompanyLongitude: &lng,
	})
	require.NoError(t, err)
	require.NotNil(t, out.CompanyAddress)
	assert.Equal(t, "1 HRM Lane", *out.CompanyAddress)
	assert.InDelta(t, 21.0285, *out.CompanyLatitude, 0.0001)
	assert.InDelta(t, 105.8542, *out.CompanyLongitude, 0.0001)
	require.NotNil(t, out.CompanyAddressUpdatedAt, "updated_at must be stamped on any address change")
	require.NotNil(t, out.CompanyAddressUpdatedBy)
	assert.Equal(t, adminEmp.ID, *out.CompanyAddressUpdatedBy)
	require.NotNil(t, out.UpdatedByName)
	assert.Equal(t, "Org Admin", *out.UpdatedByName)
}

func TestOrgSettings_UpdateCompanyProfile_NoAddressFields_NoStamp(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)
	admin, _ := makeEmpUser(t, "orgadmin@example.com", "Org Admin")

	// Empty patch — no address-touching fields supplied. updated_at must
	// NOT be stamped (avoid spurious audit-trail churn).
	out, err := svc.UpdateCompanyProfile(ctx, admin.ID, dto.CompanyProfileUpdate{})
	require.NoError(t, err)
	assert.Nil(t, out.CompanyAddressUpdatedAt)
	assert.Nil(t, out.CompanyAddressUpdatedBy)
}

func TestOrgSettings_UpdateCompanyProfile_NoEmployeeProfile_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newOrgSettingsSvc(t)
	// User without an Employee row.
	u := makeUser(t, "noemp-org@example.com", "pw-Aa123456")

	addr := "x"
	_, err := svc.UpdateCompanyProfile(ctx, u.ID, dto.CompanyProfileUpdate{CompanyAddress: &addr})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No employee record")
}
