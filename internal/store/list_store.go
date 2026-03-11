// Package store provieds CRUD with list operations
package store

import "log/slog"

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

func (l *List) Len() int {
	return len(l.items)
}

type Container struct {
	bucket map[string]*List
}

var cContainer = &Container{
	bucket: make(map[string]*List),
}

func GetContainer() *Container {
	return cContainer
}

func (c *Container) InitList(key string) *List {
	c.bucket[key] = NewList()
	return c.bucket[key]
}

func (c *Container) Append(key string, items ...string) int {
	list, ok := c.bucket[key]

	if !ok {
		list = c.InitList(key)
	}
	for i := range items {
		list.Append(items[i])
	}

	return list.Len()
}

func (c *Container) Len(key string) int {
	list, ok := c.bucket[key]
	if !ok {
		return 0
	}

	return list.Len()
}

func (c *Container) Get(key string, start, end int) []string {
	result := make([]string, 0)
	list, ok := c.bucket[key]
	if !ok {
		return result
	}

	size := c.Len(key)
	slog.Info("Array",
		slog.Any("size", size),
		slog.Any("start", start),
		slog.Any("end", end),
	)
	if end < 0 {
		end = size - (-1 * end)
	}
	if start < 0 {
		start = size - (-1 * start)
	}
	for i := start; i <= min(size-1, end); i++ {
		result = append(result, list.items[i])
	}

	return result
}
