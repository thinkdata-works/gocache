package gocache

import (
	"time"

	"github.com/thinkdata-works/gopromise/pkg/promise"
)

type Cache[K any, V any] struct {
	generateKey func(K) (string, error)
	cache       *partitionedCache[string, promise.Promise[V]]
}

func NewCache[K any, V any](partitionSize, totalPartitions int, cacheExpiry time.Duration, keygen func(K) (string, error)) *Cache[K, V] {
	return &Cache[K, V]{
		cache: newPartitionedCached[string, promise.Promise[V]](
			partitionSize, totalPartitions, cacheExpiry,
		),
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

func (c *Cache[K, V]) Get(k K, getter func(K) (*V, error)) (*V, error) {
	key, err := c.generateKey(k)
	if err != nil {
		return getter(k)
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
		v, err := getter(k)
		if err != nil {
			valpromise.Reject(err)
			return
		}
		valpromise.Resolve(v)
	}()

	return valpromise.Wait()
}
