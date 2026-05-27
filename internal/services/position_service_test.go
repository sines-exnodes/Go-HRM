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
// employee fixtures have a valid FK target. Positions themselves are no
// longer associated with a department (post-000014).
func makeDept(t *testing.T, name string) *dto.DepartmentRead {
	t.Helper()
	svc, _ := newDeptSvc(t)
	d, err := svc.Create(context.Background(), dto.DepartmentCreate{Name: name})
	require.NoError(t, err)
	return d
}

func TestPositionService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{
		Name:        "Backend Engineer",
		Description: "  builds APIs  ",
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, p.ID)
	require.Equal(t, "Backend Engineer", p.Name)
	require.Equal(t, "builds APIs", p.Description)
	// Freshly created position has zero employees.
	require.Equal(t, int64(0), p.EmployeeCount)
}

// Post-000014: position names are globally unique (case-insensitive).
// The previous "same name in different department is OK" behavior is gone.
func TestPositionService_Create_DuplicateName_GlobalConflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Manager"})
	require.NoError(t, err)

	// Case-insensitive duplicate anywhere in the table is rejected.
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "manager"})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestPositionService_Create_BlankName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	_, err := svc.Create(ctx, dto.PositionCreate{Name: "   "})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestPositionService_Get_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	created, err := svc.Create(ctx, dto.PositionCreate{Name: "Agent"})
	require.NoError(t, err)

	got, err := svc.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Agent", got.Name)
	require.Equal(t, int64(0), got.EmployeeCount)

	// Missing -> not found.
	_, err = svc.Get(ctx, uuid.New())
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestPositionService_Update_RenameOK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Analyst"})
	require.NoError(t, err)

	newName := "Senior Analyst"
	updated, err := svc.Update(ctx, p.ID, dto.PositionUpdate{Name: &newName})
	require.NoError(t, err)
	require.Equal(t, "Senior Analyst", updated.Name)
}

func TestPositionService_Update_DuplicateName_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Lead"})
	require.NoError(t, err)
	other, err := svc.Create(ctx, dto.PositionCreate{Name: "Junior"})
	require.NoError(t, err)

	clash := "lead"
	_, err = svc.Update(ctx, other.ID, dto.PositionUpdate{Name: &clash})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestPositionService_Delete_BlockedByEmployees_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "Delivery")
	svc, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Driver"})
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

	svc, _ := newPosSvc(t)
	p, err := svc.Create(ctx, dto.PositionCreate{Name: "Researcher"})
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

func TestPositionService_List_Search(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newPosSvc(t)
	_, err := svc.Create(ctx, dto.PositionCreate{Name: "Frontend Engineer"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "Backend Engineer"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.PositionCreate{Name: "Content Manager"})
	require.NoError(t, err)

	// Search "engineer" -> 2 hits.
	res, err := svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 10, Search: "engineer"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)

	// Pagination across all 3 with page size 2 -> 2 total pages.
	res, err = svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), res.Total)
	require.Equal(t, 2, res.TotalPages)
	require.Len(t, res.Items, 2)
}

// EmployeeCount hydration parity with Python. Employees still link to
// positions via employees.position_id; the position layer no longer holds
// a department FK but must still report the live employee count on every
// read path.
func TestPositionService_EmployeeCount_HydratedOnGetAndList(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	dept := makeDept(t, "EngineeringCounts")
	svc, _ := newPosSvc(t)
	a, err := svc.Create(ctx, dto.PositionCreate{Name: "Counted Engineer A"})
	require.NoError(t, err)
	b, err := svc.Create(ctx, dto.PositionCreate{Name: "Counted Engineer B"})
	require.NoError(t, err)

	makeEmployeeInDept(t, dept.ID, &a.ID)
	makeEmployeeInDept(t, dept.ID, &a.ID)

	gotA, err := svc.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), gotA.EmployeeCount)

	gotB, err := svc.Get(ctx, b.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), gotB.EmployeeCount)

	res, err := svc.List(ctx, dto.PositionListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	byID := map[uuid.UUID]int64{}
	for _, item := range res.Items {
		byID[item.ID] = item.EmployeeCount
	}
	require.Equal(t, int64(2), byID[a.ID])
	require.Equal(t, int64(0), byID[b.ID])
}
