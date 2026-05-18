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

// Uploader is the storage-side abstraction consumed by employee/profile
// services. UploadService is the Supabase-S3-backed implementation.
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
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	return &UploadService{cfg: cfg, client: client}, nil
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
		return nil // foreign or empty URL — silent no-op
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
	ref := s.cfg.ProjectRef()
	return fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s", ref, s.cfg.Bucket, key)
}

func (s *UploadService) objectPathFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	marker := fmt.Sprintf("/object/public/%s/", s.cfg.Bucket)
	idx := strings.Index(u.Path, marker)
	if idx < 0 {
		return ""
	}
	return u.Path[idx+len(marker):]
}
