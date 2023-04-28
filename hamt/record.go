package hamt

type record[K comparable, V any] struct {
	key   K
	value V
}

func newRecord[K comparable, V any](k K, v V) *record[K, V] {
	return &record[K, V]{
		key:   k,
		value: v,
	}
}
