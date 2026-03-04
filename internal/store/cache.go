package store

import (
	"strconv"
	"sync"
	"time"

	"github.com/archstrap/cache-server/pkg/model"
)

type CacheItem struct {
	item      any
	expiresAt time.Time
	valueType model.ValueType
}

func NewCacheItem(item any, expiresAt time.Time, valueType model.ValueType) *CacheItem {
	return &CacheItem{
		item:      item,
		expiresAt: expiresAt,
		valueType: valueType,
	}
}

func (c *CacheItem) Item() any {
	return c.item
}

func (c *CacheItem) IncrementBy(value int) (int, bool) {

	oldData, err := strconv.Atoi(c.Item().(string))
	if err != nil {
		return -1, false
	}

	updated := oldData + 1
	c.item = strconv.Itoa(updated)

	return updated, true
}

func (c *CacheItem) ValueType() model.ValueType {
	return c.valueType
}

func (c *CacheItem) IsExpired() bool {
	return !c.expiresAt.IsZero() && time.Now().After(c.expiresAt)
}

func (c *CacheItem) IsExpiredNow(now time.Time) bool {
	return !c.expiresAt.IsZero() && now.After(c.expiresAt)
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]*CacheItem
}

var (
	CacheStore *Cache = &Cache{
		data: make(map[string]*CacheItem),
	}
)

func GetCacheStore() *Cache {
	return CacheStore
}

func (c *Cache) RLock() {
	c.mu.RLock()
}
func (c *Cache) RUnlock() {
	c.mu.RUnlock()
}

func (c *Cache) Lock() {
	c.mu.Lock()
}
func (c *Cache) Unlock() {
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (*CacheItem, bool) {
	item, ok := c.data[key]
	return item, ok
}

func (c *Cache) Delete(key string) {
	delete(c.data, key)
}

func (c *Cache) Add(key string, item *CacheItem) *CacheItem {
	c.data[key] = item
	return item
}

func (c *Cache) GetData() map[string]*CacheItem {
	return c.data
}
