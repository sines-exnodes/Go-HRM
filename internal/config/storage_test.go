package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStorageConfigRequiresEveryAWSValue(t *testing.T) {
	for _, missing := range []string{
		"STORAGE_ACCESS_KEY",
		"STORAGE_SECRET_KEY",
		"STORAGE_REGION",
		"STORAGE_BUCKET",
	} {
		t.Run(missing, func(t *testing.T) {
			t.Setenv("STORAGE_ACCESS_KEY", "access")
			t.Setenv("STORAGE_SECRET_KEY", "secret")
			t.Setenv("STORAGE_REGION", "ap-southeast-1")
			t.Setenv("STORAGE_BUCKET", "bucket")
			t.Setenv(missing, "")

			_, err := LoadStorageConfig()

			require.EqualError(t, err, "storage config: STORAGE_ACCESS_KEY/STORAGE_SECRET_KEY/STORAGE_REGION/STORAGE_BUCKET are required")
		})
	}
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
