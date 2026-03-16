// Package store provieds CRUD with list operations
package store

import (
	"sync"
	"time"
)

type List struct {
	items []string
}

func NewList() *List {
	return &List{
		items: make([]string, 0),
	}
}

func (l *List) Append(item string) {
	l.items = append(l.items, item)
}

func (l *List) Prepend(item string) {
	l.items = append([]string{item}, l.items...)
}

func (l *List) Len() int {
	return len(l.items)
}

type Container struct {
	bucket map[string]*List
	lock   sync.RWMutex
	cond   *sync.Cond
}

var cContainer = &Container{
	bucket: make(map[string]*List),
}

func init() {
	cContainer.cond = sync.NewCond(&cContainer.lock)
}

func GetContainer() *Container {
	return cContainer
}

func (c *Container) InitList(key string) *List {
	c.bucket[key] = NewList()
	return c.bucket[key]
}

func (c *Container) Append(key string, items ...string) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	list, ok := c.bucket[key]

	if !ok {
		list = c.InitList(key)
	}
	for i := range items {
		list.Append(items[i])
	}

	c.cond.Broadcast()
	return list.Len()
}

func (c *Container) Prepend(key string, items ...string) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	list, ok := c.bucket[key]

	if !ok {
		list = c.InitList(key)
	}

	for i := range items {
		list.Prepend(items[i])
	}

	return list.Len()
}

func (c *Container) Get(key string, start, end int) []string {
	result := make([]string, 0)
	list, ok := c.bucket[key]
	if !ok {
		return result
	}

	size := list.Len()
	if end < 0 {
		end = size - (-1 * end)
	}
	if start < 0 {
		start = size - (-1 * start)
	}
	for i := max(0, start); i <= min(size-1, end); i++ {
		result = append(result, list.items[i])
	}

	return result
}

func (c *Container) GetLen(key string) int {
	list, ok := c.bucket[key]
	if !ok {
		return 0
	}

	return list.Len()
}

func (c *Container) DeleteWithLock(key string, count int) []string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.Delete(key, count)
}

func (c *Container) Delete(key string, count int) []string {
	length := c.GetLen(key)

	if length == 0 {
		return []string{}
	}

	list := c.bucket[key]
	deletedItems := make([]string, 0)
	for i := 0; i < min(length, count); i++ {
		deletedItems = append(deletedItems, list.items[i])
	}

	deleteCount := len(deletedItems)
	list.items = list.items[deleteCount:]
	return deletedItems
}

func (c *Container) BlockDelete(key string, timeOut int) []string {
	c.lock.Lock()
	defer c.lock.Unlock()

	var deadLine time.Time

	if timeOut != 0 {
		duration := time.Duration(timeOut) * time.Millisecond
		deadLine = time.Now().Add(duration)
		go func() {
			time.Sleep(duration)
			c.cond.Broadcast()
		}()
	}

	for {
		deletedItems := c.Delete(key, 1)

		if len(deletedItems) > 0 {
			return append([]string{key}, deletedItems...)
		}

		if timeOut != 0 && time.Now().After(deadLine) {
			return nil
		}

		c.cond.Wait()
	}
}
