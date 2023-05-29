package gocache

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type entry[K comparable, V any] struct {
	key          K
	value        *V
	lastAccessed time.Time
}

func (e *entry[K, V]) hasExpired(ttl time.Duration, t time.Time) bool {
	return !t.Before(e.lastAccessed.Add(ttl))
}

type partitionedCache[K comparable, V any] struct {
	numPartitions    int
	sizePerPartition int
	lru              []*lruQueue[K, V]
	maps             []map[K]*entry[K, V]
	mu               []*sync.Mutex
	ttl              time.Duration
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func newPartitionedCached[K comparable, V any](sizePerPartition int, numPartitions int, ttl time.Duration) *partitionedCache[K, V] {
	if sizePerPartition <= 0 {
		sizePerPartition = 64
	}
	if numPartitions <= 0 {
		numPartitions = 16
	}

	cache := &partitionedCache[K, V]{
		numPartitions:    numPartitions,
		sizePerPartition: sizePerPartition,
		lru:              make([]*lruQueue[K, V], numPartitions),
		maps:             make([]map[K]*entry[K, V], numPartitions),
		mu:               make([]*sync.Mutex, numPartitions),
		ttl:              ttl,
	}

	for i := 0; i < numPartitions; i++ {
		cache.lru[i] = newLRUQueue[K, V]()
		cache.maps[i] = map[K]*entry[K, V]{}
		cache.mu[i] = &sync.Mutex{}
	}

	return cache
}

func (c *partitionedCache[K, V]) partition(key K) uint32 {
	return hash(fmt.Sprintf("%v", key)) % uint32(c.numPartitions)
}

func (c *partitionedCache[K, V]) lock(key K) {
	c.mu[c.partition(key)].Lock()
}

func (c *partitionedCache[K, V]) unlock(key K) {
	c.mu[c.partition(key)].Unlock()
}

func (c *partitionedCache[K, V]) getUnsafe(key K) (*V, bool) {
	partition := c.partition(key)
	entry, ok := c.maps[partition][key]

	if !ok {
		// no entry
		return nil, false
	}

	now := time.Now()
	if !entry.hasExpired(c.ttl, now) {
		// entry has not expired, bump and return
		entry.lastAccessed = now
		c.lru[partition].update(entry)
		return entry.value, true
	}

	// entry has expired, remove it
	delete(c.maps[partition], key)
	c.lru[partition].delete(key)

	return nil, false
}

func (c *partitionedCache[K, V]) putUnsafe(key K, value *V) {
	partition := c.partition(key)
	existing, ok := c.maps[partition][key]
	if ok {
		// already exists, overwrite value
		existing.value = value
		existing.lastAccessed = time.Now()
		c.lru[partition].update(existing)
		return
	}

	// do we need to evict?
	if len(c.maps[partition]) >= c.sizePerPartition {
		oldest := c.lru[partition].pop()

		// remove oldest entry
		delete(c.maps[partition], oldest.key)
	}

	e := &entry[K, V]{
		key:          key,
		value:        value,
		lastAccessed: time.Now(),
	}

	c.maps[partition][key] = e
	c.lru[partition].push(e)
}

func (c *partitionedCache[K, V]) GetOrCreate(key K, value *V) (*V, bool) {
	c.lock(key)
	v, ok := c.getUnsafe(key)
	if ok {
		c.unlock(key)
		return v, true
	}

	c.putUnsafe(key, value)
	c.unlock(key)
	return value, false
}

func (c *partitionedCache[K, V]) Put(key K, value *V) {
	c.lock(key)
	defer c.unlock(key)

	c.putUnsafe(key, value)
}

func (c *partitionedCache[K, V]) HasKey(key K) bool {
	c.lock(key)
	defer c.unlock(key)
	_, ok := c.getUnsafe(key)
	return ok
}

func (c *partitionedCache[K, V]) PartitionLen(key K) int {
	c.lock(key)
	defer c.unlock(key)

	return len(c.maps[c.partition(key)])
}
