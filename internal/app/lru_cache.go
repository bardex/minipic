package app

import (
	"encoding/hex"
	"hash/fnv"
	"path/filepath"
	"sync"
)

type LruCache struct {
	directory string
	capacity  int
	mu        sync.Mutex
	items     map[string]*ResponseCache
	front     *ResponseCache
	back      *ResponseCache
}

func NewLruCache(cacheDir string, maxEntities int) *LruCache {
	return &LruCache{
		directory: cacheDir,
		capacity:  maxEntities,
		items:     make(map[string]*ResponseCache),
	}
}

func (c *LruCache) GetItem(key string) *ResponseCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	key = c.getHash(key)
	if item, exists := c.items[key]; exists {
		return item
	}

	return &ResponseCache{
		key:      c.getHash(key),
		filename: c.getFilename(key),
	}
}

func (c *LruCache) PushFront(item *ResponseCache) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.front == item {
		return
	}

	if len(c.items) >= c.capacity {
		if c.back != nil {
			c.remove(c.back)
		}
	}
	if len(c.items) == 0 {
		c.front = item
		c.back = item
	} else {
		c.front.prev = item
		item.next = c.front
		c.front = item
	}
	c.items[item.key] = item
}

func (c *LruCache) remove(rc *ResponseCache) {
	if c.front == rc {
		c.front = rc.next
	}
	if c.back == rc {
		c.back = rc.prev
	}
	if rc.next != nil {
		rc.next.prev = rc.prev
	}
	if rc.prev != nil {
		rc.prev.next = rc.next
	}
	rc.next = nil
	rc.prev = nil

	delete(c.items, rc.key)

	rc.Remove()
}

func (c *LruCache) getFilename(key string) string {
	return filepath.Join(c.directory, c.getHash(key)) + ".cache"
}

func (c *LruCache) getHash(key string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
