package gocache

import (
	"container/heap"
)

type heapImpl[K comparable, V any] struct {
	entries []*entry[K, V]
	indices map[K]int
}

func (h heapImpl[K, V]) Len() int {
	return len(h.entries)
}

func (h heapImpl[K, V]) Less(i, j int) bool {
	return h.entries[i].lastAccessed.Before(h.entries[j].lastAccessed)
}

func (h heapImpl[K, V]) Swap(i, j int) {
	h.entries[i], h.entries[j] = h.entries[j], h.entries[i]
	h.indices[h.entries[i].key] = i
	h.indices[h.entries[j].key] = j
}

func (h *heapImpl[K, V]) Push(x any) {
	n := len(h.entries)
	item := x.(*entry[K, V])
	h.entries = append(h.entries, item)
	h.indices[item.key] = n
}

func (h *heapImpl[K, V]) Pop() any {
	old := h.entries
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	delete(h.indices, item.key)
	h.entries = old[0 : n-1]
	return item
}

type lruQueue[K comparable, V any] struct {
	impl heapImpl[K, V]
}

func newLRUQueue[K comparable, V any]() *lruQueue[K, V] {
	l := &lruQueue[K, V]{
		impl: heapImpl[K, V]{
			indices: map[K]int{},
		},
	}
	heap.Init(&l.impl)
	return l
}

func (l *lruQueue[K, V]) push(v *entry[K, V]) {
	heap.Push(&l.impl, v)
}

func (l *lruQueue[K, V]) update(item *entry[K, V]) {
	heap.Fix(&l.impl, l.impl.indices[item.key])
}

func (l *lruQueue[K, V]) pop() *entry[K, V] {
	return heap.Pop(&l.impl).(*entry[K, V])
}

func (l *lruQueue[K, V]) delete(key K) {
	heap.Remove(&l.impl, l.impl.indices[key])
}
