package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newLabelSvc(t *testing.T) *services.LabelService {
	t.Helper()
	return services.NewLabelService(repositories.NewLabelRepository(testDB))
}

func TestLabelService_GetOrCreate_NewLabel_Created(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)

	out, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: " Announcement "})
	require.NoError(t, err)
	require.True(t, out.Created, "first call must create")
	require.Equal(t, "Announcement", out.Label.Name)
}

func TestLabelService_GetOrCreate_ExistingLabel_NotCreated_CaseInsensitive(t *testing.T) {
	// REVISION NOTES item #4 says POST is get-or-create with case-
	// insensitive lookup. This is the canonical idempotency guard —
	// without it, the FE would create N variants ("urgent", "Urgent",
	// "URGENT") of the same logical label.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)

	first, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: "Urgent"})
	require.NoError(t, err)
	require.True(t, first.Created)

	second, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: "urgent"})
	require.NoError(t, err)
	require.False(t, second.Created, "second call with same case-insensitive name must return existing row")
	require.Equal(t, first.Label.ID, second.Label.ID, "same id on idempotent re-create")
}

func TestLabelService_GetOrCreate_BlankName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)
	_, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: "   "})
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLabelService_GetOrCreate_NameTooLong_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)
	_, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: strings.Repeat("a", 51)})
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLabelService_List_SortedByNameASC(t *testing.T) {
	// Public-API contract: Python clients depend on the alphabetic
	// order. Guard against an accidental ORDER BY drift.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)

	for _, n := range []string{"Zebra", "Alpha", "Mango"} {
		_, err := svc.GetOrCreate(ctx, dto.LabelCreate{Name: n})
		require.NoError(t, err)
	}
	list, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 3)
	require.Equal(t, "Alpha", list[0].Name)
	require.Equal(t, "Mango", list[1].Name)
	require.Equal(t, "Zebra", list[2].Name)
}

func TestLabelService_List_Empty(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newLabelSvc(t)
	list, err := svc.List(ctx)
	require.NoError(t, err)
	require.Empty(t, list)
}
