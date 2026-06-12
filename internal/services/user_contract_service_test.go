package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// newContractSvc builds a UserContractService pointing at the test DB.
func newContractSvc(t *testing.T) *services.UserContractService {
	t.Helper()
	return services.NewUserContractService(
		repositories.NewUserContractRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		nil, // uploads nil — tests attachment upload separately
	)
}

// dateUTC returns a UTC midnight time for the given date.
func dateUTC(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestUserContract_Create_FixedTerm(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c1@example.com", "Alice Smith")
	expiry := dateUTC(2027, 6, 30)
	out, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 7, 1),
		ExpiryDate:   &expiry,
		IsEndless:    false,
	})
	require.Nil(t, err)
	assert.Equal(t, "labour_contract", out.ContractType)
	assert.False(t, out.IsEndless)
	require.NotNil(t, out.ExpiryDate)
}

func TestUserContract_Create_Endless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c2@example.com", "Bob Jones")
	expiry := dateUTC(2027, 1, 1) // should be cleared
	out, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
		IsEndless:    true,
	})
	require.Nil(t, err)
	assert.True(t, out.IsEndless)
	assert.Nil(t, out.ExpiryDate, "endless contract must have nil expiry in DB")
}

func TestUserContract_Create_ExpiryBeforeSigned(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c3@example.com", "Carol White")
	expiry := dateUTC(2026, 5, 31)
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 6, 1),
		ExpiryDate:   &expiry,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.HTTP)
	assert.Contains(t, err.Message, "after signed date")
}

func TestUserContract_Create_ExpiryEqualsSigned(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c4@example.com", "Dan Brown")
	same := dateUTC(2026, 6, 1)
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   same,
		ExpiryDate:   &same,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.HTTP)
}

func TestUserContract_Create_MissingExpiry_NotEndless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c5@example.com", "Eva Green")
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   nil,
		IsEndless:    false,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.HTTP)
	assert.Contains(t, err.Message, "expiry date is required")
}

func TestUserContract_Get_WrongEmployee(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	owner, _ := makeEmpUser(t, "owner@example.com", "Owner User")
	other, _ := makeEmpUser(t, "other@example.com", "Other User")
	expiry := dateUTC(2027, 1, 1)
	created, aerr := svc.Create(context.Background(), owner.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	_, aerr = svc.Get(context.Background(), other.ID, created.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.HTTP)
}

func TestUserContract_Delete_SoftDeletes(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "del@example.com", "Delete Me")
	expiry := dateUTC(2027, 6, 30)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	aerr = svc.Delete(context.Background(), u.ID, created.ID)
	require.Nil(t, aerr)

	_, aerr = svc.Get(context.Background(), u.ID, created.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.HTTP)

	var count int64
	testDB.Raw("SELECT COUNT(*) FROM user_contracts WHERE id = ? AND is_deleted = true", created.ID).Scan(&count)
	assert.Equal(t, int64(1), count)
}

// ---- List + Filters ----

func TestUserContract_List_FilterSignedDate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list1@example.com", "List User")
	expiry := dateUTC(2027, 12, 31)

	_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 6, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	_, aerr = svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2024, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	from := dateUTC(2026, 1, 1)
	to := dateUTC(2026, 12, 31)
	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{
		Page: 1, PageSize: 10,
		SignedFrom: &from,
		SignedTo:   &to,
	})
	require.Nil(t, aerr)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Items, 1)
}

func TestUserContract_List_ExpiryFilter_ExcludesEndless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list2@example.com", "List User 2")

	_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		IsEndless:    true,
	})
	require.Nil(t, aerr)

	expiry := dateUTC(2027, 6, 30)
	_, aerr = svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	from := dateUTC(2027, 1, 1)
	to := dateUTC(2027, 12, 31)
	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{
		Page: 1, PageSize: 10,
		ExpiryFrom: &from,
		ExpiryTo:   &to,
	})
	require.Nil(t, aerr)
	assert.Equal(t, int64(1), result.Total, "endless contract must not appear in expiry filter results")
}

func TestUserContract_List_Pagination(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list3@example.com", "Paginated User")
	for i := 0; i < 3; i++ {
		expiry := dateUTC(2027+i, 12, 31)
		_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
			ContractType: "labour_contract",
			SignedDate:   dateUTC(2026+i, 1, 1),
			ExpiryDate:   &expiry,
		})
		require.Nil(t, aerr)
	}

	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{Page: 1, PageSize: 2})
	require.Nil(t, aerr)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.TotalPages)
}

// ---- Update ----

func TestUserContract_Update_RemoveAttachment(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd1@example.com", "Update User")
	url := "https://example.com/file.pdf"
	expiry := dateUTC(2027, 12, 31)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType:  "labour_contract",
		SignedDate:    dateUTC(2026, 1, 1),
		ExpiryDate:    &expiry,
		AttachmentURL: &url,
	})
	require.Nil(t, aerr)

	empty := ""
	updated, aerr := svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		AttachmentURL: &empty,
	})
	require.Nil(t, aerr)
	assert.Nil(t, updated.AttachmentURL)
}

func TestUserContract_Update_EndlessToggleOn(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd2@example.com", "Endless Toggle")
	expiry := dateUTC(2027, 6, 30)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)
	assert.False(t, created.IsEndless)

	endless := true
	updated, aerr := svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		IsEndless: &endless,
	})
	require.Nil(t, aerr)
	assert.True(t, updated.IsEndless)
	assert.Nil(t, updated.ExpiryDate, "turning endless ON must clear the expiry date")
}

func TestUserContract_Update_EndlessToggleOff_RequiresExpiry(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd3@example.com", "Endless Off")
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		IsEndless:    true,
	})
	require.Nil(t, aerr)

	notEndless := false
	_, aerr = svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		IsEndless:  &notEndless,
		ExpiryDate: nil,
	})
	require.NotNil(t, aerr)
	assert.Equal(t, 400, aerr.HTTP)
	assert.Contains(t, aerr.Message, "expiry date is required")
}
