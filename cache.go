package cache

import (
	"context"
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

type Item struct {
	value any
	ttl   time.Duration
}

type Cache struct {
	items map[any]Item
	mu    *sync.Mutex
	size  int
}

func New(N int) *Cache {
	return &Cache{
		items: make(map[any]Item),
		mu:    &sync.Mutex{},
		size:  N,
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
	for key := range c.items {
		delete(c.items, key)
	}
}

func (c *Cache) Remove(key any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache) Add(key, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item{value: value}
}

func (c *Cache) AddWithTTL(key, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{value: value, ttl: ttl}
	go func(key any, ttl time.Duration) {
		ctx, cancel := context.WithTimeout(context.Background(), ttl)
		defer cancel()
		select {
		case <-ctx.Done():
			c.Remove(key)
		}
	}(key, ttl)
}

func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok = c.items[key]
	return
}
