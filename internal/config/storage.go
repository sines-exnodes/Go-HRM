package config

import (
	"fmt"
	"os"
	"strings"
)

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
}

func LoadStorageConfig() (StorageConfig, error) {
	c := StorageConfig{
		Endpoint:  os.Getenv("STORAGE_ENDPOINT"),
		AccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
		SecretKey: os.Getenv("STORAGE_SECRET_KEY"),
		Region:    os.Getenv("STORAGE_REGION"),
		Bucket:    os.Getenv("STORAGE_BUCKET"),
	}
	if c.Endpoint == "" || c.AccessKey == "" || c.SecretKey == "" || c.Bucket == "" {
		return c, fmt.Errorf("storage config: STORAGE_ENDPOINT/ACCESS_KEY/SECRET_KEY/BUCKET are required")
	}
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	return c, nil
}

// ProjectRef extracts the Supabase project ref from the endpoint host.
// e.g. https://xthlukcxeczusflabcwp.storage.supabase.co/... -> xthlukcxeczusflabcwp
func (c StorageConfig) ProjectRef() string {
	host := strings.TrimPrefix(strings.TrimPrefix(c.Endpoint, "https://"), "http://")
	parts := strings.SplitN(host, ".", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
