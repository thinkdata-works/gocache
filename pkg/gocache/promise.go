package gocache

import (
	"sync"
	"time"
)

type future[V any] struct {
	v   *V
	err error
}

type Promise[V any] struct {
	ch    chan future[V]
	v     *V
	err   error
	mu    *sync.Mutex
	count int
	done  bool
	start time.Time
}

func NewPromise[V any]() *Promise[V] {
	return &Promise[V]{
		ch:    make(chan future[V]),
		start: time.Now(),
		mu:    &sync.Mutex{},
		done:  false,
		count: 0,
		v:     nil,
		err:   nil,
	}
}

func (p *Promise[V]) Wait() (*V, error) {
	p.mu.Lock()
	if p.done {
		p.mu.Unlock()
		return p.v, p.err
	}
	p.count++
	p.mu.Unlock()

	f := <-p.ch
	return f.v, f.err
}

func (p *Promise[V]) fulfill(v *V, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.done {
		// already fulfilled
		return
	}
	p.done = true
	p.v = v
	p.err = err
	for i := 0; i < p.count; i++ {
		p.ch <- future[V]{
			v:   v,
			err: err,
		}
	}
}

func (p *Promise[V]) Resolve(v *V) {
	p.fulfill(v, nil)
}

func (p *Promise[V]) Reject(err error) {
	p.fulfill(nil, err)
}
