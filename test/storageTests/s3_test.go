package storageTests

import (
	"context"
	"testing"

	"dalista/dalista-be/integrations/storage"
)

// S3 URL building is pure (no network). Credentials resolve lazily on the first
// API call, so NewS3 succeeds here without AWS creds.
func TestS3URL(t *testing.T) {
	s, err := storage.NewS3(context.Background(), "us-east-1",
		"dalista-staging-media-123456789012",
		"https://dalista-staging-media-123456789012.s3.us-east-1.amazonaws.com/")
	if err != nil {
		t.Fatalf("NewS3: %v", err)
	}
	got := s.URL("items/abc/one.jpg")
	want := "https://dalista-staging-media-123456789012.s3.us-east-1.amazonaws.com/items/abc/one.jpg"
	if got != want {
		t.Fatalf("URL = %q, want %q", got, want)
	}
	// A leading slash on the key must not produce a double slash.
	if got := s.URL("/items/abc/two.jpg"); got != "https://dalista-staging-media-123456789012.s3.us-east-1.amazonaws.com/items/abc/two.jpg" {
		t.Fatalf("leading-slash URL wrong: %q", got)
	}
}
