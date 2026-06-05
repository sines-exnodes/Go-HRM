package services_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Stub uploader so unit tests don't need real S3 ----

type stubUploader struct {
	uploaded int32 // count of Upload calls
	deleted  int32 // count of Delete calls
	lastKey  string
}

func (u *stubUploader) Upload(_ context.Context, subdir, ext string, _ []byte, _ string) (string, error) {
	atomic.AddInt32(&u.uploaded, 1)
	url := fmt.Sprintf("https://stub.test/%s/%s%s", subdir, uuid.NewString(), ext)
	u.lastKey = url
	return url, nil
}
func (u *stubUploader) Delete(_ context.Context, _ string) error {
	atomic.AddInt32(&u.deleted, 1)
	return nil
}
func (u *stubUploader) PublicURL(key string) string { return "https://stub.test/" + key }

// ---- Helpers ----

func newSkillSvc(t *testing.T, up services.Uploader) (*services.SkillService, repositories.SkillRepository, repositories.EmployeeSkillRepository) {
	t.Helper()
	sr := repositories.NewSkillRepository(testDB)
	esr := repositories.NewEmployeeSkillRepository(testDB)
	emps := repositories.NewEmployeeRepository(testDB)
	return services.NewSkillService(sr, esr, emps, up), sr, esr
}

// 1x1 transparent PNG — minimal bytes that http.DetectContentType
// recognises as image/png.
var pngBytes = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}

// ---- Skill catalog tests ----

func TestSkillService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newSkillSvc(t, nil)
	s, err := svc.Create(ctx, dto.SkillCreate{Name: " Go ", Description: "  programming language  "}, nil)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, s.ID)
	// Name must be trimmed; description too.
	require.Equal(t, "Go", s.Name)
	require.Equal(t, "programming language", s.Description)
	require.Nil(t, s.IconURL)
}

func TestSkillService_Create_BlankName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "   "}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestSkillService_Create_InvalidCharacter_BadRequest(t *testing.T) {
	// REVISION NOTES item #2: name regex ^[a-zA-Z0-9 &.+#/()-]+$.
	// The skill name "Bad@Skill" contains '@' which is NOT in the set.
	// This test guards the regex — if someone relaxes it without
	// updating the Python contract, the test fails loudly.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "Bad@Skill"}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok, "expected AppError, got %v", err)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
	require.Contains(t, strings.ToLower(ae.Message), "invalid")
}

func TestSkillService_Create_NameTooLong_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	// 101 chars of legal alphabet.
	long := strings.Repeat("a", 101)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: long}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestSkillService_Create_DescriptionTooLong_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	long := strings.Repeat("a", 501)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "Go", Description: long}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestSkillService_Create_DuplicateName_CaseInsensitive_Conflict(t *testing.T) {
	// Critical regression guard: Python preserves case-insensitive
	// uniqueness on skill name. The partial unique index in 000006
	// also enforces this at the DB level — but the service-layer
	// check must produce a friendly 409 instead of a 500 Postgres
	// error, so this test ensures we hit the pre-check path.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "Python"}, nil)
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.SkillCreate{Name: "python"}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestSkillService_Create_WithIcon_StubUpload(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	up := &stubUploader{}
	svc, _, _ := newSkillSvc(t, up)
	s, err := svc.Create(ctx, dto.SkillCreate{Name: "Go"}, &services.SkillIconUpload{
		Content: pngBytes,
		Ext:     ".png",
	})
	require.NoError(t, err)
	require.NotNil(t, s.IconURL)
	require.Contains(t, *s.IconURL, "skill-icons/")
	require.Equal(t, int32(1), atomic.LoadInt32(&up.uploaded))
}

func TestSkillService_Create_IconContentSpoof_Rejected(t *testing.T) {
	// Critical security guard: review-fix #2 forbids trusting the
	// client-supplied Content-Type. A file that *claims* to be an
	// image but whose sniffed bytes are not in the allowlist must be
	// rejected before the upload call is even made.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	up := &stubUploader{}
	svc, _, _ := newSkillSvc(t, up)
	// "GIF87a" prefix would sniff as image/gif, so use a clear non-image.
	notAnImage := []byte("<?php echo \"pwned\"; ?>")
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "Evil"}, &services.SkillIconUpload{
		Content:     notAnImage,
		ContentType: "image/png", // lying header
		Ext:         ".png",
	})
	ae, ok := apperrors.As(err)
	require.True(t, ok, "expected AppError, got %v", err)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
	require.Equal(t, int32(0), atomic.LoadInt32(&up.uploaded), "upload must NOT happen on a content-spoofed file")
}

