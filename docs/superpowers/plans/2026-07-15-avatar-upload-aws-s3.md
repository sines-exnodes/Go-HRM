# AWS S3 Avatar Upload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upload employee avatars to AWS S3 under `hrm-app/avatars/<uuid>.<ext>`, persist AWS public URLs, and remove all active Supabase storage behavior.

**Architecture:** Keep the existing `Uploader` interface and feature-specific handlers. Replace the shared adapter's custom Supabase endpoint logic with AWS SDK regional endpoint resolution, make AWS storage configuration mandatory at boot, and change only the employee avatar prefix. Add deterministic unit tests plus an explicitly enabled live S3 probe.

**Tech Stack:** Go 1.25, Gin, AWS SDK for Go v2 S3, GORM, Testify.

## Global Constraints

- AWS S3 is the only supported object-storage backend.
- Avatar object keys are exactly `hrm-app/avatars/<uuid>.<lowercase-extension>`.
- Required variables are `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, `STORAGE_REGION`, and `STORAGE_BUCKET`.
- `STORAGE_ENDPOINT`, Supabase URL construction, custom S3 base endpoints, and path-style addressing are removed.
- Existing avatar routes, multipart field `avatar`, authorization, 5 MB limit, and byte-sniffed PNG/JPEG/WEBP validation remain unchanged.
- Existing skill, leave, and contract prefixes remain unchanged while their shared transport switches to AWS S3.
- Only URLs belonging to the configured AWS bucket and region are eligible for deletion; foreign and legacy URLs remain silent no-ops.
- Historical plans, specifications, and verification records remain historical. Remove Supabase references only from active code, examples, tests, and current reference documentation.
- Never stage or commit unrelated user changes already present in the worktree.

---

## File Map

- `internal/config/storage.go` — AWS-only storage fields and validation.
- `internal/config/storage_test.go` — deterministic storage configuration tests.
- `internal/config/config.go` — fail boot when storage configuration is incomplete.
- `internal/services/upload_service.go` — AWS S3 client, upload/delete, and AWS public URLs.
- `internal/services/upload_service_test.go` — AWS client options and URL ownership tests.
- `internal/services/avatar_upload_test.go` — avatar prefix and compensation behavior without a database.
- `internal/services/employee_service.go` — avatar prefix and cleanup after persistence failure.
- `internal/services/employee_service_test.go` — AWS-shaped fake URLs in existing integration coverage.
- `internal/services/skill_service_test.go` — remove stale Supabase-only wording.
- `internal/services/upload_service_integration_test.go` — opt-in live AWS upload/read/delete probe.
- `.env.example`, `.env.docker.example` — AWS-only example configuration.
- `CLAUDE.md`, `docs/API-REFERENCE.md` — current storage documentation.

---

### Task 1: Replace Supabase storage adapter with AWS S3

**Files:**
- Modify: `internal/config/storage.go:9-43`
- Create: `internal/config/storage_test.go`
- Modify: `internal/config/config.go:158-162`
- Modify: `internal/services/upload_service.go:20-103`
- Modify: `internal/services/upload_service_test.go:1-38`
- Modify: `.env.example:49-57`
- Modify: `.env.docker.example:41-46`

**Interfaces:**
- Consumes: four `STORAGE_*` AWS variables and `context.Context`.
- Produces: `config.StorageConfig{AccessKey, SecretKey, Region, Bucket}` and unchanged `Uploader` methods `Upload`, `Delete`, and `PublicURL`.

- [ ] **Step 1: Write failing AWS configuration tests**

Create `internal/config/storage_test.go`:

```go
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStorageConfigRequiresEveryAWSValue(t *testing.T) {
	t.Setenv("STORAGE_ACCESS_KEY", "access")
	t.Setenv("STORAGE_SECRET_KEY", "secret")
	t.Setenv("STORAGE_REGION", "")
	t.Setenv("STORAGE_BUCKET", "bucket")

	_, err := LoadStorageConfig()

	require.EqualError(t, err, "storage config: STORAGE_ACCESS_KEY/STORAGE_SECRET_KEY/STORAGE_REGION/STORAGE_BUCKET are required")
}

func TestLoadStorageConfigLoadsAWSValues(t *testing.T) {
	t.Setenv("STORAGE_ACCESS_KEY", "access")
	t.Setenv("STORAGE_SECRET_KEY", "secret")
	t.Setenv("STORAGE_REGION", "ap-southeast-1")
	t.Setenv("STORAGE_BUCKET", "devshared-ap-southeast-1-public-storage")

	got, err := LoadStorageConfig()

	require.NoError(t, err)
	assert.Equal(t, StorageConfig{
		AccessKey: "access",
		SecretKey: "secret",
		Region:    "ap-southeast-1",
		Bucket:    "devshared-ap-southeast-1-public-storage",
	}, got)
}
```

- [ ] **Step 2: Replace upload-service tests with AWS expectations**

Replace `internal/services/upload_service_test.go` with:

```go
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
)

