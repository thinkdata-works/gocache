package gocache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUQueue(t *testing.T) {
	queue := newLRUQueue[string, int]()

	now := time.Now()

	queue.push(&entry[string, int]{
		key:          "a",
		lastAccessed: now.Add(time.Second * 10),
	})

	queue.push(&entry[string, int]{
		key:          "b",
		lastAccessed: now,
	})

	queue.push(&entry[string, int]{
		key:          "c",
		lastAccessed: now.Add(time.Second * -10),
	})

	// pop the oldest
	e := queue.pop()
	assert.Equal(t, e.key, "c")

	queue.push(&entry[string, int]{
		key:          "d",
		lastAccessed: now.Add(time.Second * -5),
	})

	// pop the oldest
	e = queue.pop()
	assert.Equal(t, e.key, "d")

	// pop the oldest
	e = queue.pop()
	assert.Equal(t, e.key, "b")

	queue.push(&entry[string, int]{
		key:          "e",
		lastAccessed: now.Add(time.Second * 9),
	})

	queue.push(&entry[string, int]{
		key:          "f",
		lastAccessed: now,
	})

	queue.push(&entry[string, int]{
		key:          "g",
		lastAccessed: now.Add(time.Second * -10),
	})

	// pop the oldest
	e = queue.pop()
	assert.Equal(t, e.key, "g")

	// pop the oldest
	queue.delete("f")

	e = queue.pop()
	assert.Equal(t, e.key, "e")

	e = queue.pop()
	assert.Equal(t, e.key, "a")
}