func TestSkillService_Update_RenameAndDescription(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	s, err := svc.Create(ctx, dto.SkillCreate{Name: "Go", Description: "old"}, nil)
	require.NoError(t, err)
	newName := "Go Lang"
	newDesc := "new"
	out, err := svc.Update(ctx, s.ID, dto.SkillUpdate{Name: &newName, Description: &newDesc}, nil)
	require.NoError(t, err)
	require.Equal(t, "Go Lang", out.Name)
	require.Equal(t, "new", out.Description)
}

func TestSkillService_Update_NewIcon_ReplacesPrevious(t *testing.T) {
	// Verifies that when an update supplies a new icon, the PRIOR icon
	// object is deleted from storage (best-effort cleanup so we don't
	// leak orphan objects in Supabase).
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	up := &stubUploader{}
	svc, _, _ := newSkillSvc(t, up)
	s, err := svc.Create(ctx, dto.SkillCreate{Name: "Go"}, &services.SkillIconUpload{Content: pngBytes, Ext: ".png"})
	require.NoError(t, err)
	require.NotNil(t, s.IconURL)
	prior := *s.IconURL
	_, err = svc.Update(ctx, s.ID, dto.SkillUpdate{}, &services.SkillIconUpload{Content: pngBytes, Ext: ".png"})
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&up.uploaded))
	require.Equal(t, int32(1), atomic.LoadInt32(&up.deleted), "prior icon must be deleted on replacement")
	_ = prior
}

func TestSkillService_Update_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	name := "Anything"
	_, err := svc.Update(ctx, uuid.New(), dto.SkillUpdate{Name: &name}, nil)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestSkillService_Delete_BlockedWhenAssigned_409WithDetails(t *testing.T) {
	// REVISION NOTES item #2 + #3 mandate the conflict body include
	// employee_count so the FE can render a useful message. If the
	// guard ever regresses to "just return 409 with no body", this
	// test catches it.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, esr := newSkillSvc(t, nil)

	s, err := svc.Create(ctx, dto.SkillCreate{Name: "SQL"}, nil)
	require.NoError(t, err)

	// Attach to an employee.
	u := makeUser(t, "emp-delete-guard@example.com", "pw-Aa123456")
	e := makeEmployee(t, u, "Delete Guard Subject")
	require.NoError(t, esr.ReplaceForEmployee(ctx, e.ID, []uuid.UUID{s.ID}))

	err = svc.Delete(ctx, s.ID)
	ae, ok := apperrors.As(err)
	require.True(t, ok, "expected AppError, got %v", err)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.NotNil(t, ae.Details)
	require.Equal(t, int64(1), ae.Details["employee_count"], "conflict body must expose employee_count for the FE")
	require.Equal(t, s.ID, ae.Details["skill_id"])
}

func TestSkillService_Delete_OKWhenUnassigned(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, sr, _ := newSkillSvc(t, nil)
	s, err := svc.Create(ctx, dto.SkillCreate{Name: "Rust"}, nil)
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, s.ID))
	// FindByID with NotDeleted scope should now miss.
	_, err = sr.FindByID(ctx, s.ID)
	require.Error(t, err, "soft-deleted skill must not be returned by FindByID")
}

func TestSkillService_Delete_RemovesIconFromStorage(t *testing.T) {
	// Parity with the Python source + BA DR-008-003-04 SR-08/AC-14: deleting a
	// skill must also remove its icon object from storage (best-effort). The
	// stub uploader's delete counter proves the cleanup ran.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	up := &stubUploader{}
	svc, _, _ := newSkillSvc(t, up)

	s, err := svc.Create(ctx, dto.SkillCreate{Name: "WithIcon"}, &services.SkillIconUpload{Content: pngBytes, Ext: ".png"})
	require.NoError(t, err)
	require.NotNil(t, s.IconURL)
	require.Equal(t, int32(0), atomic.LoadInt32(&up.deleted))

	require.NoError(t, svc.Delete(ctx, s.ID))
	require.Equal(t, int32(1), atomic.LoadInt32(&up.deleted), "skill icon must be deleted from storage on delete")
}

func TestSkillService_Create_IconTooLarge_Rejected(t *testing.T) {
	// BA DR-008-003-02 AC-09/SR-08: icons are capped at 2MB. A 2MB+1 payload
	// must be rejected before any upload happens.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	up := &stubUploader{}
	svc, _, _ := newSkillSvc(t, up)

	tooBig := make([]byte, 2*1024*1024+1)
	_, err := svc.Create(ctx, dto.SkillCreate{Name: "Big"}, &services.SkillIconUpload{Content: tooBig, Ext: ".png"})
	ae, ok := apperrors.As(err)
	require.True(t, ok, "expected AppError, got %v", err)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
	require.Contains(t, ae.Message, "2MB")
	require.Equal(t, int32(0), atomic.LoadInt32(&up.uploaded), "oversized icon must not be uploaded")
}