func awsTestStorageConfig() config.StorageConfig {
	return config.StorageConfig{
		AccessKey: "access",
		SecretKey: "secret",
		Region:    "ap-southeast-1",
		Bucket:    "devshared-ap-southeast-1-public-storage",
	}
}

func TestNewUploadServiceUsesAWSDefaults(t *testing.T) {
	svc, err := NewUploadService(context.Background(), awsTestStorageConfig())
	require.NoError(t, err)

	opts := svc.client.Options()
	assert.Equal(t, "ap-southeast-1", opts.Region)
	assert.Nil(t, opts.BaseEndpoint, "AWS must resolve its regional endpoint")
	assert.False(t, opts.UsePathStyle, "AWS must use virtual-hosted bucket addressing")
}

func TestBuildPublicURL(t *testing.T) {
	svc := &UploadService{cfg: awsTestStorageConfig()}
	got := svc.PublicURL("hrm-app/avatars/abc.png")
	want := "https://devshared-ap-southeast-1-public-storage.s3.ap-southeast-1.amazonaws.com/hrm-app/avatars/abc.png"
	assert.Equal(t, want, got)
}

func TestExtractObjectPath(t *testing.T) {
	svc := &UploadService{cfg: awsTestStorageConfig()}
	owned := "https://devshared-ap-southeast-1-public-storage.s3.ap-southeast-1.amazonaws.com/hrm-app/avatars/abc.png"

	assert.Equal(t, "hrm-app/avatars/abc.png", svc.objectPathFromURL(owned))
	assert.Empty(t, svc.objectPathFromURL(""))
	assert.Empty(t, svc.objectPathFromURL("https://other-bucket.s3.ap-southeast-1.amazonaws.com/x.png"))
}
```

- [ ] **Step 3: Run tests and confirm Supabase implementation fails them**

Run:

```powershell
go test ./internal/config ./internal/services -run 'TestLoadStorageConfig|TestNewUploadServiceUsesAWSDefaults|TestBuildPublicURL|TestExtractObjectPath' -count=1 -v
```

Expected: failures mention missing region validation, non-nil `BaseEndpoint`, enabled `UsePathStyle`, or Supabase public URL mismatch.

- [ ] **Step 4: Implement AWS-only storage configuration**

Replace `internal/config/storage.go` with:

```go
package config

import (
	"fmt"
	"os"
)

type StorageConfig struct {
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
}

func LoadStorageConfig() (StorageConfig, error) {
	c := StorageConfig{
		AccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
		SecretKey: os.Getenv("STORAGE_SECRET_KEY"),
		Region:    os.Getenv("STORAGE_REGION"),
		Bucket:    os.Getenv("STORAGE_BUCKET"),
	}
	if c.AccessKey == "" || c.SecretKey == "" || c.Region == "" || c.Bucket == "" {
		return c, fmt.Errorf("storage config: STORAGE_ACCESS_KEY/STORAGE_SECRET_KEY/STORAGE_REGION/STORAGE_BUCKET are required")
	}
	return c, nil
}
```

Replace the optional-storage block in `internal/config/config.go` with:

```go
	storage, err := LoadStorageConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg.Storage = storage
```

- [ ] **Step 5: Implement AWS S3 client and public URLs**

Replace `internal/services/upload_service.go` with:

```go
package services

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/config"
)

// Uploader is the storage-side abstraction consumed by feature services.
// UploadService is the AWS-S3-backed implementation.
type Uploader interface {
	Upload(ctx context.Context, subdir, ext string, content []byte, contentType string) (string, error)
	Delete(ctx context.Context, publicURL string) error
	PublicURL(key string) string
}

type UploadService struct {
	cfg    config.StorageConfig
	client *s3.Client
}

var _ Uploader = (*UploadService)(nil)

func NewUploadService(ctx context.Context, cfg config.StorageConfig) (*UploadService, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("upload: load aws config: %w", err)
	}
	return &UploadService{cfg: cfg, client: s3.NewFromConfig(awsCfg)}, nil
}

