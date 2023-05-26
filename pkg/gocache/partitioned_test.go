package gocache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func randLetter() string {
	return string(letterRunes[rand.Intn(len(letterRunes))])
}

func TestPartitionedCacheEviction(t *testing.T) {
	sizeOfPartition := 16
	numPartitions := 4

	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, time.Hour)
	for i := 0; i < 100000; i++ {
		key := randString(16)
		_, existsAlready := cache.GetOrCreate(key, nil)
		assert.False(t, existsAlready)
		assert.LessOrEqual(t, cache.PartitionLen(key), sizeOfPartition)
	}
}

func TestPartitionedCacheTTL(t *testing.T) {
	sizeOfPartition := 16
	numPartitions := 4
	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, time.Microsecond)

	_, existsAlready := cache.GetOrCreate("A", nil)
	assert.False(t, existsAlready)

	time.Sleep(time.Millisecond)

	_, existsAlready = cache.GetOrCreate("A", nil)
	assert.False(t, existsAlready)
}

func TestPartitionedCacheAsyncTest(t *testing.T) {
	sizeOfPartition := 640
	numPartitions := 4
	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, time.Second)

	numWorkers := 32
	wg := &sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			numActions := 1024
			for j := 0; j < numActions; j++ {
				v := j
				k := randLetter()
				cache.GetOrCreate(k, &v)
				assert.GreaterOrEqual(t, cache.PartitionLen(k), 1)
				assert.LessOrEqual(t, cache.PartitionLen(k), sizeOfPartition)
				time.Sleep(time.Millisecond)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