func TestSkillService_List_Pagination_And_SortByNameASC(t *testing.T) {
	// REVISION NOTES item #2: "sort by name ASC". This is a public-API
	// contract — Python clients depend on it. Guard against accidental
	// resort.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)

	// Insert in non-alphabetical order.
	for _, n := range []string{"Zebra", "Apple", "Mango"} {
		_, err := svc.Create(ctx, dto.SkillCreate{Name: n}, nil)
		require.NoError(t, err)
	}
	out, err := svc.List(ctx, dto.SkillListQuery{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(3), out.Total)
	require.Equal(t, "Apple", out.Items[0].Name)
	require.Equal(t, "Mango", out.Items[1].Name)
	require.Equal(t, "Zebra", out.Items[2].Name)
}

func TestSkillService_List_SearchILIKE(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	for _, n := range []string{"Go", "Golang", "Python", "JavaScript"} {
		_, err := svc.Create(ctx, dto.SkillCreate{Name: n}, nil)
		require.NoError(t, err)
	}
	out, err := svc.List(ctx, dto.SkillListQuery{Search: "go"})
	require.NoError(t, err)
	require.Equal(t, int64(2), out.Total, "ILIKE %go% must match both 'Go' and 'Golang'")
}

// ---- Employee ↔ Skill assignment tests ----

func TestSkillService_ReplaceForEmployee_SetsAndUnsets(t *testing.T) {
	// PUT-replace semantics are the load-bearing piece of the assignment
	// API (REVISION NOTES item #3): the whole set is replaced atomically.
	// This test exercises an add + a remove in a single call.
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	a, _ := svc.Create(ctx, dto.SkillCreate{Name: "A"}, nil)
	b, _ := svc.Create(ctx, dto.SkillCreate{Name: "B"}, nil)
	c, _ := svc.Create(ctx, dto.SkillCreate{Name: "C"}, nil)

	u := makeUser(t, "emp-replace@example.com", "pw-Aa123456")
	e := makeEmployee(t, u, "Replace Test")

	// Round 1: assign {A, B}.
	got, err := svc.ReplaceForEmployee(ctx, e.ID, []uuid.UUID{a.ID, b.ID})
	require.NoError(t, err)
	require.Len(t, got, 2)

	// Round 2: replace with {B, C} — A must be dropped, C added.
	got, err = svc.ReplaceForEmployee(ctx, e.ID, []uuid.UUID{b.ID, c.ID})
	require.NoError(t, err)
	require.Len(t, got, 2)
	names := []string{got[0].Name, got[1].Name}
	require.Contains(t, names, "B")
	require.Contains(t, names, "C")
	require.NotContains(t, names, "A")
}

func TestSkillService_ReplaceForEmployee_Reactivation(t *testing.T) {
	// Soft-deleted join rows from a previous assignment must be
	// re-activated (rather than insert + tripping the partial unique
	// index uq_employee_skills_pair_live).
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	a, _ := svc.Create(ctx, dto.SkillCreate{Name: "A"}, nil)
	u := makeUser(t, "emp-react@example.com", "pw-Aa123456")
	e := makeEmployee(t, u, "Reactivation Test")

	require.NoError(t, mustReplace(svc, ctx, e.ID, a.ID))
	require.NoError(t, mustReplace(svc, ctx, e.ID /* clear */))
	require.NoError(t, mustReplace(svc, ctx, e.ID, a.ID))

	got, err := svc.ListForEmployee(ctx, e.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, a.ID, got[0].ID)
}

func TestSkillService_ReplaceForEmployee_InvalidSkillID_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	u := makeUser(t, "emp-invalid@example.com", "pw-Aa123456")
	e := makeEmployee(t, u, "Invalid Skill Test")
	_, err := svc.ReplaceForEmployee(ctx, e.ID, []uuid.UUID{uuid.New()})
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestSkillService_ReplaceForEmployee_EmployeeMissing_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc, _, _ := newSkillSvc(t, nil)
	a, _ := svc.Create(ctx, dto.SkillCreate{Name: "A"}, nil)
	_, err := svc.ReplaceForEmployee(ctx, uuid.New(), []uuid.UUID{a.ID})
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

func TestSkillService_ListForEmployee_EmployeeMissing_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _, _ := newSkillSvc(t, nil)
	_, err := svc.ListForEmployee(context.Background(), uuid.New())
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)
}

// mustReplace is a tiny shim so the round-trip reactivation test stays
// readable.
func mustReplace(svc *services.SkillService, ctx context.Context, empID uuid.UUID, skillIDs ...uuid.UUID) error {
	_, err := svc.ReplaceForEmployee(ctx, empID, skillIDs)
	return err
}
