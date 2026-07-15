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
