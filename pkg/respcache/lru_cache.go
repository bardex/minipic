package respcache

import (
	"encoding/hex"
	"errors"
	"hash/fnv"
	"net/http"
	"path/filepath"
	"sync"
)

type LruCache struct {
	directory string
	capacity  int
	mu        sync.RWMutex
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

func (c *LruCache) CreateResponseCache(key string) *ResponseCache {
	return &ResponseCache{
		key:      key,
		filename: c.getFilename(key),
	}
}

func (c *LruCache) Read(key string, w http.ResponseWriter) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rc, exists := c.items[key]
	if !exists {
		return false, nil
	}

	err := rc.Read(w)
	if errors.Is(err, ErrCacheFileNotExists) {
		c.remove(rc)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	c.PushFront(rc)
	return true, nil
}

func (c *LruCache) PushFront(rc *ResponseCache) {
	if c.front == rc {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.capacity {
		last := c.back
		if last != nil {
			c.remove(last)
		}
	}

	if len(c.items) == 0 {
		c.front = rc
		c.back = rc
	} else {
		c.front.prev = rc
		rc.next = c.front
		c.front = rc
	}
	c.items[rc.key] = rc
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
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	key = hex.EncodeToString(hasher.Sum(nil))
	return filepath.Join(c.directory, key) + ".cache"
}