// Upload stores content at `subdir/<uuid><ext>` and returns the public URL.
func (s *UploadService) Upload(ctx context.Context, subdir, ext string, content []byte, contentType string) (string, error) {
	if subdir == "" {
		return "", fmt.Errorf("upload: subdir required")
	}
	key := path.Join(subdir, uuid.NewString()+strings.ToLower(ext))
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload: put object: %w", err)
	}
	return s.PublicURL(key), nil
}

func (s *UploadService) Delete(ctx context.Context, publicURL string) error {
	key := s.objectPathFromURL(publicURL)
	if key == "" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("upload: delete object: %w", err)
	}
	return nil
}

func (s *UploadService) PublicURL(key string) string {
	return fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		s.cfg.Bucket,
		s.cfg.Region,
		strings.TrimLeft(key, "/"),
	)
}

func (s *UploadService) objectPathFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	expectedHost := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.cfg.Bucket, s.cfg.Region)
	if !strings.EqualFold(u.Hostname(), expectedHost) {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}
```

- [ ] **Step 6: Replace example configuration**

Use this storage block in both `.env.example` and `.env.docker.example`:

```dotenv
# ---------- AWS S3 Storage ----------
# Bucket must exist and allow public reads for returned avatar/file URLs.
STORAGE_ACCESS_KEY=replace-me
STORAGE_SECRET_KEY=replace-me
STORAGE_REGION=ap-southeast-1
STORAGE_BUCKET=devshared-ap-southeast-1-public-storage
```

- [ ] **Step 7: Format and run focused tests**

Run:

```powershell
gofmt -w internal/config/storage.go internal/config/storage_test.go internal/services/upload_service.go internal/services/upload_service_test.go
go test ./internal/config ./internal/services -run 'TestLoadStorageConfig|TestNewUploadServiceUsesAWSDefaults|TestBuildPublicURL|TestExtractObjectPath' -count=1 -v
```

Expected: all named tests pass; none skip.

- [ ] **Step 8: Commit Task 1**

```powershell
git add -- internal/config/storage.go internal/config/storage_test.go internal/config/config.go internal/services/upload_service.go internal/services/upload_service_test.go .env.example .env.docker.example
git diff --cached --check
git commit -m "refactor(storage): switch uploader to AWS S3"
```

---

### Task 2: Put avatars under the application namespace

**Files:**
- Create: `internal/services/avatar_upload_test.go`
- Modify: `internal/services/employee_service.go:819-860`
- Modify: `internal/services/employee_service_test.go:18-39`
- Modify: `internal/services/skill_service_test.go:207`

**Interfaces:**
- Consumes: unchanged `Uploader.Upload(ctx, subdir, ext, content, contentType)`.
- Produces: avatar prefix `hrm-app/avatars` and compensation delete when `UpdateAvatarURL` fails.

- [ ] **Step 1: Write deterministic failing avatar test**

Create `internal/services/avatar_upload_test.go`:

```go
package services

import (
	"context"
	"errors"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/repositories"
)

type recordingAvatarUploader struct {
	subdir      string
	contentType string
	uploadedURL string
	deleted     []string
}

func (u *recordingAvatarUploader) Upload(_ context.Context, subdir, ext string, _ []byte, contentType string) (string, error) {
	u.subdir = subdir
	u.contentType = contentType
	u.uploadedURL = u.PublicURL(path.Join(subdir, "new-avatar"+ext))
	return u.uploadedURL, nil
}

func (u *recordingAvatarUploader) Delete(_ context.Context, publicURL string) error {
	u.deleted = append(u.deleted, publicURL)
	return nil
}

func (u *recordingAvatarUploader) PublicURL(key string) string {
	return "https://devshared-ap-southeast-1-public-storage.s3.ap-southeast-1.amazonaws.com/" + key
}

type failingAvatarRepository struct {
	repositories.EmployeeRepository
	err error
}

func (r failingAvatarRepository) UpdateAvatarURL(context.Context, uuid.UUID, *string) error {
	return r.err
}

