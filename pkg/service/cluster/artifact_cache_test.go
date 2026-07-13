package cluster

import (
	"testing"
)

func TestArtifactCache_GetSet(t *testing.T) {
	cache, err := newArtifactCache(10)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	bucket := "test-bucket"
	key := "test-key"
	content := []byte("test content")

	// Cache miss on first get
	_, found := cache.Get(bucket, key)
	if found {
		t.Error("expected cache miss, got hit")
	}

	// Set content
	cache.Set(bucket, key, content)

	// Cache hit on second get
	retrieved, found := cache.Get(bucket, key)
	if !found {
		t.Error("expected cache hit, got miss")
	}
	if string(retrieved) != string(content) {
		t.Errorf("expected content %q, got %q", content, retrieved)
	}
}

func TestArtifactCache_MultipleEntries(t *testing.T) {
	cache, err := newArtifactCache(10)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Add multiple entries
	entries := map[string][]byte{
		"bucket1/key1": []byte("content1"),
		"bucket1/key2": []byte("content2"),
		"bucket2/key1": []byte("content3"),
	}

	for key, content := range entries {
		cache.Set("bucket", key, content)
	}

	// Verify all entries are cached
	for key, expectedContent := range entries {
		retrieved, found := cache.Get("bucket", key)
		if !found {
			t.Errorf("expected cache hit for key %q", key)
		}
		if string(retrieved) != string(expectedContent) {
			t.Errorf("for key %q, expected %q, got %q", key, expectedContent, retrieved)
		}
	}
}

func TestArtifactCache_LRUEviction(t *testing.T) {
	cache, err := newArtifactCache(2) // Small cache for testing eviction
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Fill cache to capacity
	cache.Set("bucket", "key1", []byte("content1"))
	cache.Set("bucket", "key2", []byte("content2"))

	// Add one more entry, should evict oldest
	cache.Set("bucket", "key3", []byte("content3"))

	// key1 should have been evicted
	_, found := cache.Get("bucket", "key1")
	if found {
		t.Error("expected key1 to be evicted")
	}

	// key2 and key3 should still be present
	_, found = cache.Get("bucket", "key2")
	if !found {
		t.Error("expected key2 to be present")
	}

	_, found = cache.Get("bucket", "key3")
	if !found {
		t.Error("expected key3 to be present")
	}
}
