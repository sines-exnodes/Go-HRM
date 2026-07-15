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
	smithyhttp "github.com/aws/smithy-go/transport/http"
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
	t.Cleanup(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCleanup()
		_ = svc.Delete(cleanupCtx, publicURL)
	})

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
	var responseErr *smithyhttp.ResponseError
	require.ErrorAs(t, err, &responseErr, "deleted probe lookup must return an AWS HTTP response error")
	require.Equal(t, http.StatusNotFound, responseErr.HTTPStatusCode(), "deleted probe object must return HTTP 404")
}
