package main

import (
	"sync"
	"time"
)

type cacheEntry struct {
	item      *ContentItem
	expiresAt time.Time
}

type cache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (c *cache) get(path string) (*ContentItem, bool) {
	c.mu.RLock()
	entry, ok := c.items[path]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	// Expired entries are removed lazily to avoid a separate cleanup loop.
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		// Check again under write lock to avoid races with set().
		if latest, stillPresent := c.items[path]; stillPresent && time.Now().After(latest.expiresAt) {
			delete(c.items, path)
		}
		c.mu.Unlock()
		return nil, false
	}

	return entry.item, true
}

func (c *cache) set(path string, item *ContentItem) {
	c.mu.Lock()
	c.items[path] = cacheEntry{item: item, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
