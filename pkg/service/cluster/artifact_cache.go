package cluster

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stackrox/infra/pkg/service/metrics"
)

const (
	// Default cache size: estimate 10KB per artifact × 100 clusters × 2 artifacts = ~2MB
	// Setting to 1000 entries to allow for growth
	defaultCacheSize = 1000
)

// artifactCache provides thread-safe LRU caching of immutable GCS artifact contents.
// Since workflow artifacts don't change once written, no TTL is needed.
type artifactCache struct {
	cache *lru.Cache[string, []byte]
}

// newArtifactCache creates a new artifact cache with the specified size.
func newArtifactCache(size int) (*artifactCache, error) {
	cache, err := lru.New[string, []byte](size)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	return &artifactCache{
		cache: cache,
	}, nil
}

// Get retrieves cached artifact content if present.
// Returns the content and true if found, nil and false otherwise.
func (c *artifactCache) Get(bucket, key string) ([]byte, bool) {
	cacheKey := makeCacheKey(bucket, key)
	content, found := c.cache.Get(cacheKey)
	if !found {
		metrics.ArtifactCacheMissesCounter.Inc()
		return nil, false
	}

	metrics.ArtifactCacheHitsCounter.Inc()
	return content, true
}

// Set stores artifact content in the cache.
func (c *artifactCache) Set(bucket, key string, content []byte) {
	cacheKey := makeCacheKey(bucket, key)
	c.cache.Add(cacheKey, content)
	metrics.ArtifactCacheSizeGauge.Set(float64(c.cache.Len()))
}

// makeCacheKey creates a unique cache key from bucket and object key.
func makeCacheKey(bucket, key string) string {
	return fmt.Sprintf("%s/%s", bucket, key)
}
