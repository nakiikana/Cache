package cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear()
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration)
	Get(key any) (value any, ok bool)
	Remove(key any)
}

type Item struct {
	key   any
	value any
	ttl   time.Time
}

type Cache struct {
	items          map[any]*list.Element
	mu             sync.RWMutex
	cap            int
	lruQueue       *list.List
	done           chan struct{}
	cleanFrequency time.Duration
}

const noEviction time.Duration = 1<<63 - 1 //about 292 years

func NewLRUCache(N int, f time.Duration) *Cache {
	c := &Cache{
		items:          make(map[any]*list.Element),
		mu:             sync.RWMutex{},
		cap:            N,
		lruQueue:       list.New(),
		done:           make(chan struct{}),
		cleanFrequency: f,
	}
	go c.checkIfExpired()
	return c
}

func (c *Cache) Cap() int {
	return c.cap
}

func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.lruQueue.Len()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[any]*list.Element)
	c.lruQueue.Init()
}

func (c *Cache) Add(key, value any) {
	c.AddWithTTL(key, value, 0)
}

func (c *Cache) AddWithTTL(key, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if ttl == 0 {
		ttl = noEviction
	}
	if elem, ok := c.items[key]; ok {
		c.lruQueue.MoveToFront(elem)
		elem.Value.(*Item).value = value
		elem.Value.(*Item).ttl = now.Add(ttl)
		return
	}

	if c.lruQueue.Len() == c.cap {
		c.removeLRU()
	}

	item := &Item{
		key:   key,
		value: value,
		ttl:   now.Add(ttl),
	}
	newElem := c.lruQueue.PushFront(item)
	c.items[key] = newElem
}

func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		if time.Now().After(elem.Value.(*Item).ttl) {
			c.unsafeRemove(key)
			return nil, false
		}
		c.lruQueue.MoveToFront(elem)
		return elem.Value.(*Item).value, true
	}
	return nil, false
}

func (c *Cache) Remove(key any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unsafeRemove(key)
}

func (c *Cache) unsafeRemove(key any) {
	if elem, ok := c.items[key]; ok {
		delete(c.items, key)
		c.lruQueue.Remove(elem)
	}
}

func (c *Cache) removeLRU() {
	if elem := c.lruQueue.Back(); elem != nil {
		delete(c.items, elem.Value.(*Item).key)
		c.lruQueue.Remove(elem)
	}
}

func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, elem := range c.items {
		if now.After(elem.Value.(*Item).ttl) {
			log.Println("Removing expired item with key: ", key)
			c.unsafeRemove(key)
		}
	}
}

func (c *Cache) checkIfExpired() {
	tick := time.NewTicker(c.cleanFrequency)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			c.removeExpired()
		case <-c.done:
			return
		}
	}
}

func (c *Cache) finishWork() {
	close(c.done)
}
