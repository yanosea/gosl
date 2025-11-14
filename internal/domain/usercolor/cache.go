package usercolor

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

// Cache interface for managing user color mappings
type Cache interface {
	// Get retrieves color for userID from cache.
	// Returns:
	//   - color: AdaptiveColor if exists
	//   - ok: true if cache hit, false otherwise
	Get(userID string) (color AdaptiveColor, ok bool)

	// Set stores color for userID in cache.
	// Evicts least recently used entry if cache is full.
	Set(userID string, color AdaptiveColor)

	// Clear removes all cached entries.
	// Should be called when message cache is invalidated.
	Clear()

	// Len returns current number of cached entries.
	Len() int
}

// lruCache implements Cache interface using golang-lru/v2
type lruCache struct {
	cache *lru.Cache[string, AdaptiveColor]
}

// NewUserColorCache creates a new LRU cache with the specified size.
// Size must be greater than 0.
func NewUserColorCache(size int) Cache {
	cache, err := lru.New[string, AdaptiveColor](size)
	if err != nil {
		// This should never happen as long as size > 0
		panic(err)
	}

	return &lruCache{
		cache: cache,
	}
}

// Get retrieves color for userID from cache
func (c *lruCache) Get(userID string) (AdaptiveColor, bool) {
	return c.cache.Get(userID)
}

// Set stores color for userID in cache
func (c *lruCache) Set(userID string, color AdaptiveColor) {
	c.cache.Add(userID, color)
}

// Clear removes all cached entries
func (c *lruCache) Clear() {
	c.cache.Purge()
}

// Len returns current number of cached entries
func (c *lruCache) Len() int {
	return c.cache.Len()
}
