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
	c.mu.Lock()
	defer c.mu.Unlock()

	// use a permanent key instead of the built-in, for to restore cache from files after restarting (todo).
	internalKey := c.getHash(key)

	item, exists := c.items[internalKey]
	if !exists {
		return false, nil
	}

	if err := item.writeTo(w); err != nil {
		return false, err
	}

	return true, c.pushFront(item)
}

func (c *LruCache) Save(key string, headers http.Header, body []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	internalKey := c.getHash(key)
	item, exists := c.items[internalKey]

	if !exists {
		if err := c.removeLast(); err != nil {
			return err
		}
		item = &ResponseCache{
			key:      internalKey,
			filepath: filepath.Join(c.directory, internalKey) + ".cache",
		}
		c.items[internalKey] = item
	}

	if err := item.save(headers, body); err != nil {
		return err
	}

	return c.pushFront(item)
}

func (c *LruCache) removeLast() error {
	if len(c.items) >= c.capacity && c.back != nil {
		last := c.back
		c.back = last.prev
		if c.front == last {
			c.front = last.next
		}
		last.next = nil
		last.prev = nil

		delete(c.items, last.key)

		if err := last.remove(); err != nil {
			return err
		}
	}
	return nil
}

func (c *LruCache) pushFront(item *ResponseCache) error {
	if c.front == item {
		return nil
	}
	if c.front == nil {
		c.front = item
		c.back = item
	} else {
		c.front.prev = item
		item.next = c.front
		c.front = item
	}
	return nil
}

func (c *LruCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, rc := range c.items {
		delete(c.items, i)
		if err := rc.remove(); err != nil {
			return err
		}
	}
	return nil
}

func (c *LruCache) getHash(key string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