func TestUploadAvatarUsesAppPrefixAndCleansObjectWhenPersistenceFails(t *testing.T) {
	persistErr := errors.New("avatar persistence failed")
	uploader := &recordingAvatarUploader{}
	svc := &EmployeeService{
		emps:    failingAvatarRepository{err: persistErr},
		uploads: uploader,
	}
	pngHeader := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

	_, err := svc.uploadAvatar(context.Background(), uuid.New(), nil, pngHeader, "text/plain", ".PNG")

	require.ErrorIs(t, err, persistErr)
	assert.Equal(t, "hrm-app/avatars", uploader.subdir, "application namespace prevents cross-app object collisions")
	assert.Equal(t, "image/png", uploader.contentType, "storage receives the sniffed MIME type")
	assert.Equal(t, []string{uploader.uploadedURL}, uploader.deleted, "failed persistence must not leak the uploaded object")
}
```

- [ ] **Step 2: Run test and verify the old prefix and orphan behavior fail**

Run:

```powershell
go test ./internal/services -run '^TestUploadAvatarUsesAppPrefixAndCleansObjectWhenPersistenceFails$' -count=1 -v
```

Expected: FAIL because subdirectory is `avatars` and `uploader.deleted` is empty.

- [ ] **Step 3: Implement prefix and compensation cleanup**

Change the avatar constants in `internal/services/employee_service.go` to:

```go
const (
	maxAvatarBytes = 5 * 1024 * 1024
	avatarSubdir   = "hrm-app/avatars"
)
```

Replace the persistence block inside `uploadAvatar` with:

```go
	if err := s.emps.UpdateAvatarURL(ctx, employeeID, &url); err != nil {
		_ = s.uploads.Delete(ctx, url)
		return nil, err
	}
	if prev != nil && *prev != "" {
		_ = s.uploads.Delete(ctx, *prev)
	}
```

- [ ] **Step 4: Remove Supabase-shaped fakes and comments from active tests**

In `internal/services/employee_service_test.go`, add a `subdir` field to `fakeUploader`, record it in `Upload`, and use AWS-shaped fake URLs:

```go
type fakeUploader struct {
	uploadedURL string
	subdir      string
	deleted     []string
}

func (f *fakeUploader) Upload(_ context.Context, subdir, ext string, _ []byte, _ string) (string, error) {
	f.subdir = subdir
	f.uploadedURL = "https://fake-bucket.s3.ap-southeast-1.amazonaws.com/" + subdir + "/" + uuid.NewString() + ext
	return f.uploadedURL, nil
}

func (f *fakeUploader) Delete(_ context.Context, publicURL string) error {
	f.deleted = append(f.deleted, publicURL)
	return nil
}

func (f *fakeUploader) PublicURL(key string) string {
	return "https://fake-bucket.s3.ap-southeast-1.amazonaws.com/" + key
}
```

After the valid upload assertion in `TestEmployeeService_UpdateAvatar_ChecksImageType`, add:

```go
	assert.Equal(t, "hrm-app/avatars", up.subdir)
```

Change the `internal/services/skill_service_test.go` orphan comment to:

```go
	// A failed database write must not leak an orphan object in S3.
```

- [ ] **Step 5: Format and run focused tests**

Run:

```powershell
gofmt -w internal/services/avatar_upload_test.go internal/services/employee_service.go internal/services/employee_service_test.go internal/services/skill_service_test.go
go test ./internal/services -run 'TestUploadAvatarUsesAppPrefixAndCleansObjectWhenPersistenceFails|TestEmployeeService_UpdateAvatar' -count=1 -v
```

Expected: deterministic avatar unit test passes. DB-backed employee tests pass when `TEST_DATABASE_URL` exists; otherwise output must explicitly show each `SKIP`.

- [ ] **Step 6: Commit Task 2**

```powershell
git add -- internal/services/avatar_upload_test.go internal/services/employee_service.go internal/services/employee_service_test.go internal/services/skill_service_test.go
git diff --cached --check
git commit -m "feat(employees): store avatars under app prefix"
```

---

### Task 3: Add and run live AWS S3 verification

**Files:**
- Create: `internal/services/upload_service_integration_test.go`

**Interfaces:**
- Consumes: real AWS credentials from process environment and `RUN_AWS_S3_INTEGRATION=1`.
- Produces: proof that upload, public read, ownership parsing, and deletion work against configured AWS S3.

- [ ] **Step 1: Add opt-in live integration test**

Create `internal/services/upload_service_integration_test.go`:

```go
package services

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
)

