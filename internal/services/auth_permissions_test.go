package services

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
)

func TestResolvePermsFromUser_AnnouncementManageImpliesRead(t *testing.T) {
	svc := &AuthService{}

	perms, err := svc.resolvePermsFromUser([]models.Role{
		{Permissions: models.StringSlice{string(permissions.PermAnnounceManage)}},
	})

	require.NoError(t, err)
	require.True(t, perms[permissions.PermAnnounceManage])
	require.True(t, perms[permissions.PermAnnounceRead])
}
