package helpers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheItem represents a cached response.
type CacheItem struct {
	Headers http.Header
	Status  int
	Body    []byte
}

// Cache represents a simple in-memory cache.
type Cache struct {
	cache *cache.Cache
}

// NewCache initializes a new Cache.
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set adds an item to the cache.
func (c *Cache) Set(key string, item CacheItem, ttl time.Duration) {
	c.cache.Set(key, item, ttl)
}

// Get retrieves an item from the cache.
func (c *Cache) Get(key string) (CacheItem, bool) {
	item, found := c.cache.Get(key)
	if !found {
		return CacheItem{}, false
	}
	return item.(CacheItem), true
}

// GenerateCacheKey generates a cache key based on the provided inputs.
func GenerateCacheKey(inputs ...string) string {
	combined := strings.Join(inputs, "")
	return fmt.Sprintf("%x", sha256.Sum256([]byte(combined)))
}
