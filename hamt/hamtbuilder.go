package hamt

type HAMTBuilber[K comparable, V any] struct {
	m *HAMT[K, V]
}

func NewHAMTBuilder[K comparable, V any](hasher Hasher[K]) *HAMTBuilber[K, V] {

	m := New[K, V](hasher)
	m.mutable = true // make it mutable
	return &HAMTBuilber[K, V]{
		m: m,
	}

}

func (b *HAMTBuilber[K, V]) Len() int {

	return b.m.Len()

}

func (b *HAMTBuilber[K, V]) Get(key K) (V, bool) {

	return b.m.Get(key)

}

func (b *HAMTBuilber[K, V]) Set(key K, value V) {
	keyHash := b.m.hash(key, 0)

	if b.m.root == nil {
		b.m.root = newNode[K, V]()
	}
	b.m.replaceOrInsert(b.m.root, keyHash, 0, newRecord[K, V](key, value))

}

func (b *HAMTBuilber[K, V]) ToHAMT() *HAMT[K, V] {
	m := b.m
	m.mutable = false // make it immutable
	b.m = nil
	return m

}
