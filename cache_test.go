package cache

import (
	"testing"
	"time"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)
	if cache.Cap() != 2 {
		t.Errorf("Expected capacity 2, got %d", cache.Cap())
	}
}

func TestCache_AddAndGet(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}
}

func TestCache_Remove(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	cache.Add("key1", "value1")
	cache.Remove("key1")

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache length to be 0, got %d", cache.Len())
	}

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be cleared")
	}

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("Expected key2 to be cleared")
	}
}

func TestCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")
	cache.Add("key3", "value3")

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be evicted")
	}

	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected key2 to remain, got %v", value)
	}

	if value, ok := cache.Get("key3"); !ok || value != "value3" {
		t.Errorf("Expected key3 to remain, got %v", value)
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	cache := NewLRUCache(2, 500*time.Millisecond)

	cache.AddWithTTL("key1", "value1", 300*time.Millisecond)

	time.Sleep(400 * time.Millisecond)

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to expire")
	}
}

func TestCache_NoTTL(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	cache.Add("key1", "value1")

	time.Sleep(2 * time.Second)

	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected key1 to remain without TTL, got %v", value)
	}
}

func TestCache_StopTicker(t *testing.T) {
	cache := NewLRUCache(2, 500*time.Millisecond)

	cache.AddWithTTL("key1", "value1", 300*time.Millisecond)

	time.Sleep(400 * time.Millisecond)

	cache.finishWork()

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to expire")
	}
}

func TestCache_TTLExpirationWithMultipleItems(t *testing.T) {
	cache := NewLRUCache(2, 500*time.Millisecond)

	cache.AddWithTTL("key1", "value1", 200*time.Millisecond)
	cache.AddWithTTL("key2", "value2", 300*time.Millisecond)

	time.Sleep(250 * time.Millisecond)

	// key1 должен быть удален
	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to expire")
	}

	// key2 должен быть доступен
	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected key2 to remain, got %v", value)
	}
}

func TestCache_StopTickerWhileItemsAreExpired(t *testing.T) {
	cache := NewLRUCache(2, 500*time.Millisecond)

	cache.AddWithTTL("key1", "value1", 200*time.Millisecond)
	cache.AddWithTTL("key2", "value2", 200*time.Millisecond)

	time.Sleep(250 * time.Millisecond)

	// остановка проверки на истечение срока работы

	cache.finishWork()

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to expire")
	}

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("Expected key2 to expire")
	}
}

func TestCache_SliceAsKey(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Second)

	key := []int{1, 2, 3}
	value := "value"

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when using a slice as a key, but did not get one")
		}
	}()
	cache.Add(key, value)
}

func TestCache_ClearMap(t *testing.T) {
	cache := NewLRUCache(3, time.Second*1)

	cache.Add("key1", 1)
	cache.Add("key2", 2)
	cache.Add("key3", 3)
	cache.Add("key4", 4)
	cache.AddWithTTL("key5", 5, time.Millisecond*100)

	if cache.Len() != 3 {
		t.Errorf("Expected cache length to be 3 before clearing, got %d", cache.Len())
	}

	time.Sleep(time.Second * 2)

	if val, ok := cache.items["key5"]; ok {
		t.Errorf("Expected delete from the map - found extra value: %v", val)
	}

	if val, ok := cache.items["key1"]; ok {
		t.Errorf("Expected delete from the map - found extra value: %v", val)
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache length to be 0 after clearing, got %d", cache.Len())
	}

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be cleared")
	}

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("Expected key2 to be cleared")
	}

	if _, ok := cache.Get("key3"); ok {
		t.Errorf("Expected key3 to be cleared")
	}
}

func TestCache_DifferentKeyTypes(t *testing.T) {
	cache := NewLRUCache(5, 1*time.Second)

	cache.Add("key1", "value1")
	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	cache.Add(42, "value42")
	if value, ok := cache.Get(42); !ok || value != "value42" {
		t.Errorf("Expected value42, got %v", value)
	}

	type myStruct struct {
		field1 string
		field2 int
	}
	keyStruct := myStruct{"test", 123}
	cache.Add(keyStruct, "valueStruct")
	if value, ok := cache.Get(keyStruct); !ok || value != "valueStruct" {
		t.Errorf("Expected valueStruct, got %v", value)
	}

	keyPointer := &myStruct{"pointer", 456}
	cache.Add(keyPointer, "valuePointer")
	if value, ok := cache.Get(keyPointer); !ok || value != "valuePointer" {
		t.Errorf("Expected valuePointer, got %v", value)
	}

}

func TestCache_UltimateExpiration(t *testing.T) {
	cache := NewLRUCache(3, 100*time.Millisecond)

	cache.AddWithTTL("key1", "value1", 150*time.Millisecond)
	cache.AddWithTTL("key2", "value2", 300*time.Millisecond)
	cache.AddWithTTL("key3", "value3", 450*time.Millisecond)

	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}
	if value, ok := cache.Get("key3"); !ok || value != "value3" {
		t.Errorf("Expected value3, got %v", value)
	}

	time.Sleep(200 * time.Millisecond)

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to expire, but it is still present")
	}

	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected value2 to remain, got %v", value)
	}
	if value, ok := cache.Get("key3"); !ok || value != "value3" {
		t.Errorf("Expected value3 to remain, got %v", value)
	}

	time.Sleep(200 * time.Millisecond)

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("Expected key2 to expire, but it is still present")
	}

	if value, ok := cache.Get("key3"); !ok || value != "value3" {
		t.Errorf("Expected value3 to remain, got %v", value)
	}

	time.Sleep(200 * time.Millisecond)

	if _, ok := cache.Get("key3"); ok {
		t.Errorf("Expected key3 to expire, but it is still present")
	}

	if cache.Len() != 0 {
		t.Errorf("Expected cache to be empty, but it has %d items", cache.Len())
	}

	if len(cache.items) > 0 {
		t.Errorf("Map to be empty, but it has %d items", len(cache.items))
	}
}
