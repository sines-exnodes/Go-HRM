package config

// This file anchors the aws-sdk-go-v2 modules as direct dependencies so
// `go mod tidy` retains them ahead of the S3 storage client that consumes
// them in a later Phase 2 task. Remove these blank imports once the storage
// client imports the packages directly.
import (
	_ "github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/credentials"
	_ "github.com/aws/aws-sdk-go-v2/service/s3"
)
