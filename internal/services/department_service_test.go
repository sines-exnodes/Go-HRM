package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newDeptSvc(t *testing.T) (*services.DepartmentService, repositories.DepartmentRepository, repositories.PositionRepository) {
	t.Helper()
	dr := repositories.NewDepartmentRepository(testDB)
	pr := repositories.NewPositionRepository(testDB)
	return services.NewDepartmentService(dr, pr), dr, pr
}

func newPosSvc(t *testing.T) (*services.PositionService, repositories.PositionRepository, repositories.DepartmentRepository) {
	t.Helper()
	pr := repositories.NewPositionRepository(testDB)
	dr := repositories.NewDepartmentRepository(testDB)
	return services.NewPositionService(pr, dr), pr, dr
}

func TestDepartmentService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	d, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Engineering", Description: "  builds things  "})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, d.ID)
	require.Equal(t, "Engineering", d.Name)
	require.Equal(t, "builds things", d.Description)
	require.Nil(t, d.ParentID)
}

func TestDepartmentService_Create_DuplicateName_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	_, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Finance"})
	require.NoError(t, err)

	// Case-insensitive duplicate must be rejected.
	_, err = svc.Create(ctx, dto.DepartmentCreate{Name: "finance"})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestDepartmentService_Create_BlankName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	_, err := svc.Create(ctx, dto.DepartmentCreate{Name: "   "})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestDepartmentService_Create_InvalidParent_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	missing := uuid.New()
	_, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Sub", ParentID: &missing})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestDepartmentService_Update_RenameAndReparent(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	parent, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Parent"})
	require.NoError(t, err)
	child, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Child"})
	require.NoError(t, err)

	newName := "Child Renamed"
	updated, err := svc.Update(ctx, child.ID, dto.DepartmentUpdate{
		Name:     &newName,
		ParentID: &parent.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "Child Renamed", updated.Name)
	require.NotNil(t, updated.ParentID)
	require.Equal(t, parent.ID, *updated.ParentID)

	// ClearParent makes it a root again.
	cleared, err := svc.Update(ctx, child.ID, dto.DepartmentUpdate{ClearParent: true})
	require.NoError(t, err)
	require.Nil(t, cleared.ParentID)
}

func TestDepartmentService_Update_CycleRejected(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	a, err := svc.Create(ctx, dto.DepartmentCreate{Name: "A"})
	require.NoError(t, err)
	b, err := svc.Create(ctx, dto.DepartmentCreate{Name: "B", ParentID: &a.ID})
	require.NoError(t, err)

	// Making A a child of B would create A->B->A.
	_, err = svc.Update(ctx, a.ID, dto.DepartmentUpdate{ParentID: &b.ID})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Self-parent is also rejected.
	_, err = svc.Update(ctx, a.ID, dto.DepartmentUpdate{ParentID: &a.ID})
	require.Error(t, err)
	ae, ok = apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestDepartmentService_Delete_BlockedByChildren_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	parent, err := svc.Create(ctx, dto.DepartmentCreate{Name: "ParentDept"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.DepartmentCreate{Name: "ChildDept", ParentID: &parent.ID})
	require.NoError(t, err)

	err = svc.Delete(ctx, parent.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "child departments")
}

func TestDepartmentService_Delete_BlockedByPositions_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "WithPositions"})
	require.NoError(t, err)

	posSvc, _, _ := newPosSvc(t)
	_, err = posSvc.Create(ctx, dto.PositionCreate{Name: "Engineer", DepartmentID: dept.ID})
	require.NoError(t, err)

	err = svc.Delete(ctx, dept.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "this department")
}

func TestDepartmentService_Delete_BlockedByEmployees_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Sales"})
	require.NoError(t, err)

	makeEmployeeInDept(t, dept.ID, nil)

	err = svc.Delete(ctx, dept.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "Reassign all employees")
}

func TestDepartmentService_Delete_SoftDeletesBothColumns(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Ops"})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(ctx, dept.ID))

	var isDeleted bool
	var hasDeletedAt bool
	row := testDB.Raw(
		"SELECT is_deleted, deleted_at IS NOT NULL FROM departments WHERE id = ?", dept.ID,
	).Row()
	require.NoError(t, row.Scan(&isDeleted, &hasDeletedAt))
	require.True(t, isDeleted)
	require.True(t, hasDeletedAt)

	// Soft-deleted department is invisible to Get.
	_, err = svc.Get(ctx, dept.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestDepartmentService_List_SearchAndParentFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newDeptSvc(t)
	root, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Alpha Group"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.DepartmentCreate{Name: "Beta Group"})
	require.NoError(t, err)
	childA, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Alpha Child", ParentID: &root.ID})
	require.NoError(t, err)

	// Search "alpha" -> Alpha Group + Alpha Child.
	res, err := svc.List(ctx, dto.DepartmentListQuery{Page: 1, PageSize: 10, Search: "alpha"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)

	// Root-only filter -> the two top-level departments.
	res, err = svc.List(ctx, dto.DepartmentListQuery{Page: 1, PageSize: 10, ParentID: "root"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)

	// Children of root -> only Alpha Child.
	res, err = svc.List(ctx, dto.DepartmentListQuery{Page: 1, PageSize: 10, ParentID: root.ID.String()})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Total)
	require.Equal(t, childA.ID, res.Items[0].ID)

	// Pagination: page size 1 over the 2 roots -> 2 total pages.
	res, err = svc.List(ctx, dto.DepartmentListQuery{Page: 1, PageSize: 1, ParentID: "root"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)
	require.Equal(t, 2, res.TotalPages)
	require.Len(t, res.Items, 1)

	// Invalid parent_id -> bad request.
	_, err = svc.List(ctx, dto.DepartmentListQuery{Page: 1, PageSize: 10, ParentID: "not-a-uuid"})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}
