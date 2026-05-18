package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// makeDept is a small helper to create a department via the service so
// positions have a valid FK target.
func makeDept(t *testing.T, name string) *dto.DepartmentRead {
	t.Helper()
	svc, _, _ := newDeptSvc(t)
	d, err := svc.Create(context.Background(), dto.DepartmentCreate{Name: name})
	require.NoError(t, err)
	return d
}

func TestPositionService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "Engineering")
	svc, _, _ := newPosSvc(t)

	p, err := svc.Create(ctx, dto.PositionCreate{
		Name:         "Backend Engineer",
		Description:  "  builds APIs  ",
		DepartmentID: dept.ID,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, p.ID)
	require.Equal(t, "Backend Engineer", p.Name)
	require.Equal(t, "builds APIs", p.Description)
	require.Equal(t, dept.ID, p.DepartmentID)
	require.NotNil(t, p.Department)
	require.Equal(t, "Engineering", p.Department.Name)
}

func TestPositionService_Create_MissingDept_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newPosSvc(t)
	_, err := svc.Create(ctx, dto.PositionCreate{
		Name:         "Ghost",
		DepartmentID: uuid.New(),
	})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestPositionService_Create_DuplicateNameInDept_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "HR")
	svc, _, _ := newPosSvc(t)

	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Recruiter", DepartmentID: dept.ID})
	require.NoError(t, err)

	// Case-insensitive duplicate within the same department is rejected.
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "recruiter", DepartmentID: dept.ID})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestPositionService_Create_DuplicateNameInDifferentDept_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	deptA := makeDept(t, "Dept A")
	deptB := makeDept(t, "Dept B")
	svc, _, _ := newPosSvc(t)

	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Manager", DepartmentID: deptA.ID})
	require.NoError(t, err)
	// Same name in a different department is allowed.
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Manager", DepartmentID: deptB.ID})
	require.NoError(t, err)
	require.Equal(t, deptB.ID, p.DepartmentID)
}

func TestPositionService_Get_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "Support")
	svc, _, _ := newPosSvc(t)
	created, err := svc.Create(ctx, dto.PositionCreate{Name: "Agent", DepartmentID: dept.ID})
	require.NoError(t, err)

	got, err := svc.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Agent", got.Name)

	// Missing -> not found.
	_, err = svc.Get(ctx, uuid.New())
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestPositionService_Update_MoveToOtherDept(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	deptA := makeDept(t, "Origin")
	deptB := makeDept(t, "Target")
	svc, _, _ := newPosSvc(t)

	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Analyst", DepartmentID: deptA.ID})
	require.NoError(t, err)

	newName := "Senior Analyst"
	updated, err := svc.Update(ctx, p.ID, dto.PositionUpdate{
		Name:         &newName,
		DepartmentID: &deptB.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "Senior Analyst", updated.Name)
	require.Equal(t, deptB.ID, updated.DepartmentID)

	// Moving to a non-existent department is rejected.
	missing := uuid.New()
	_, err = svc.Update(ctx, p.ID, dto.PositionUpdate{DepartmentID: &missing})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestPositionService_Delete_BlockedByEmployees_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "Delivery")
	svc, _, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Driver", DepartmentID: dept.ID})
	require.NoError(t, err)

	makeEmployeeInDept(t, dept.ID, &p.ID)

	err = svc.Delete(ctx, p.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "Reassign all employees")
}

func TestPositionService_Delete_SoftDeletesBothColumns(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "Lab")
	svc, _, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Researcher", DepartmentID: dept.ID})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(ctx, p.ID))

	var isDeleted bool
	var hasDeletedAt bool
	row := testDB.Raw(
		"SELECT is_deleted, deleted_at IS NOT NULL FROM positions WHERE id = ?", p.ID,
	).Row()
	require.NoError(t, row.Scan(&isDeleted, &hasDeletedAt))
	require.True(t, isDeleted)
	require.True(t, hasDeletedAt)

	// Soft-deleted position is invisible to Get.
	_, err = svc.Get(ctx, p.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestPositionService_List_SearchAndDeptFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	deptA := makeDept(t, "Eng")
	deptB := makeDept(t, "Mkt")
	svc, _, _ := newPosSvc(t)

	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Frontend Engineer", DepartmentID: deptA.ID})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "Backend Engineer", DepartmentID: deptA.ID})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "Content Manager", DepartmentID: deptB.ID})
	require.NoError(t, err)

	// Search "engineer" -> 2 (both in deptA).
	res, err := svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 10, Search: "engineer"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)

	// Filter by deptB -> 1.
	res, err = svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 10, DepartmentID: deptB.ID.String()})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Total)
	require.Equal(t, "Content Manager", res.Items[0].Name)

	// Pagination over deptA's 2 positions, page size 1 -> 2 total pages.
	res, err = svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 1, DepartmentID: deptA.ID.String()})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)
	require.Equal(t, 2, res.TotalPages)
	require.Len(t, res.Items, 1)
}
