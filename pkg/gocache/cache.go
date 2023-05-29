package gocache

import (
	"time"

	"github.com/thinkdata-works/gopromise/pkg/promise"
)

type Cache[K comparable, V any] struct {
	cache *partitionedCache[K, promise.Promise[V]]
}

func NewCache[K comparable, V any](partitionSize, totalPartitions int, cacheExpiry time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		cache: newPartitionedCached[K, promise.Promise[V]](
			partitionSize, totalPartitions, cacheExpiry,
		),
	}
}

func (c *Cache[K, V]) HasKey(k K) (bool, error) {
	return c.cache.HasKey(k), nil
}

func (c *Cache[K, V]) Get(k K, getter func() (*V, error)) (*V, error) {
	valpromise, alreadyExists := c.cache.GetOrCreate(k, promise.NewPromise[V]())
	if alreadyExists {
		val, err := valpromise.Wait()
		if err != nil {
			return nil, err
		}
		return val, nil
	}

	// otherwise, get
	go func() {
		v, err := getter()
		if err != nil {
			valpromise.Reject(err)
			return
		}
		valpromise.Resolve(v)
	}()

	return valpromise.Wait()
}
