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

//
//type records[K comparable, V any] []record[K, V]
//
//// insertAt inserts a value into the given index, pushing all subsequent values
//// forward.
//func (s *records[K, V]) insertAt(index int, k K, v V) {
//	item := newRecord[K, V](k, v)
//
//	var zero record[K, V]
//	*s = append(*s, zero)
//	if index < len(*s) {
//		copy((*s)[index+1:], (*s)[index:])
//	}
//	(*s)[index] = item
//}
//
//func (s *records[K, V]) append(k K, v V) {
//	item := newRecord[K, V](k, v)
//	*s = append(*s, item)
//}
