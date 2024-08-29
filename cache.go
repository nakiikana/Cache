package cache

import (
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear() // удаляет все ключи
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration) // добавляет ключ со сроком жизни ttl
	Get(key any) (value any, ok bool)
	Remove(key any)
}

type Cache struct {
	items map[any]any
	mu    sync.Mutex
	size  int
}

func New(N int64) *Cache {
	return &Cache{
		items: make(map[any]any),
	}
}

func (c *Cache) Cap() int {
	return c.size
}

func (c *Cache) Len() int {
	return len(c.items)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		delete(c.items, k)
	}
}

func (c *Cache) Remove(k any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, k)
}

func (c *Cache) Add(k, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = v
}
