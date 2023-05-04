package hamt

import (
	"golang.org/x/exp/slices"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

type mapNode[K comparable, V any] struct {
	bitmap   uint64
	refCount int32

	contentArray []unsafe.Pointer // either suffix hash or []entry[K, V], *Value is stored in even slots, []entry[K, V] in odd slots.
}

func newMapNode[K comparable, V any]() *mapNode[K, V] {
	return &mapNode[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
	}
}

//func newMapNodeWithRef[K comparable, V any]() *mapNode[K, V] {
//	return &mapNode[K, V]{
//		contentArray: make([]unsafe.Pointer, 0),
//		refCount:     1,
//	}
//}

func newMapNodeWithRef[K comparable, V any]() *mapNode[K, V] {
	return &mapNode[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
		refCount:     1,
	}
}

func newMapNode2[K comparable, V any]() *mapNode[K, V] {
	return &mapNode[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
	}
}

func (n *mapNode[K, V]) shallowCloneWithRef() *mapNode[K, V] {
	atomic.AddInt32(&n.refCount, -1)

	n1 := &mapNode[K, V]{
		bitmap:   n.bitmap,
		refCount: 1,
	}

	n1.contentArray = make([]unsafe.Pointer, len(n.contentArray))

	for i := 0; i < len(n.contentArray)/2; i++ {

		recordIdx := i * 2
		nodeIdx := i*2 + 1
		n1.contentArray[recordIdx] = n.contentArray[recordIdx]
		if n1.contentArray[nodeIdx] != nil {
			n1.contentArray[nodeIdx] = n.contentArray[nodeIdx]
			(*mapNode[K, V])(n1.contentArray[nodeIdx]).incRef()
		}

	}
	//copy(n1.contentArray, n.contentArray)

	return n1

}

func (n *mapNode[K, V]) shallowClone() *mapNode[K, V] {
	n1 := &mapNode[K, V]{
		bitmap:   n.bitmap,
		refCount: 1,
	}

	n1.contentArray = make([]unsafe.Pointer, len(n.contentArray))
	copy(n1.contentArray, n.contentArray)

	return n1
}

func (n *mapNode[K, V]) incRef() *mapNode[K, V] {
	if n != nil {
		atomic.AddInt32(&n.refCount, 1)
	}
	return n
}

func (n *mapNode[K, V]) decRef() {
	if n == nil {
		return
	}

	if atomic.AddInt32(&n.refCount, -1) == 0 {
		// free the node
		for i := 0; i < len(n.contentArray)/2; i++ {
			recordIdx := i * 2
			nodeIdx := i*2 + 1

			if n.contentArray[recordIdx] != nil {
				// free the record
				n.contentArray[recordIdx] = nil
			}

			if n.contentArray[nodeIdx] != nil {
				// free the node
				n1 := (*mapNode[K, V])(n.contentArray[nodeIdx])
				n1.decRef()
			}
		}
		n.contentArray = nil
	}
}

func (n *mapNode[K, V]) contentBlockInfo(loc int) (mask uint64, blockIndex int) {
	mask = 1 << loc

	if mask > 0 {
		// find the block index
		blockIndex = bits.OnesCount64(n.bitmap & (mask - 1))
	}

	return
}

// TryInsertBlock tries to set the record at the given location. If the location is already occupied, return false
func (n *mapNode[K, V]) TryInsertBlock(loc int, r *record[K, V], n1 *mapNode[K, V]) bool {

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
//func (n *mapNode[K, V]) UpdateBlock(loc int, r *record[K, V], n1 *mapNode[K, V]) bool {
//
//	_, blockIdx := n.contentBlockInfo(loc)
//	recordIdx := width * blockIdx
//
//	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808
//
//	//n.bitmap |= mask
//	n.contentArray[recordIdx] = unsafe.Pointer(r)
//	n.contentArray[recordIdx+1] = unsafe.Pointer(n1)
//
//	return true
//
//}

func (n *mapNode[K, V]) UpdateBlock(loc int, r *record[K, V], n1 *mapNode[K, V]) bool {

	_, blockIdx := n.contentBlockInfo(loc)
	recordIdx := width * blockIdx

	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808

	//n.bitmap |= mask
	n.contentArray[recordIdx] = unsafe.Pointer(r)
	n.contentArray[recordIdx+1] = unsafe.Pointer(n1)

	return true

}

func (n *mapNode[K, V]) TryGetBlock(loc int) (*record[K, V], *mapNode[K, V]) {

	_, blockIdx := n.contentBlockInfo(loc)
	recordIdx := width * blockIdx
	nodeIdx := recordIdx + 1

	if recordIdx >= len(n.contentArray) {
		return nil, nil
	}

	//_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	return (*record[K, V])(n.contentArray[recordIdx]), (*mapNode[K, V])(n.contentArray[nodeIdx])
}

type PersistentHAMT[K comparable, V any] struct {
	root *mapNode[K, V]
	len  int

	hasher Hasher[K]

	mutable bool // if true, the HAMT is mutable, otherwise it is immutable.
}

func NewPersistentHAMT[K comparable, V any](h Hasher[K]) *PersistentHAMT[K, V] {
	return &PersistentHAMT[K, V]{
		hasher: h,
		root:   newMapNode[K, V](),
	}
}

func (m *PersistentHAMT[K, V]) hash(key K, depth int) uint64 {

	prevHashCount := depth / (exhaustedLevel + 1)
	if prevHashCount == 0 {
		return m.hasher.Hash(key)
	} else if depth > maxDepth {
		panic("hash depth too deep, which mean there are too many collisions in the chosen hash function. Please choose a better hash function.")
	}

	return m.hasher.Rehash(key, prevHashCount)
}

func (m *PersistentHAMT[K, V]) Len() int {
	return m.len
}

// subroutine of mInsert
func (m *PersistentHAMT[K, V]) mInsertDoubleRecord(n *mapNode[K, V], keyHash uint64, r1 *record[K, V], colHash uint64, r2 *record[K, V], depth int) bool {
	// The double record situation only happens when we have a collision at the previous level.
	// When a collision happens, we have to create a new node and insert the two records into the new node.
	// So we know that n is a new node for sure.

	level := depth % (exhaustedLevel + 1)
	shift := level * arity

	bitpos := bucket(keyHash, shift)
	colpos := bucket(colHash, shift)

	if bitpos == colpos { // collision again
		// we have to create a new node
		n1 := newMapNodeWithRef[K, V]()
		if !n.TryInsertBlock(bitpos, nil, n1) {
			panic("impossible case. Need to investigate")
		}
		if level == exhaustedLevel {
			keyHash = m.hash(r1.key, depth)
			colHash = m.hash(r2.key, depth)

			return m.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
		}

		return m.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
	}

	n.TryInsertBlock(bitpos, r1, nil)
	n.TryInsertBlock(colpos, r2, nil)

	return true
}

func (m *PersistentHAMT[K, V]) replaceOrInsert(root *mapNode[K, V], keyHash uint64, depth int, r *record[K, V]) (n *mapNode[K, V], oldValue V, replaced bool) {

	if root.refCount > 1 {
		n = root.shallowCloneWithRef()

	} else {
		n = root
	}

	level := depth % (exhaustedLevel + 1)
	shift := level * arity
	loc := bucket(keyHash, shift)

	// if the block is empty, we can insert the record directly
	if n.TryInsertBlock(loc, r, nil) {
		return
	}

	colRecord, n1 := n.TryGetBlock(loc)

	switch {
	case colRecord == nil: // collision
		// we have to create a new node
		if level == exhaustedLevel {
			keyHash = m.hash(r.key, depth)
		}
		var next *mapNode[K, V]
		next, oldValue, replaced = m.replaceOrInsert(n1, keyHash, depth+1, r)
		if next == n1 {
			n1.decRef()
		}
		n.UpdateBlock(loc, nil, next)
		return
	case n1 == nil: // n1 == nil && colRecord != nil => update or expand

		if colRecord.key == r.key { // update
			n.UpdateBlock(loc, r, nil)
			return n, colRecord.value, true
		}

		n2 := newMapNodeWithRef[K, V]()
		n.UpdateBlock(loc, nil, n2)

		colHash := m.hash(colRecord.key, depth)
		if level == exhaustedLevel {
			keyHash = m.hash(r.key, depth)
			m.mInsertDoubleRecord(n2, keyHash, r, colHash, colRecord, depth+1)
			return
		}

		m.mInsertDoubleRecord(n2, keyHash, r, colHash, colRecord, depth+1)

	}
	return
}

func (m *PersistentHAMT[K, V]) get(n *mapNode[K, V], k K, keyHash uint64, depth int) (_ V, _ bool) {

	level := depth % (exhaustedLevel + 1)
	shift := level * arity

	loc := bucket(keyHash, shift)

	colRecord, n1 := n.TryGetBlock(loc)

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
		keyHash = m.hash(k, depth)
	}

	return m.get(n1, k, keyHash, depth+1)

}

func (m *PersistentHAMT[K, V]) Set(k K, v V) *PersistentHAMT[K, V] {
	keyHash := m.hash(k, 0)

	trie := NewPersistentHAMT[K, V](m.hasher)
	//var oldValue V
	var replaced bool

	if m.root != nil {

		trie.len = m.len
		trie.root, _, replaced = trie.replaceOrInsert(m.root, keyHash, 0, newRecord[K, V](k, v))

	} else {
		trie.root, _, replaced = trie.replaceOrInsert(trie.root, keyHash, 0, newRecord[K, V](k, v))
	}

	if !replaced {
		trie.len++
	}

	return trie

}

func (m *PersistentHAMT[K, V]) Get(k K) (_ V, _ bool) {
	if m.root == nil {
		return
	}

	keyHash := m.hash(k, 0)
	return m.get(m.root, k, keyHash, 0)
}

func (m *PersistentHAMT[K, V]) Clone() *PersistentHAMT[K, V] {
	return &PersistentHAMT[K, V]{
		root:   m.root.incRef(),
		len:    m.len,
		hasher: m.hasher,
	}
}
