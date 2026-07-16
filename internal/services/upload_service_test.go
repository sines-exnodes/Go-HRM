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
