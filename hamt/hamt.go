package hamt

import (
	"golang.org/x/exp/slices"
	"unsafe"
)

const (
	childPerNode   = 64
	arity          = 6
	width          = 2
	exhaustedLevel = childPerNode / arity

	maxDepth = 55 // 64-bit hash, 6-bit per level, 64/6 ~ 11, 11*5 = 55 (accepted 5 rehashes)
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
//func (n *node[K, V]) InsertContentAt(index int, k K, v V) bool {
//	recordIdx := width * index
//	nodeIdx := index + 1
//
//	_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808
//
//	// n.contentArray = slices.Insert(n.contentArray, index, unsafe.Pointer(&r), unsafe.Pointer((*node[K, V])(nil)))
//	n.contentArray = slices.Insert(n.contentArray, index, unsafe.Pointer(newRecord[K, V](k, v)), nil)
//
//	return true
//}

func (n *node[K, V]) InsertContentAt(bitpos int, k K, v V) bool {

	if n.bitmap&(1<<uint(bitpos)) != 0 {
		return false
	}

	recordIdx := width * index(n.bitmap, bitpos)
	nodeIdx := recordIdx + 1

	_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	// n.contentArray = slices.Insert(n.contentArray, index, unsafe.Pointer(&r), unsafe.Pointer((*node[K, V])(nil)))
	n.contentArray = slices.Insert(n.contentArray, recordIdx, unsafe.Pointer(newRecord[K, V](k, v)), nil)
	n.bitmap |= 1 << uint(bitpos)
	return true
}

// panic if index is out of range
func (n *node[K, V]) SetValueAt(bitpos int, k K, v V) bool {
	recordIdx := width * index(n.bitmap, bitpos)

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	r := newRecord[K, V](k, v)

	n.bitmap |= 1 << uint(bitpos)
	n.contentArray[recordIdx] = unsafe.Pointer(&r)

	return true

}

// panic if index is out of range
func (n *node[K, V]) SetRecordAt(bitpos int, r *record[K, V]) bool {

	recordIdx := width * index(n.bitmap, bitpos)

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	n.bitmap |= 1 << uint(bitpos)
	n.contentArray[recordIdx] = unsafe.Pointer(r)

	return true

}

func (n *node[K, V]) SetNodeAt(bitpos int, subNode *node[K, V]) bool {
	nodeIdx := width*index(n.bitmap, bitpos) + 1

	_ = n.contentArray[nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	n.bitmap |= 1 << uint(bitpos)
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

func (t *HAMT[K, V]) hash(key K, level int) uint64 {
	if level == 0 {
		return t.hasher.Hash(key)
	}
	return t.hasher.Rehash(key, level)
}

func (t *HAMT[K, V]) Len() int {
	return t.len
}

// subroutin of mInsert
func (t *HAMT[K, V]) mInsertDoubleRecord(n *node[K, V], keyHash uint64, r1 *record[K, V], colHash uint64, r2 *record[K, V], depth int) bool {

	level := depth % (exhaustedLevel + 1)
	shift, hashCount := level*arity, level/arity

	//bitpos := t.hasher.Rehash(k1, 0)
	bitpos := 1 << mask(keyHash, shift)
	colpos := 1 << mask(colHash, shift)

	if bitpos == colpos { // collision again
		// we have to create a new node

		n1 := newNode[K, V]()
		n.SetNodeAt(bitpos, n1)

		if level == exhaustedLevel {
			keyHash = t.hash(r1.key, hashCount)
			colHash = t.hash(r2.key, hashCount)

			return t.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
		}

		return t.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
	}

	n.InsertContentAt(bitpos, r1.key, r1.value)
	n.InsertContentAt(colpos, r2.key, r2.value)
	return true
}

func (t *HAMT[K, V]) mInsertRecord(n *node[K, V], keyHash uint64, depth int, r *record[K, V]) (_ V, _ bool) {

	level := depth % (exhaustedLevel + 1)
	shift, hashCount := level*arity, level/arity
	loc := 1 << mask(keyHash, shift)

	if n.InsertContentAt(loc, r.key, r.value) {
		return
	}

	colRecord, n1 := n.GetContentAt(loc)

	switch {
	case colRecord == nil: // collision
		// we have to create a new node
		if level == exhaustedLevel {
			keyHash = t.hash(r.key, hashCount)
			return t.mInsertRecord(n1, keyHash, depth+1, r)
		}
		return t.mInsertRecord(n1, keyHash, depth+1, r)
	case n1 == nil: // n1 == nil && colRecord != nil => update or expand
		if colRecord.key == r.key { // update
			n.SetRecordAt(loc, r)
			return colRecord.value, true
		}

		n2 := newNode[K, V]()
		n.SetNodeAt(loc, n2)
		n.SetRecordAt(loc, nil)

		colHash := t.hash(colRecord.key, hashCount)
		if level == exhaustedLevel {
			keyHash = t.hash(r.key, hashCount)
			t.mInsertDoubleRecord(n2, keyHash, r, colHash, colRecord, depth+1)
			return
		}

		t.mInsertDoubleRecord(n2, keyHash, r, colHash, colRecord, depth+1)

	}
	return
}
