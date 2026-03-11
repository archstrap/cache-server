// Package store provieds CRUD with list operations
package store

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
