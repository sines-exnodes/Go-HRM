package services

import (
	"context"
	"encoding/base64"
	"errors"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

type recordingAvatarUploader struct {
	subdir      string
	contentType string
	uploadedURL string
	deleted     []string
	deleteErr   error
	deleteLimit bool
}

func (u *recordingAvatarUploader) Upload(_ context.Context, subdir, ext string, _ []byte, contentType string) (string, error) {
	u.subdir = subdir
	u.contentType = contentType
	u.uploadedURL = u.PublicURL(path.Join(subdir, "new-avatar"+ext))
	return u.uploadedURL, nil
}

func (u *recordingAvatarUploader) Delete(ctx context.Context, publicURL string) error {
	u.deleteErr = ctx.Err()
	_, u.deleteLimit = ctx.Deadline()
	if u.deleteErr != nil {
		return u.deleteErr
	}
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
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.uploadAvatar(ctx, uuid.New(), nil, pngHeader, "text/plain", ".PNG")

	require.ErrorIs(t, err, persistErr)
	assert.Equal(t, "hrm-app/avatars", uploader.subdir, "application namespace prevents cross-app object collisions")
	assert.Equal(t, "image/png", uploader.contentType, "storage receives the sniffed MIME type")
	assert.NoError(t, uploader.deleteErr, "request cancellation must not cancel compensation cleanup")
	assert.True(t, uploader.deleteLimit, "compensation cleanup must have a finite deadline")
	assert.Equal(t, []string{uploader.uploadedURL}, uploader.deleted, "failed persistence must not leak the uploaded object")
}

func TestUploadAvatarRejectsGIF(t *testing.T) {
	uploader := &recordingAvatarUploader{}
	svc := &EmployeeService{
		emps:    failingAvatarRepository{err: errors.New("unexpected avatar persistence")},
		uploads: uploader,
	}
	gifPayload, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==")
	require.NoError(t, err)

	_, err = svc.uploadAvatar(context.Background(), uuid.New(), nil, gifPayload, "image/gif", ".gif")

	ae, ok := apperrors.As(err)
	require.True(t, ok, "expected AppError, got %v", err)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
	assert.Equal(t, "Avatar must be a valid image (PNG, JPEG, or WEBP)", ae.Message)
	assert.Empty(t, uploader.uploadedURL, "rejected GIF must not reach storage")
}
