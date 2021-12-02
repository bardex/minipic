package app

import (
	"encoding/hex"
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

func (c *LruCache) GetAndWriteTo(key string, w http.ResponseWriter) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// use a permanent key instead of the built-in, for to restore cache from files after restarting (todo).
	internalKey := c.getHash(key)
	item, exists := c.items[internalKey]
	if !exists {
		return false, nil
	}
	if err := item.WriteTo(w); err != nil {
		return false, err
	}
	c.pushFront(item)

	return true, nil
}

func (c *LruCache) Save(key string, headers http.Header, body []byte) error {
	rc := c.GetOrCreateItem(key)
	if err := rc.Save(headers, body); err != nil {
		return err
	}
	c.pushFront(rc)
	return nil
}

func (c *LruCache) GetOrCreateItem(key string) *ResponseCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	// use a permanent key instead of the built-in, for to restore cache from files after restarting (todo).
	internalKey := c.getHash(key)
	if item, exists := c.items[internalKey]; exists {
		return item
	}

	return &ResponseCache{
		key:      internalKey,
		filename: filepath.Join(c.directory, internalKey) + ".cache",
	}
}

func (c *LruCache) pushFront(item *ResponseCache) {
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

func (c *LruCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, rc := range c.items {
		if err := c.remove(rc); err != nil {
			return err
		}
	}
	return nil
}

func (c *LruCache) Len() int {
	return len(c.items)
}

func (c *LruCache) Cap() int {
	return c.capacity
}

func (c *LruCache) remove(rc *ResponseCache) error {
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

	return rc.Remove()
}

func (c *LruCache) getHash(key string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
