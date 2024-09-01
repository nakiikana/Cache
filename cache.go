package cache

import (
	"container/list"
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
	key   any
	value any
	ttl   time.Duration
}

type Cache struct {
	items    map[any]*list.Element
	mu       *sync.Mutex
	cap      int
	lruQueue *list.List
}

func New(N int) *Cache {
	return &Cache{
		items:    make(map[any]*list.Element),
		mu:       &sync.Mutex{},
		cap:      N,
		lruQueue: list.New(),
	}
}

func (c *Cache) Cap() int {
	return c.cap
}

func (c *Cache) Len() int {
	return c.lruQueue.Len()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[any]*list.Element)
	c.lruQueue.Init()
}

func (c *Cache) Add(key, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.lruQueue.MoveToFront(elem)
		elem.Value.(*Item).value = value
		return
	}
	if c.lruQueue.Len() == c.cap {
		c.removeLRU()
	}
	item := &Item{
		key:   key,
		value: value,
		ttl:   0,
	}
	newElem := c.lruQueue.PushFront(item)
	c.items[key] = newElem
}

func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.lruQueue.MoveToFront(elem)
		return elem.Value.(*Item).value, true
	}
	return nil, false
}

func (c *Cache) Remove(key any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		delete(c.items, key)
		c.lruQueue.Remove(elem)
	}
}

func (c *Cache) removeLRU() {
	if elem := c.lruQueue.Back(); elem != nil {
		delete(c.items, elem.Value.(*Item).key)
		c.lruQueue.Remove(c.lruQueue.Back())
	}
}

// func (c *Cache) AddWithTTL(key, value any, ttl time.Duration) {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	c.items[key] = item{value: value, ttl: ttl}
// 	go func(key any, ttl time.Duration) {
// 		ctx, cancel := context.WithTimeout(context.Background(), ttl)
// 		defer cancel()
// 		select {
// 		case <-ctx.Done():
// 			c.Remove(key)
// 		}
// 	}(key, ttl)
// }
