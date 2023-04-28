package hamt

import (
	"golang.org/x/exp/slices"
	"math/bits"
	"unsafe"
)

const (
	childPerNode   = 64
	arity          = 6
	width          = 2
	exhaustedLevel = childPerNode / arity

	maxDepth = 55 // 64-bit hash, 6-bit per level, 64/6 ~ 11, 11*5 = 55 (accepted 5 rehashes)
)

type Hasher[K comparable] interface {
	Hash(key K) uint64

	Rehash(key K, level int) uint64
}

type node[K comparable, V any] struct {
	bitmap uint64

	contentArray []unsafe.Pointer // either suffix hash or []entry[K, V], *Value is stored in even slots, []entry[K, V] in odd slots.
}

func newNode[K comparable, V any]() *node[K, V] {
	return &node[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
	}
}

func (n *node[K, V]) contentBlockInfo(loc int) (mask uint64, blockIndex int) {
	mask = 1 << loc

	if mask > 0 {
		// find the block index
		blockIndex = bits.OnesCount64(n.bitmap & (mask - 1))
	}

	return
}

// TrySetBlock tries to set the record at the given location. If the location is already occupied, return false
func (n *node[K, V]) TrySetBlock(loc int, r *record[K, V], n1 *node[K, V]) bool {

	mask, blockIdx := n.contentBlockInfo(loc)

	// if the block is not empty, we can't set. return false
	if n.bitmap&mask != 0 {
		return false
	}

	recordIdx := width * blockIdx

	n.contentArray = slices.Insert(n.contentArray, recordIdx, unsafe.Pointer(r), unsafe.Pointer(n1))
	n.bitmap |= mask

	return true

}

// UpdateBlock set the record at the given location. It will overwrite the existing record if there is any
// Panics if there is no existing records at the give location. Only call it when you are sure the location already has a record.
func (n *node[K, V]) UpdateBlock(loc int, r *record[K, V], n1 *node[K, V]) bool {

	_, blockIdx := n.contentBlockInfo(loc)
	recordIdx := width * blockIdx

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	//n.bitmap |= mask
	n.contentArray[recordIdx] = unsafe.Pointer(r)
	n.contentArray[recordIdx+1] = unsafe.Pointer(n1)

	return true

}

func (n *node[K, V]) GetContentAt(loc int) (*record[K, V], *node[K, V]) {

	_, blockIdx := n.contentBlockInfo(loc)
	recordIdx := width * blockIdx
	nodeIdx := recordIdx + 1

	_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	return (*record[K, V])(n.contentArray[recordIdx]), (*node[K, V])(n.contentArray[nodeIdx])
}

type HAMT[K comparable, V any] struct {
	root *node[K, V]
	len  int

	hasher Hasher[K]

	mutable bool // if true, the HAMT is mutable, otherwise it is immutable.

}

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
	// The double record situation only happens when we have a collision at the previous level.
	// When a collision happens, we have to create a new node and insert the two records into the new node.
	// So we know that n is a new node for sure.

	level := depth % (exhaustedLevel + 1)
	shift, hashCount := level*arity, level/arity

	//bitpos := t.hasher.Rehash(k1, 0)
	bitpos := bucket(keyHash, shift)
	colpos := bucket(colHash, shift)

	if bitpos == colpos { // collision again
		// we have to create a new node
		n1 := newNode[K, V]()
		if !n.TrySetBlock(bitpos, nil, n1) {
			panic("impossible case. Need to investigate")
		}
		if level == exhaustedLevel {
			keyHash = t.hash(r1.key, hashCount)
			colHash = t.hash(r2.key, hashCount)

			return t.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
		}

		return t.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
	}

	n.TrySetBlock(bitpos, r1, nil)
	n.TrySetBlock(colpos, r2, nil)

	return true
}

func (t *HAMT[K, V]) mInsertRecord(n *node[K, V], keyHash uint64, depth int, r *record[K, V]) (_ V, _ bool) {

	level := depth % (exhaustedLevel + 1)
	shift, hashCount := level*arity, level/arity
	loc := bucket(keyHash, shift)

	if n.TrySetBlock(loc, r, nil) {
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
			n.UpdateBlock(loc, r, nil)
			return colRecord.value, true
		}

		n2 := newNode[K, V]()
		n.UpdateBlock(loc, nil, n2)

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

func (t *HAMT[K, V]) get(n *node[K, V], k K, keyHash uint64, depth int) (_ V, _ bool) {

	level := depth % (exhaustedLevel + 1)
	shift, hashCount := level*arity, level/arity

	loc := bucket(keyHash, shift)

	colRecord, n1 := n.GetContentAt(loc)

	if colRecord == nil && n1 == nil {
		return
	}

	if colRecord != nil {
		if colRecord.key == k {
			return colRecord.value, true
		}
		return
	}

	if level == exhaustedLevel {
		keyHash = t.hash(k, hashCount)
	}

	return t.get(n1, k, keyHash, depth+1)

}

func (t *HAMT[K, V]) ReplaceOrInsert(k K, v V) (_ V, _ bool) {
	keyHash := t.hash(k, 0)
	if t.root == nil {
		t.root = newNode[K, V]()
	}

	ret, ok := t.mInsertRecord(t.root, keyHash, 0, newRecord[K, V](k, v))
	if !ok {
		t.len++
	}

	return ret, ok

}

func (t *HAMT[K, V]) Get(k K) (_ V, _ bool) {
	if t.root == nil {
		return
	}

	keyHash := t.hash(k, 0)
	return t.get(t.root, k, keyHash, 0)
}
