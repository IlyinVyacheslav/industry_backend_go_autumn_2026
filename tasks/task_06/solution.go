package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, items: make(map[K]*list.Element)}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	v, ok := c.items[key]
	if !ok {
		return value, ok
	}
	c.ll.MoveToFront(v)
	return v.Value.(*entry[K, V]).value, true
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	if elem, ok := c.items[key]; ok {
		elem.Value.(*entry[K, V]).value = value
	} else {
		if c.capacity == len(c.items) {
			backElement := c.ll.Back()
			delete(c.items, backElement.Value.(*entry[K, V]).key)
			c.ll.Remove(backElement)
		}
		c.items[key] = c.ll.PushFront(&entry[K, V]{key: key, value: value})
	}
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
