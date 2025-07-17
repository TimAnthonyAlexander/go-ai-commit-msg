package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Cache represents a file-based cache
type Cache struct {
	baseDir string
}

// NewCache creates a new cache instance
func NewCache(gitDir string) *Cache {
	cacheDir := filepath.Join(gitDir, ".git", "gh-smart-commit-cache")
	return &Cache{baseDir: cacheDir}
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (string, bool, error) {
	if err := c.ensureCacheDir(); err != nil {
		return "", false, err
	}

	filePath := c.getFilePath(key)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var entry CacheEntry
	if err := json.NewDecoder(file).Decode(&entry); err != nil {
		return "", false, fmt.Errorf("failed to decode cache entry: %w", err)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Clean up expired entry
		os.Remove(filePath)
		return "", false, nil
	}

	return entry.Value, true, nil
}

// Set stores a value in cache
func (c *Cache) Set(key, value string, ttl time.Duration) error {
	if err := c.ensureCacheDir(); err != nil {
		return err
	}

	entry := CacheEntry{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	filePath := c.getFilePath(key)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(entry); err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *Cache) Delete(key string) error {
	filePath := c.getFilePath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}
	return nil
}

// Clear removes all cache entries
func (c *Cache) Clear() error {
	if _, err := os.Stat(c.baseDir); os.IsNotExist(err) {
		return nil // Nothing to clear
	}

	return os.RemoveAll(c.baseDir)
}

// ensureCacheDir creates the cache directory if it doesn't exist
func (c *Cache) ensureCacheDir() error {
	return os.MkdirAll(c.baseDir, 0755)
}

// getFilePath returns the file path for a given cache key
func (c *Cache) getFilePath(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(c.baseDir, filename)
}

// GenerateCacheKey creates a cache key from multiple components
func GenerateCacheKey(components ...string) string {
	h := sha256.New()
	for _, component := range components {
		io.WriteString(h, component)
		io.WriteString(h, "|") // separator
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
