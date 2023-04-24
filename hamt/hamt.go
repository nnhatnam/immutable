package hamt

import (
	"golang.org/x/exp/slices"
	"unsafe"
)

const (
	bSpace         = 64
	arity          = 6
	width          = 2
	exhaustedLevel = bSpace / arity
)

//
//type entry[K comparable, V any] struct {
//	key   K
//	value V
//}
//
//func newEntry[K comparable, V any](k K, v V) entry[K, V] {
//	return entry[K, V]{
//		key:   k,
//		value: v,
//	}
//}
//
//type entries[K comparable, V any] []entry[K, V]
//
//// insertAt inserts a value into the given index, pushing all subsequent values
//// forward.
//func (s *entries[K, V]) insertAt(index int, k K, v V) {
//	item := newEntry[K, V](k, v)
//
//	var zero entry[K, V]
//	*s = append(*s, zero)
//	if index < len(*s) {
//		copy((*s)[index+1:], (*s)[index:])
//	}
//	(*s)[index] = item
//}
//
//func (s *entries[K, V]) append(k K, v V) {
//	item := newEntry[K, V](k, v)
//	*s = append(*s, item)
//}
//
//func getEntry[K comparable, V any](e entries[K, V], key K) entry[K, V] {
//
//	for _, v := range e {
//		if v.key == key {
//			return v
//		}
//	}
//
//	return entry[K, V]{}
//}

type Hasher[K comparable] interface {
	Hash(key K) uint64

	Rehash(key K, level int) uint64
}

type node[K comparable, V any] struct {
	bitmap uint64

	//vPos []int // position of value in values array
	//values []entry[K, V]

	contentArray []unsafe.Pointer // either suffix hash or []entry[K, V], *Value is stored in even slots, []entry[K, V] in odd slots.
}

func newNode[K comparable, V any]() *node[K, V] {
	return &node[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
	}
}

// panic if index is out of range
func (n *node[K, V]) InsertContentAt(index int, k K, v V) bool {
	recordIdx := width * index
	nodeIdx := index + 1

	_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	// n.contentArray = slices.Insert(n.contentArray, index, unsafe.Pointer(&r), unsafe.Pointer((*node[K, V])(nil)))
	n.contentArray = slices.Insert(n.contentArray, index, unsafe.Pointer(newRecord[K, V](k, v)), nil)

	return true

}

// panic if index is out of range
func (n *node[K, V]) SetValueAt(index int, k K, v V) bool {
	recordIdx := width * index

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	r := newRecord[K, V](k, v)
	n.contentArray[recordIdx] = unsafe.Pointer(&r)

	return true

}

// panic if index is out of range
func (n *node[K, V]) SetRecordAt(index int, r *record[K, V]) bool {
	recordIdx := width * index

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	n.contentArray[recordIdx] = unsafe.Pointer(r)

	return true

}

func (n *node[K, V]) SetNodeAt(index int, subNode *node[K, V]) bool {
	nodeIdx := width*index + 1

	_ = n.contentArray[nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	n.contentArray[nodeIdx] = unsafe.Pointer(subNode)

	return true

}

func (n *node[K, V]) GetContentAt(index int) (*record[K, V], *node[K, V]) {

	recordIdx := width * index
	nodeIdx := index + 1

	_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	return (*record[K, V])(n.contentArray[recordIdx]), (*node[K, V])(n.contentArray[nodeIdx])
}

type HAMT[K comparable, V any] struct {
	root *node[K, V]
	len  int

	hasher Hasher[K]

	mutable bool // if true, the HAMT is mutable, otherwise it is immutable.

}

//func New[K comparable, V any]() *HAMT[K, V] {
//	return &HAMT[K, V]{
//		hasher: maphash.NewHasher[K](),
//		//values: make([]V, 1, 32),
//	}
//}

func New[K comparable, V any](h Hasher[K]) *HAMT[K, V] {
	return &HAMT[K, V]{
		hasher: h,
		//values: make([]V, 1, 32),
	}
}

func (t *HAMT[K, V]) Len() int {
	return t.len
}

// subroutin of mInsert
func (t *HAMT[K, V]) mInsertCollision(n *node[K, V], keyHash uint64, r1 *record[K, V], colHash uint64, r2 *record[K, V], shift int) bool {

	//bitpos := t.hasher.Rehash(k1, 0)

}

func (t *HAMT[K, V]) mInsertRecord(n *node[K, V], keyHash uint64, shift int, hashCount int, r *record[K, V]) (_ V, _ bool) {

	bitpos := 1 << mask(keyHash, shift)

	idx := index(n.bitmap, bitpos)

	if n.bitmap&uint64(bitpos) == 0 { // no collision
		n.bitmap |= uint64(bitpos)
		n.InsertContentAt(idx, r.key, r.value)
		return
	}

	colRecord, n1 := n.GetContentAt(idx)

	if colRecord == nil { // collision

		if shift+width >= bSpace { // we have reached the end of the hash, so we have to rehash
			// rehash
			keyHash = t.hasher.Rehash(r.key, hashCount)
			return t.mInsertRecord(n, keyHash, 0, hashCount+1, r)
		}
		return t.mInsertRecord(n1, keyHash, shift+width, hashCount, r)

	}

	if colRecord.key == r.key { // update
		n.SetRecordAt(idx, r)
		return colRecord.value, true
	}

	n2 := newNode[K, V]()
	if hashCount > 0 {

		t.mInsertCollision(n2, keyHash, r, t.hasher.Rehash(colRecord.key, hashCount-1), colRecord, shift)
		return
	}
	t.mInsertCollision(n2, keyHash, r, t.hasher.Hash(colRecord.key), colRecord, shift)
	return
}