func TestUploadServiceLiveAWS(t *testing.T) {
	if os.Getenv("RUN_AWS_S3_INTEGRATION") != "1" {
		t.Skip("set RUN_AWS_S3_INTEGRATION=1 to run live AWS S3 verification")
	}

	cfg, err := config.LoadStorageConfig()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	svc, err := NewUploadService(ctx, cfg)
	require.NoError(t, err)
	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	require.NoError(t, err)

	publicURL, err := svc.Upload(ctx, "hrm-app/avatars", ".png", png, "image/png")
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Delete(context.Background(), publicURL) })

	key := svc.objectPathFromURL(publicURL)
	require.True(t, strings.HasPrefix(key, "hrm-app/avatars/"))
	_, err = svc.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	})
	require.NoError(t, err)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(publicURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "configured bucket must permit public avatar reads")

	require.NoError(t, svc.Delete(ctx, publicURL))
	_, err = svc.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	})
	require.Error(t, err, "deleted probe object must no longer exist")
}
```

- [ ] **Step 2: Format and compile the gated test**

Run:

```powershell
gofmt -w internal/services/upload_service_integration_test.go
go test ./internal/services -run '^TestUploadServiceLiveAWS$' -count=1 -v
```

Expected: one explicit `SKIP` because live access was not enabled. This confirms compilation without contacting AWS.

- [ ] **Step 3: Load local environment without printing secrets and run the live probe**

Run in one PowerShell process:

```powershell
Get-Content -LiteralPath '.env' | ForEach-Object {
    if ($_ -match '^([^#=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
    }
}
$env:RUN_AWS_S3_INTEGRATION = '1'
go test ./internal/services -run '^TestUploadServiceLiveAWS$' -count=1 -v
```

Expected: PASS. A `403` public-read assertion means bucket policy is incompatible with public avatar URLs; an AWS API error means credentials, bucket, or region still needs correction. Do not call the feature complete until this probe passes.

- [ ] **Step 4: Commit Task 3**

```powershell
git add -- internal/services/upload_service_integration_test.go
git diff --cached --check
git commit -m "test(storage): add live AWS S3 probe"
```

---

### Task 4: Update active documentation and run final verification

**Files:**
- Modify: `CLAUDE.md:42`
- Modify: `docs/API-REFERENCE.md:1902`

**Interfaces:**
- Consumes: completed AWS-only implementation.
- Produces: current documentation with no active Supabase storage instructions.

- [ ] **Step 1: Update current storage documentation**

Change the `CLAUDE.md` technology row to:

```markdown
| File storage | AWS SDK Go v2 S3 (`aws-sdk-go-v2/service/s3`) using regional AWS endpoints (`STORAGE_*` env) |
```

Replace the upload-storage sentence in `docs/API-REFERENCE.md` with:

```markdown
Server trả URL public sau upload từ bucket AWS S3 đã cấu hình.
```

- [ ] **Step 2: Prove active Supabase behavior is gone**

Run:

```powershell
rg -n -i '(supabase|STORAGE_ENDPOINT|ProjectRef)' internal .env.example .env.docker.example CLAUDE.md docs/API-REFERENCE.md
rg -n '(BaseEndpoint|UsePathStyle)' internal --glob '!**/*_test.go'
```

Expected: neither command finds matches. Tests may assert nil/false AWS SDK options, but production code must not configure them. Historical files under `docs/superpowers`, `ba-requirements`, and verification archives are intentionally excluded.

- [ ] **Step 3: Run formatting and complete Go verification**

Run:

```powershell
gofmt -w internal/config/storage.go internal/config/storage_test.go internal/services/upload_service.go internal/services/upload_service_test.go internal/services/avatar_upload_test.go internal/services/employee_service.go internal/services/employee_service_test.go internal/services/skill_service_test.go internal/services/upload_service_integration_test.go
go test ./...
```

Expected: command exits 0. Do not summarize this as “all tests passed” until skip visibility is checked.

- [ ] **Step 4: Surface relevant skip status explicitly**

Run:

```powershell
go test ./internal/services -run 'TestEmployeeService_UpdateAvatar' -count=1 -v
```

Expected: DB-backed tests pass when a dedicated `TEST_DATABASE_URL` exists. Otherwise each test shows `SKIP`; record that limitation in final handoff. Never construct this value from the development database because these tests truncate tables.

- [ ] **Step 5: Review exact change scope**

Run:

```powershell
git diff --check
git status --short
git diff --stat HEAD~3..HEAD
```

Expected: only files listed in this plan plus this plan/spec documentation. Existing unrelated worktree changes remain unstaged and uncommitted.

- [ ] **Step 6: Commit documentation**

```powershell
git add -- CLAUDE.md docs/API-REFERENCE.md
git diff --cached --check
git commit -m "docs(storage): document AWS S3 uploads"
```

## Completion Checklist

- [ ] Live AWS probe uploads under `hrm-app/avatars`, reads publicly, and deletes successfully.
- [ ] AWS URL persists through existing avatar service flow tests.
- [ ] New object is deleted when employee persistence fails.
- [ ] Previous AWS avatar remains best-effort cleanup after successful persistence.
- [ ] No active Supabase storage code or configuration remains.
- [ ] Full test command exits 0; every relevant skip is reported.
- [ ] Unrelated user changes are untouched.
