package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache("/tmp/test-repo")
	expectedDir := filepath.Join("/tmp/test-repo", ".git", "gh-smart-commit-cache")

	if cache.baseDir != expectedDir {
		t.Errorf("Expected baseDir %s, got %s", expectedDir, cache.baseDir)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewCache(tmpDir)

	// Test setting and getting a value
	key := "test-key"
	value := "test-value"
	ttl := 1 * time.Hour

	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}

	retrievedValue, found, err := cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache value: %v", err)
	}

	if !found {
		t.Error("Expected to find cached value")
	}

	if retrievedValue != value {
		t.Errorf("Expected value %s, got %s", value, retrievedValue)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewCache(tmpDir)

	_, found, err := cache.Get("non-existent-key")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if found {
		t.Error("Expected not to find non-existent key")
	}
}

func TestCacheExpiration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewCache(tmpDir)

	key := "expiring-key"
	value := "expiring-value"
	ttl := 50 * time.Millisecond

	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, found, err := cache.Get(key)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if found {
		t.Error("Expected expired key not to be found")
	}
}

func TestCacheDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewCache(tmpDir)

	key := "delete-key"
	value := "delete-value"
	ttl := 1 * time.Hour

	// Set a value
	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}

	// Verify it exists
	_, found, _ := cache.Get(key)
	if !found {
		t.Error("Expected to find the value before deletion")
	}

	// Delete it
	err = cache.Delete(key)
	if err != nil {
		t.Errorf("Failed to delete cache value: %v", err)
	}

	// Verify it's gone
	_, found, _ = cache.Get(key)
	if found {
		t.Error("Expected not to find the value after deletion")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewCache(tmpDir)

	// Set multiple values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err = cache.Set(key, "value", 1*time.Hour)
		if err != nil {
			t.Errorf("Failed to set cache value: %v", err)
		}
	}

	// Clear cache
	err = cache.Clear()
	if err != nil {
		t.Errorf("Failed to clear cache: %v", err)
	}

	// Verify all values are gone
	for _, key := range keys {
		_, found, _ := cache.Get(key)
		if found {
			t.Errorf("Expected key %s not to be found after clearing cache", key)
		}
	}
}

func TestGenerateCacheKey(t *testing.T) {
	// Test that same components generate same key
	key1 := GenerateCacheKey("component1", "component2", "component3")
	key2 := GenerateCacheKey("component1", "component2", "component3")

	if key1 != key2 {
		t.Error("Expected same components to generate same cache key")
	}

	// Test that different components generate different keys
	key3 := GenerateCacheKey("component1", "component2", "different")

	if key1 == key3 {
		t.Error("Expected different components to generate different cache keys")
	}

	// Test that key is not empty
	if key1 == "" {
		t.Error("Expected cache key to not be empty")
	}
}
