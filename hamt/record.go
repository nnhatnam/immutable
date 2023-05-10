package hamt

import "sync/atomic"

type record[K comparable, V any] struct {
	key      K
	value    V
	refCount int32 // atomic
	release  func(key K, value V)
}

func newRecord[K comparable, V any](k K, v V, release func(key K, value V)) *record[K, V] {
	return &record[K, V]{
		key:      k,
		value:    v,
		refCount: 1, // atomic
		release:  release,
	}
}

func (r *record[K, V]) incRef() *record[K, V] {
	if r == nil {
		return nil
	}
	atomic.AddInt32(&r.refCount, 1)
	return r
}

func (r *record[K, V]) decRef() {
	if r != nil {
		if atomic.AddInt32(&r.refCount, -1) == 0 {
			r.release(r.key, r.value)
			r.release = nil
		}
	}
}
