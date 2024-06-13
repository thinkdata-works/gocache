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

// go test -v -run TestPartitionedCache$ ./pkg/gocache
func TestPartitionedCache(t *testing.T) {
	sizeOfPartition := 16
	numPartitions := 4
	num := 6

	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, 1*time.Hour)
	_, exists, p := cache.GetOrCreate("A", &num)
	assert.False(t, exists)

	p.Wait()

	exists = cache.HasKey("A")
	assert.True(t, exists)

	val, exists, _ := cache.GetOrCreate("A", &num)
	assert.True(t, exists)
	assert.Equal(t, 6, *val)
}

func TestPartitionedCacheEviction(t *testing.T) {
	sizeOfPartition := 16
	numPartitions := 4

	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, time.Hour)
	for i := 0; i < 100000; i++ {
		key := randString(16)
		_, existsAlready, p := cache.GetOrCreate(key, nil)
		assert.False(t, existsAlready)
		p.Wait()
		assert.LessOrEqual(t, cache.PartitionLen(key), sizeOfPartition)
	}
}

func TestPartitionedCacheTTL(t *testing.T) {
	sizeOfPartition := 16
	numPartitions := 4
	cache := newPartitionedCached[string, int](sizeOfPartition, numPartitions, time.Microsecond)

	_, existsAlready, p := cache.GetOrCreate("A", nil)
	assert.False(t, existsAlready)
	p.Wait()

	time.Sleep(time.Millisecond)

	_, existsAlready, _ = cache.GetOrCreate("A", nil)
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
				_, _, p := cache.GetOrCreate(k, &v)
				p.Wait()
				assert.GreaterOrEqual(t, cache.PartitionLen(k), 1)
				assert.LessOrEqual(t, cache.PartitionLen(k), sizeOfPartition)
				time.Sleep(time.Millisecond)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
