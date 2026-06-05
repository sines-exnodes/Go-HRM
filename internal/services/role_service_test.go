package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newRoleSvc(t *testing.T) *services.RoleService {
	t.Helper()
	return services.NewRoleService(repositories.NewRoleRepository(testDB))
}

func TestRoleService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	r, err := svc.Create(ctx, dto.RoleCreate{
		Name:        "  Auditor  ",
		Description: "  reads things  ",
		Level:       40,
		Permissions: []string{string(permissions.PermUsersRead), string(permissions.PermRolesRead)},
	})
	require.NoError(t, err)
	require.Equal(t, "Auditor", r.Name)
	require.Equal(t, "reads things", r.Description)
	require.Equal(t, 40, r.Level)
	require.False(t, r.IsSystem)
	require.Equal(t, 2, r.PermissionCount)
	require.Len(t, r.Permissions, 2)
}

func TestRoleService_Create_DuplicateName_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Reviewer", Level: 30})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "reviewer", Level: 30})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestRoleService_Create_InvalidName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "bad@name!", Level: 30})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Create_UnknownPermission_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Weird", Level: 30, Permissions: []string{"made:up"}})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Update_Partial_And_SystemGuards(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	r, err := svc.Create(ctx, dto.RoleCreate{Name: "Temp", Level: 30, Permissions: []string{string(permissions.PermUsersRead)}})
	require.NoError(t, err)
	newName := "Temp Renamed"
	newLevel := 35
	empty := []string{}
	upd, err := svc.Update(ctx, r.ID, dto.RoleUpdate{Name: &newName, Level: &newLevel, Permissions: &empty})
	require.NoError(t, err)
	require.Equal(t, "Temp Renamed", upd.Name)
	require.Equal(t, 35, upd.Level)
	require.Equal(t, 0, upd.PermissionCount)

	sys := makeRole(t, "System One", []permissions.Permission{permissions.PermAuthLogin}, true)
	renamed := "Nope"
	_, err = svc.Update(ctx, sys.ID, dto.RoleUpdate{Name: &renamed})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)

	lvl := 5
	_, err = svc.Update(ctx, sys.ID, dto.RoleUpdate{Level: &lvl})
	require.Error(t, err)

	perms := []string{string(permissions.PermAuthLogin), string(permissions.PermUsersRead)}
	okUpd, err := svc.Update(ctx, sys.ID, dto.RoleUpdate{Permissions: &perms})
	require.NoError(t, err)
	require.Equal(t, 2, okUpd.PermissionCount)
}

func TestRoleService_Delete_SystemRole_Rejected(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	sys := makeRole(t, "Protected", []permissions.Permission{permissions.PermAuthLogin}, true)
	err := svc.Delete(ctx, sys.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Delete_BlockedByAssignedUsers_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	role := makeRole(t, "Assigned", []permissions.Permission{permissions.PermAuthLogin}, false)
	makeUser(t, "holder@example.com", "pw-Aa123456", role)

	err := svc.Delete(ctx, role.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "reassign")
}

func TestRoleService_Delete_SoftDeletes_And_NameReusable(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	r, err := svc.Create(ctx, dto.RoleCreate{Name: "Recyclable", Level: 30})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(ctx, r.ID))

	_, err = svc.Get(ctx, r.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)

	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Recyclable", Level: 30})
	require.NoError(t, err)
}

func TestRoleService_List_SortedByLevelThenName_WithSearch(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Zeta", Level: 10})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Alpha", Level: 90})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Beta", Level: 10})
	require.NoError(t, err)

	res, err := svc.List(ctx, dto.RoleListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(3), res.Total)
	require.Equal(t, "Beta", res.Items[0].Name)
	require.Equal(t, "Zeta", res.Items[1].Name)
	require.Equal(t, "Alpha", res.Items[2].Name)

	res, err = svc.List(ctx, dto.RoleListQuery{Page: 1, PageSize: 10, Search: "et"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total)
}
