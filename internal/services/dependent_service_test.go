package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newDepSvc(db *gorm.DB) *services.DependentService {
	return services.NewDependentService(
		repositories.NewDependentRepository(db),
		repositories.NewEmployeeRepository(db),
	)
}

func TestDependentService_CRUD(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	depSvc := newDepSvc(testDB)
	ctx := context.Background()

	emp, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "dep@example.com", Password: "Pass12345", FullName: "Owner",
	})
	require.NoError(t, err)

	created, err := depSvc.Create(ctx, emp.ID, dto.DependentCreate{
		FullName: "Child One", Relationship: "child",
	})
	require.NoError(t, err)
	assert.Equal(t, "Child One", created.FullName)
	assert.Equal(t, emp.ID, created.EmployeeID)

	items, total, err := depSvc.List(ctx, emp.ID, dto.DependentListQuery{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)

	newName := "Child Renamed"
	upd, err := depSvc.Update(ctx, created.ID, dto.DependentUpdate{FullName: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Child Renamed", upd.FullName)

	require.NoError(t, depSvc.Delete(ctx, created.ID))
	_, total, err = depSvc.List(ctx, emp.ID, dto.DependentListQuery{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestDependentService_CreateForMissingEmployee_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	depSvc := newDepSvc(testDB)
	ctx := context.Background()

	_, err := depSvc.Create(ctx, uuid.New(), dto.DependentCreate{FullName: "Ghost", Relationship: "child"})
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestDependentService_NonOwnerForbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	depSvc := newDepSvc(testDB)
	ctx := context.Background()

	owner, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "owner@example.com", Password: "Pass12345", FullName: "Owner",
	})
	require.NoError(t, err)
	stranger, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "stranger@example.com", Password: "Pass12345", FullName: "Stranger",
	})
	require.NoError(t, err)

	// Stranger tries to manage owner's dependents without admin perm -> forbidden.
	err = depSvc.AuthorizeOwnerOrAdmin(ctx, stranger.UserID, owner.ID, false)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)

	// Owner OK.
	err = depSvc.AuthorizeOwnerOrAdmin(ctx, owner.UserID, owner.ID, false)
	require.NoError(t, err)

	// Admin OK regardless of ownership.
	err = depSvc.AuthorizeOwnerOrAdmin(ctx, stranger.UserID, owner.ID, true)
	require.NoError(t, err)
}

func TestDependentService_OwnerEmployeeIDForDependent(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	depSvc := newDepSvc(testDB)
	ctx := context.Background()

	emp, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "ownerlookup@example.com", Password: "Pass12345", FullName: "OwnerLookup",
	})
	require.NoError(t, err)
	dep, err := depSvc.Create(ctx, emp.ID, dto.DependentCreate{FullName: "Kid", Relationship: "child"})
	require.NoError(t, err)

	ownerID, err := depSvc.OwnerEmployeeIDForDependent(ctx, dep.ID)
	require.NoError(t, err)
	assert.Equal(t, emp.ID, ownerID)

	_, err = depSvc.OwnerEmployeeIDForDependent(ctx, uuid.New())
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeNotFound, ae.Code)
}
