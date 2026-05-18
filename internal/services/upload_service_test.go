package services

import (
	"testing"

	"github.com/exnodes/hrm-api/internal/config"
)

func TestBuildPublicURL(t *testing.T) {
	cfg := config.StorageConfig{
		Endpoint: "https://xthlukcxeczusflabcwp.storage.supabase.co/storage/v1/s3",
		Bucket:   "hrm-uploads",
	}
	svc := &UploadService{cfg: cfg}
	got := svc.PublicURL("avatars/abc.png")
	want := "https://xthlukcxeczusflabcwp.supabase.co/storage/v1/object/public/hrm-uploads/avatars/abc.png"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractObjectPath(t *testing.T) {
	cfg := config.StorageConfig{
		Endpoint: "https://xthlukcxeczusflabcwp.storage.supabase.co/storage/v1/s3",
		Bucket:   "hrm-uploads",
	}
	svc := &UploadService{cfg: cfg}
	u := "https://xthlukcxeczusflabcwp.supabase.co/storage/v1/object/public/hrm-uploads/avatars/abc.png"
	if got := svc.objectPathFromURL(u); got != "avatars/abc.png" {
		t.Fatalf("got %q", got)
	}
	if got := svc.objectPathFromURL(""); got != "" {
		t.Fatalf("empty input must return empty, got %q", got)
	}
	if got := svc.objectPathFromURL("https://other.example.com/x.png"); got != "" {
		t.Fatalf("foreign url must return empty, got %q", got)
	}
}
