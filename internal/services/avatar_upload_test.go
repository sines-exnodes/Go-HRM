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
