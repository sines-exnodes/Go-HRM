package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---------------------------------------------------------------------------
// Salary/banking field-level access control (PR B — audit decision #6).
// These are pure functions (no DB), so they run without TEST_DATABASE_URL.
// ---------------------------------------------------------------------------

func spv(s string) *string   { return &s }
func fpv(f float64) *float64 { return &f }

func sampleEmployeeRead() *dto.EmployeeRead {
	return &dto.EmployeeRead{
		BasicSalary:     fpv(1000),
		InsuranceSalary: fpv(800),
		BankAccount:     spv("123456789012"),
		BankName:        spv("ACB"),
		BankHolderName:  spv("Jane Doe"),
		PaymentMethod:   spv("bank_transfer"),
	}
}

func TestFieldVisibility_StripsBothWhenNoPerms(t *testing.T) {
	v := sampleEmployeeRead()
	services.ApplyEmployeeFieldVisibility(v, services.EmployeeFieldPerms{}, false)
	assert.Nil(t, v.BasicSalary, "salary hidden without salary_view")
	assert.Nil(t, v.InsuranceSalary)
	assert.Nil(t, v.BankAccount, "banking hidden without banking_view")
	assert.Nil(t, v.BankName)
	assert.Nil(t, v.BankHolderName)
	assert.Nil(t, v.PaymentMethod)
}

func TestFieldVisibility_SalaryViewKeepsSalaryStripsBanking(t *testing.T) {
	v := sampleEmployeeRead()
	services.ApplyEmployeeFieldVisibility(v, services.EmployeeFieldPerms{SalaryView: true}, false)
	require.NotNil(t, v.BasicSalary)
	assert.InDelta(t, 1000, *v.BasicSalary, 0.01)
	assert.Nil(t, v.BankAccount, "banking still hidden — sections gate independently")
}

func TestFieldVisibility_BankingViewMasksAccountOnRead(t *testing.T) {
	v := sampleEmployeeRead()
	services.ApplyEmployeeFieldVisibility(v, services.EmployeeFieldPerms{BankingView: true}, false)
	require.NotNil(t, v.BankAccount)
	assert.Equal(t, "•••• 9012", *v.BankAccount, "account number masked on read")
	require.NotNil(t, v.BankName, "other banking fields remain visible")
}

func TestFieldVisibility_UnmaskedOnWriteEcho(t *testing.T) {
	v := sampleEmployeeRead()
	services.ApplyEmployeeFieldVisibility(v, services.EmployeeFieldPerms{BankingView: true}, true)
	require.NotNil(t, v.BankAccount)
	assert.Equal(t, "123456789012", *v.BankAccount, "write echo returns the full account number")
}

func TestFieldVisibility_ManageImpliesViewButStillMasksOnRead(t *testing.T) {
	v := sampleEmployeeRead()
	services.ApplyEmployeeFieldVisibility(v, services.EmployeeFieldPerms{SalaryManage: true, BankingManage: true}, false)
	require.NotNil(t, v.BasicSalary, "manage implies view for salary")
	require.NotNil(t, v.BankAccount, "manage implies view for banking")
	assert.Equal(t, "•••• 9012", *v.BankAccount, "still masked on read even for managers")
}

func TestGuardSalaryWrite(t *testing.T) {
	err := services.GuardSalaryWrite(true, services.EmployeeFieldPerms{})
	require.Error(t, err, "setting salary without salary_manage is forbidden")
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)

	require.NoError(t, services.GuardSalaryWrite(true, services.EmployeeFieldPerms{SalaryManage: true}))
	require.NoError(t, services.GuardSalaryWrite(false, services.EmployeeFieldPerms{}), "not setting salary is always allowed")
}

func TestGuardBankingWrite(t *testing.T) {
	err := services.GuardBankingWrite(true, services.EmployeeFieldPerms{})
	require.Error(t, err, "setting banking without banking_manage is forbidden")
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)

	require.NoError(t, services.GuardBankingWrite(true, services.EmployeeFieldPerms{BankingManage: true}))
	require.NoError(t, services.GuardBankingWrite(false, services.EmployeeFieldPerms{}), "not setting banking is always allowed")
}
