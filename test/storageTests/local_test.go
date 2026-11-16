package storageTests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dalista/dalista-be/integrations/storage"
)

func TestLocalPutURLDelete(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocal(dir, "http://localhost:8080/media/")
	ctx := context.Background()

	key := "items/abc/one.png"
	if err := s.Put(ctx, key, strings.NewReader("hello"), 5, "image/png"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	// File landed under the base dir at the key path.
	if b, err := os.ReadFile(filepath.Join(dir, "items", "abc", "one.png")); err != nil || string(b) != "hello" {
		t.Fatalf("stored file wrong: %q err=%v", b, err)
	}
	// URL joins base + key (trailing slash on base trimmed).
	if got := s.URL(key); got != "http://localhost:8080/media/items/abc/one.png" {
		t.Fatalf("URL = %q", got)
	}
	// Delete removes it; deleting again is not an error.
	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "items", "abc", "one.png")); !os.IsNotExist(err) {
		t.Fatalf("file not removed: %v", err)
	}
	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("Delete of missing object should be nil, got %v", err)
	}
}

func TestLocalRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocal(dir, "http://x/media")
	err := s.Put(context.Background(), "../escape.png", strings.NewReader("x"), 1, "image/png")
	if err == nil {
		t.Fatal("expected traversal key to be rejected")
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(dir), "escape.png")); statErr == nil {
		t.Fatal("traversal wrote outside the base dir")
	}
}
