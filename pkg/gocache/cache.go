package gocache

import (
	"time"

	"github.com/thinkdata-works/gopromise/pkg/promise"
)

type Cache[K any, V any] struct {
	getter      func(K) (*V, error)
	generateKey func(K) (string, error)
	cache       *partitionedCache[string, promise.Promise[V]]
}

func NewCache[K any, V any](partitionSize, totalPartitions int, cacheExpiry time.Duration, getter func(K) (*V, error), keygen func(K) (string, error)) *Cache[K, V] {
	return &Cache[K, V]{
		cache: newPartitionedCached[string, promise.Promise[V]](
			partitionSize, totalPartitions, cacheExpiry,
		),
		getter:      getter,
		generateKey: keygen,
	}
}

func (c *Cache[K, V]) HasKey(k K) (bool, error) {
	key, err := c.generateKey(k)
	if err != nil {
		return false, err
	}
	return c.cache.HasKey(key), nil
}

func (c *Cache[K, V]) Get(k K) (*V, error) {
	key, err := c.generateKey(k)
	if err != nil {
		return c.getter(k)
	}

	valpromise, alreadyExists := c.cache.GetOrCreate(key, promise.NewPromise[V]())
	if alreadyExists {
		val, err := valpromise.Wait()
		if err != nil {
			return nil, err
		}
		return val, nil
	}

	// otherwise, get
	go func() {
		v, err := c.getter(k)
		if err != nil {
			valpromise.Reject(err)
			return
		}
		valpromise.Resolve(v)
	}()

	return valpromise.Wait()
}
